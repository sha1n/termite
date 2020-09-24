package termite

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// TickFn a tick handle
type TickFn = func() bool

// ProgressBar a progress bar interface
type ProgressBar interface {
	Tick() bool
	IsDone() bool
	Start() (TickFn, context.CancelFunc, error)
}

type bar struct {
	max         int
	ticks       int
	term        Terminal
	cursor      Cursor
	width       int
	leftBorder  string
	rightBorder string
	fill        string
	active      bool
	mx          *sync.RWMutex
}

// NewProgressBar creates a new progress bar
// t - the terminal to use for io interactions
// maxTicks - how many ticks are to be considered 100% of the progress
// width - bar width in characters
// leftBorder - left border character
// rightBorder - right border character
// fill - fill character
func NewProgressBar(t Terminal, maxTicks int, width int, leftBorder rune, rightBorder rune, fill rune) ProgressBar {
	return &bar{
		max:         maxTicks,
		ticks:       0,
		term:        t,
		cursor:      NewCursor(t),
		width:       min(width, t.Width()-7), // 7 = 2 borders, 3 digits, % sign + 1 padding char
		leftBorder:  string(leftBorder),
		rightBorder: string(rightBorder),
		fill:        string(fill),
		mx:          &sync.RWMutex{},
	}
}

// NewDefaultProgressBar creates a progress bar with default settings
func NewDefaultProgressBar(t Terminal, maxTicks int) ProgressBar {
	return NewProgressBar(
		t, maxTicks, t.Width()/2, '\u258F', '\u2595', '\u2587',
	)
}

// IsDone returns whether or not this progress bar has reached 100%
func (b *bar) IsDone() bool {
	return b.ticks >= b.max
}

// Tick increments the progress by one tick. Does not imply visual change.
func (b *bar) Tick() bool {
	if b.IsDone() {
		return false
	}

	// return to overwrite the previous progress bar only if it exists
	if b.ticks > 0 {
		b.cursor.Up(1)
	}
	b.ticks++

	totalChars := b.width
	percent := float32(b.ticks) / float32(b.max)
	charsToFill := int(percent * float32(totalChars))
	spaceChars := totalChars - charsToFill

	b.term.OverwriteLine(
		fmt.Sprintf(
			"%s%s%s%s %d%%%s",
			b.leftBorder, strings.Repeat(b.fill, charsToFill),
			strings.Repeat(" ", spaceChars),
			b.rightBorder,
			int(percent*100),
			// this ensures neater end of line when the program ends
			// with a progress bar and no line feed is entered.
			TermControlCRLF,
		),
	)

	return spaceChars > 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Start starts the progress bar in the background and returns a tick handle, a cancellation handle and an error in case
// this bar has already been started.
func (b *bar) Start() (tick TickFn, cancel context.CancelFunc, err error) {
	defer b.mx.Unlock()
	b.mx.Lock()

	if b.active {
		return nil, nil, errors.New("Progress bar already running in the background")
	}
	b.active = true

	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())

	events := make(chan bool)
	var done bool
	waitStart := &sync.WaitGroup{}
	waitStart.Add(1)

	tick = func() bool {
		if ctx.Err() != nil {
			return false
		}

		if !done {
			events <- true
			done = !<-events
		}
		return !done
	}

	go func() {
		waitStart.Done()
		for {
			select {
			case <-ctx.Done():
				for b.Tick() {
				}
				return

			case <-events:
				events <- b.Tick()
			}
		}
	}()

	waitStart.Wait()

	return tick, cancel, err
}
