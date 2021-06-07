package termite

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

const (
	// DefaultProgressBarLeftBorder default progress bar left border character
	DefaultProgressBarLeftBorder = '\u258F'

	// DefaultProgressBarRightBorder default progress bar right border character
	DefaultProgressBarRightBorder = '\u2595'

	// DefaultProgressBarFill default progress bar fill character
	DefaultProgressBarFill = '\u2587'
)

// DefaultProgressBarFormatter returns a new instance of the default ProgressBarFormatter
func DefaultProgressBarFormatter() *SimpleProgressBarFormatter {
	return &SimpleProgressBarFormatter{
		LeftBorderChar:  DefaultProgressBarLeftBorder,
		RightBorderChar: DefaultProgressBarRightBorder,
		FillChar:        DefaultProgressBarFill,
	}
}

// ProgressBarFormatter a formatter to control the style of a ProgressBar.
type ProgressBarFormatter interface {
	// FormatLeftBorder returns a string that contains one visible character and optionally
	// additional styling charatcers such as color codes, background and other effects.
	FormatLeftBorder() string

	// FormatRightBorder returns a string that contains one visible character and optionally
	// additional styling charatcers such as color codes, background and other effects.
	FormatRightBorder() string

	// FormatFill returns a string that contains one visible character and optionally
	// additional styling charatcers such as color codes, background and other effects.
	FormatFill() string
}

// SimpleProgressBarFormatter a simple ProgressBarFormatter implementation which is based on constructor values.
type SimpleProgressBarFormatter struct {
	LeftBorderChar  rune
	RightBorderChar rune
	FillChar        rune
}

// FormatLeftBorder returns the left border char
func (f *SimpleProgressBarFormatter) FormatLeftBorder() string {
	return fmt.Sprintf("%c", f.LeftBorderChar)
}

// FormatRightBorder returns the right border char
func (f *SimpleProgressBarFormatter) FormatRightBorder() string {
	return fmt.Sprintf("%c", f.RightBorderChar)
}

// FormatFill returns the fill char
func (f *SimpleProgressBarFormatter) FormatFill() string {
	return fmt.Sprintf("%c", f.FillChar)
}

// TickFn a tick handle
type TickFn = func() bool

// ProgressBar a progress bar interface
type ProgressBar interface {
	Tick() bool
	IsDone() bool
	Start() (TickFn, context.CancelFunc, error)
}

type bar struct {
	maxTicks  int
	ticks     int
	writer    io.Writer
	width     int
	formatter ProgressBarFormatter
	active    bool
	mx        *sync.RWMutex
}

// NewProgressBar creates a new progress bar
// writer 				- the writer to use for output
// maxTicks 			- how many ticks are to be considered 100% of the progress
// terminalWidth 	- the width of the terminal
// width 					- bar width in characters
// formatter 		  - a formatter for this progress bar
func NewProgressBar(writer io.Writer, maxTicks int, terminalWidth int, width int, formatter ProgressBarFormatter) ProgressBar {
	return &bar{
		maxTicks:  maxTicks,
		ticks:     0,
		writer:    writer,
		width:     max(0, min(width, terminalWidth-7)), // 7 = 2 borders, 3 digits, % sign + 1 padding char
		formatter: formatter,
		mx:        &sync.RWMutex{},
	}
}

// NewDefaultProgressBar creates a progress bar with styling
func NewDefaultProgressBar(writer io.Writer, terminalWidth int, maxTicks int) ProgressBar {
	return NewProgressBar(
		writer, maxTicks, terminalWidth/2, terminalWidth, DefaultProgressBarFormatter(),
	)
}

// IsDone returns whether or not this progress bar has reached 100%
func (b *bar) IsDone() bool {
	return b.ticks >= b.maxTicks
}

// Tick increments the progress by one tick. Does not imply visual change.
func (b *bar) Tick() bool {
	if b.IsDone() {
		return false
	}

	b.ticks++

	return b.render()
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
	b.render()

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
				return

			case <-events:
				events <- b.Tick()
			}
		}
	}()

	waitStart.Wait()

	return tick, cancel, err
}

func (b *bar) render() bool {
	totalChars := b.width
	percent := float32(b.ticks) / float32(b.maxTicks)
	charsToFill := int(percent * float32(totalChars))
	spaceChars := totalChars - charsToFill

	_, _ = io.WriteString(
		b.writer,
		fmt.Sprintf(
			"%s%s%s%s%s %d%%\r",
			TermControlEraseLine,
			b.formatter.FormatLeftBorder(),
			strings.Repeat(b.formatter.FormatFill(), charsToFill),
			strings.Repeat(" ", spaceChars),
			b.formatter.FormatRightBorder(),
			int(percent*100),
		),
	)

	return b.maxTicks > b.ticks
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
