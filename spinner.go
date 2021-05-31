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

var defaultSpinnerCharSeq = []string{
	"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
}

// Spinner a spinning progress indicator
type Spinner interface {
	Start() (context.CancelFunc, error)
	Stop(string) error
	SetTitle(title string)
}

type spinner struct {
	writer   io.Writer
	interval time.Duration
	mx       *sync.RWMutex
	titleMx  *sync.RWMutex
	active   bool
	stopC    chan bool
	title    string
}

// NewSpinner creates a new Spinner with the specified update interval
func NewSpinner(writer io.Writer, title string, interval time.Duration) Spinner {
	return &spinner{
		writer:   writer,
		interval: interval,
		mx:       &sync.RWMutex{},
		titleMx:  &sync.RWMutex{},
		active:   false,
		stopC:    make(chan bool),
		title:    title,
	}
}

// NewDefaultSpinner creates a new Spinner that writes to Stdout with a default update interval
func NewDefaultSpinner() Spinner {
	return NewSpinner(StdoutWriter, "", 500)
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
		var spinring = createSpinnerRing()
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
				if title != "" {
					s.writeString(fmt.Sprintf("%s%s %s", TermControlEraseLine, spinring.Value, title))
				} else {
					s.writeString(fmt.Sprintf("%s%s", TermControlEraseLine, spinring.Value))
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

func createSpinnerRing() *ring.Ring {
	r := ring.New(len(defaultSpinnerCharSeq))

	for _, ch := range defaultSpinnerCharSeq {
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
