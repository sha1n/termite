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

	// DefaultProgressBarBlank default progress bar fill character
	DefaultProgressBarBlank = '\u2591'

	percentAreaSpace = 7
)

// DefaultProgressBarFormatter returns a new instance of the default ProgressBarFormatter
func DefaultProgressBarFormatter() *SimpleProgressBarFormatter {
	return &SimpleProgressBarFormatter{
		LeftBorderChar:  DefaultProgressBarLeftBorder,
		RightBorderChar: DefaultProgressBarRightBorder,
		FillChar:        DefaultProgressBarFill,
		BlankChar:       DefaultProgressBarBlank,
	}
}

// DefaultProgressBarFormatterWidth returns a default formatter with custom message area width.
func DefaultProgressBarFormatterWidth(width int) *SimpleProgressBarFormatter {
	return &SimpleProgressBarFormatter{
		LeftBorderChar:  DefaultProgressBarLeftBorder,
		RightBorderChar: DefaultProgressBarRightBorder,
		FillChar:        DefaultProgressBarFill,
		BlankChar:       DefaultProgressBarBlank,
		MessageWidth:    width,
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

	// FormatBlank returns a string that contains one visible character and optionally
	// additional styling charatcers such as color codes, background and other effects.
	FormatBlank() string

	// MessageAreaWidth return the number of character used for the message area.
	MessageAreaWidth() int
}

// SimpleProgressBarFormatter a simple ProgressBarFormatter implementation which is based on constructor values.
type SimpleProgressBarFormatter struct {
	LeftBorderChar  rune
	RightBorderChar rune
	FillChar        rune
	BlankChar       rune
	MessageWidth    int
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

// FormatBlank returns the blank char
func (f *SimpleProgressBarFormatter) FormatBlank() string {
	return fmt.Sprintf("%c", f.BlankChar)
}

// MessageAreaWidth returns zero
func (f *SimpleProgressBarFormatter) MessageAreaWidth() int {
	return f.MessageWidth
}

// TickMessageFn a tick handle
type TickMessageFn = func(string) bool

// ProgressBar a progress bar interface
type ProgressBar interface {
	Tick() bool
	TickMessage(message string) bool
	IsDone() bool
	Start() (TickMessageFn, context.CancelFunc, error)
}

type bar struct {
	maxTicks           int
	ticks              int
	writer             io.Writer
	width              int
	formatter          ProgressBarFormatter
	renderStringFormat string
	active             bool
	mx                 *sync.RWMutex
}

type progressEvent struct {
	ok  bool
	msg string
}

// NewProgressBar creates a new progress bar
// writer 				- the writer to use for output
// maxTicks 			- how many ticks are to be considered 100% of the progress
// terminalWidth 	- the width of the terminal
// width 					- bar width in characters
// formatter 		  - a formatter for this progress bar
func NewProgressBar(writer io.Writer, maxTicks int, terminalWidth int, width int, formatter ProgressBarFormatter) ProgressBar {
	renderFormat := fmt.Sprintf("%%s%%%ds %%s%%s%%s%%s %%d%%%%", formatter.MessageAreaWidth())
	return &bar{
		maxTicks:           maxTicks,
		ticks:              0,
		writer:             writer,
		width:              max(0, min(width, terminalWidth-percentAreaSpace-formatter.MessageAreaWidth())),
		formatter:          formatter,
		renderStringFormat: renderFormat,
		mx:                 &sync.RWMutex{},
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
	return b.TickMessage("")
}

// TickTickMessage increments the progress by one tick. Does not imply visual change.
func (b *bar) TickMessage(message string) bool {
	if b.IsDone() {
		return false
	}

	b.ticks++

	return b.render(message)
}

// Start starts the progress bar in the background and returns a tick handle, a cancellation handle and an error in case
// this bar has already been started.
func (b *bar) Start() (tick TickMessageFn, cancel context.CancelFunc, err error) {
	defer b.mx.Unlock()
	b.mx.Lock()

	if b.active {
		return nil, nil, errors.New("Progress bar already running in the background")
	}
	b.active = true
	b.render("")

	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())

	events := make(chan progressEvent)
	var done bool
	waitStart := &sync.WaitGroup{}
	waitStart.Add(1)

	tick = func(msg string) bool {
		if ctx.Err() != nil {
			return false
		}

		if !done {
			events <- progressEvent{ok: true, msg: msg}
			maybeDoneEvent := <-events
			done = !maybeDoneEvent.ok
		}
		return !done
	}

	go func() {
		waitStart.Done()
		for {
			select {
			case <-ctx.Done():
				return

			case evt := <-events:
				evt.ok = b.TickMessage(evt.msg)
				events <- evt
			}
		}
	}()

	waitStart.Wait()

	return tick, cancel, err
}

func (b *bar) render(message string) bool {
	totalChars := b.width
	percent := float32(b.ticks) / float32(b.maxTicks)
	charsToFill := int(percent * float32(totalChars))
	spaceChars := totalChars - charsToFill

	_, _ = io.WriteString(
		b.writer,
		fmt.Sprintf(
			b.renderStringFormat,
			TermControlEraseLine,
			TruncateString(message, b.formatter.MessageAreaWidth()),
			b.formatter.FormatLeftBorder(),
			strings.Repeat(b.formatter.FormatFill(), charsToFill),
			strings.Repeat(b.formatter.FormatBlank(), spaceChars),
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
