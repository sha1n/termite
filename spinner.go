package termite

import (
	"container/ring"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// DefaultSpinnerCharSeq returns the default character sequence of a spinner.
func DefaultSpinnerCharSeq() []string {
	return []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
}

// DefaultSpinnerFormatter returns a default
func DefaultSpinnerFormatter() SpinnerFormatter {
	return &SimpleSpinnerFormatter{}
}

// SpinnerFormatter a formatter to be used with a Spinner to customize its style.
type SpinnerFormatter interface {
	// FormatTitle returns the input string with optional styling codes or anything else.
	FormatTitle(s string) string

	// FormatIndicator returns a string that contains one visible character - the one passed as input -
	// and optionally additional styling charatcers such as color codes, background and other effects.
	FormatIndicator(char string) string

	// CharSeq the character sequence to use as indicators.
	CharSeq() []string
}

// SimpleSpinnerFormatter a simple spinner formatter implementation that uses the default
// spinner character sequence and passes the title and the indicator setrings unchanged.
type SimpleSpinnerFormatter struct{}

// FormatTitle returns the input title as is
func (f *SimpleSpinnerFormatter) FormatTitle(s string) string {
	return s
}

// FormatIndicator returns the input char as is
func (f *SimpleSpinnerFormatter) FormatIndicator(char string) string {
	return char
}

// CharSeq returns the default character sequence.
func (f *SimpleSpinnerFormatter) CharSeq() []string {
	return DefaultSpinnerCharSeq()
}

// Spinner a spinning progress indicator
type Spinner interface {
	Start() (context.CancelFunc, error)
	Stop(string) error
	SetTitle(title string)
}

type spinner struct {
	writer    io.Writer
	interval  time.Duration
	mx        *sync.RWMutex
	titleMx   *sync.RWMutex
	active    bool
	stopC     chan bool
	title     string
	formatter SpinnerFormatter
}

// NewSpinner creates a new Spinner with the specified update interval
func NewSpinner(writer io.Writer, title string, interval time.Duration, formatter SpinnerFormatter) Spinner {
	return &spinner{
		writer:    writer,
		interval:  interval,
		mx:        &sync.RWMutex{},
		titleMx:   &sync.RWMutex{},
		active:    false,
		stopC:     make(chan bool),
		title:     title,
		formatter: formatter,
	}
}

// NewDefaultSpinner creates a new Spinner that writes to Stdout with a default update interval
func NewDefaultSpinner() Spinner {
	return NewSpinner(StdoutWriter, "", time.Millisecond*100, DefaultSpinnerFormatter())
}

func (s *spinner) writeString(str string) (n int, err error) {
	return io.WriteString(s.writer, str)
}

// Start starts the spinner in the background and returns a cancellation handle and an error in case the spinner is already running.
func (s *spinner) Start() (cancel context.CancelFunc, err error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	if s.active {
		return nil, errors.New("spinner already active")
	}

	s.active = true
	context, cancel := context.WithCancel(context.Background())
	waitStart := &sync.WaitGroup{}
	waitStart.Add(1)

	go func() {
		var spinring = s.createSpinnerRing()
		timer := time.NewTicker(s.interval)

		waitStart.Done()

		defer s.setActiveSafe(false)

		for {
			select {
			case <-context.Done():
				timer.Stop()
				s.printExitMessage("Cancelled...")
				return

			case <-s.stopC:
				timer.Stop()
				return

			case <-timer.C:
				spinring = spinring.Next()
				title := s.getTitle()
				indicatorValue := s.formatter.FormatIndicator(fmt.Sprintf("%v", spinring.Value))
				if title != "" {
					s.writeString(fmt.Sprintf("%s%s %s", TermControlEraseLine, indicatorValue, s.formatter.FormatTitle(title)))
				} else {
					s.writeString(fmt.Sprintf("%s%s", TermControlEraseLine, indicatorValue))
				}

			}
		}
	}()

	waitStart.Wait()

	return cancel, err
}

// Stop stops the spinner and displays the specified message
func (s *spinner) Stop(message string) (err error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	if !s.active {
		err = errors.New("spinner not active")
	} else {
		s.stopC <- true
		s.active = false
		s.printExitMessage(message)
	}

	return err
}

// SetTitle updates the spinner text.
func (s *spinner) SetTitle(title string) {
	s.titleMx.Lock()
	defer s.titleMx.Unlock()

	s.title = title
}

func (s *spinner) getTitle() string {
	s.titleMx.RLock()
	defer s.titleMx.RUnlock()

	return s.title
}

func (s *spinner) printExitMessage(message string) {
	s.writeString(TermControlEraseLine)
	s.writeString(message)
}

func (s *spinner) createSpinnerRing() *ring.Ring {
	r := ring.New(len(s.formatter.CharSeq()))

	for _, ch := range s.formatter.CharSeq() {
		r.Value = ch
		r = r.Next()
	}

	return r
}

func (s *spinner) isActiveSafe() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.active
}

func (s *spinner) setActiveSafe(active bool) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.active = active
}
