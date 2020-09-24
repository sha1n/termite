package termite

import (
	"container/ring"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var defaultSpinnerCharacters = []string{
	"\u259B", "\u2599", "\u259F", "\u259C",
}

// Spinner a spinning progress indicator
type Spinner interface {
	Start() (context.CancelFunc, error)
	Stop(string) error
}

type spinner struct {
	terminal Terminal
	cursor   Cursor
	interval time.Duration
	mx       *sync.RWMutex
	active   bool
	stopC    chan bool
}

// NewSpinner creates a new Spinner with the specified update interval
func NewSpinner(t Terminal, interval int32) Spinner {
	return &spinner{
		terminal: t,
		interval: time.Duration(interval),
		mx:       &sync.RWMutex{},
		active:   false,
		stopC:    make(chan bool),
	}
}

// NewDefaultSpinner creates a new Spinner with a default update interval
func NewDefaultSpinner(t Terminal) Spinner {
	return NewSpinner(t, 500)
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
		timer := time.NewTicker(time.Millisecond * s.interval)

		waitStart.Done()

		defer func() {
			s.mx.Lock()
			defer s.mx.Unlock()

			s.active = false
		}()

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
				s.terminal.OverwriteLine(fmt.Sprintf("%s", spinring.Value))
			}
		}
	}()

	waitStart.Wait()

	return cancel, err
}

func (s *spinner) isActive() bool {
	s.mx.Lock()
	defer s.mx.Unlock()

	return s.active
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

func (s *spinner) printExitMessage(message string) {
	s.terminal.EraseLine()
	s.terminal.Println(message)
}

func createSpinnerRing() *ring.Ring {
	r := ring.New(4)

	r.Value = defaultSpinnerCharacters[0]
	r = r.Next()
	r.Value = defaultSpinnerCharacters[1]
	r = r.Next()
	r.Value = defaultSpinnerCharacters[2]
	r = r.Next()
	r.Value = defaultSpinnerCharacters[3]
	r = r.Next()

	return r
}
