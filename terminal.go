package termite

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	Tty = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	if Tty {
		var err error
		terminalWidth, terminalHeight, err = terminal.GetSize(int(os.Stdin.Fd()))

		if err != nil {
			println("failed to resolve terminal dimensions")
		}
	}
}

var (
	terminalWidth  int
	terminalHeight int

	// Tty whether or not we have a terminal
	Tty bool
)

const (
	// TermControlEraseLine clears the current line and positions the cursor at the beginning
	TermControlEraseLine = "\r\033[K"

	// TermControlClearScreen emulates the bash/sh clear command
	TermControlClearScreen = "\033[H\033[2J"

	// TermControlCRLF line feed
	TermControlCRLF = "\r\n"
)

// Terminal privides terminal related APIs
type Terminal interface {
	StdOut() io.Writer
	StdErr() io.Writer
	AllocateNewLines(int)

	Width() int
	Height() int

	Print(e interface{})
	Println(e interface{})
}

type term struct {
	Out io.Writer
	Err io.Writer
}

// NewTerminal creates a instance of Terminal
func NewTerminal() Terminal {
	return &term{
		Out: StdoutWriter,
		Err: StderrWriter,
	}
}

// Width returns the current terminal width.
// If no TTY returns 0
// If it fails to get the width it panics with an error.
func (t *term) Width() (width int) {
	return terminalWidth
}

// Height returns the current terminal height.
// If no TTY returns 0
// If it fails to get the width it panics with an error.
func (t *term) Height() (height int) {
	return terminalHeight
}

func (t *term) StdOut() io.Writer {
	return t.Out
}

func (t *term) StdErr() io.Writer {
	return t.Err
}

func (t *term) Print(e interface{}) {
	io.WriteString(t.Out, fmt.Sprintf("%v", e))
}

func (t *term) Println(e interface{}) {
	io.WriteString(t.Out, fmt.Sprintf("%v%s", e, TermControlCRLF))
}

func (t *term) AllocateNewLines(count int) {
	io.WriteString(t.Out, strings.Repeat("\n", count)) // allocate 4 lines
	NewCursor(t.Out).Up(count)                         // return to start position
}
