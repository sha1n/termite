package termite

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/mattn/go-isatty"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	StdoutWriter = os.Stdout
	StderrWriter = os.Stderr
	StdinReader = os.Stdin

	Tty = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	if Tty {
		var err error
		terminalWidth, terminalHeight, err = terminal.GetSize(int(os.Stdin.Fd()))

		if err != nil {
			println("failed to resolve retminal dimensions")
		}
	}
}

var (
	terminalWidth  int
	terminalHeight int
	
	// StdoutWriter to be used as standard out
	StdoutWriter io.Writer

	// StderrWriter to be used as standard err
	StderrWriter io.Writer

	// StdinReader to be used as standard in
	StdinReader io.Reader

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
	WriteString(s string) (int, error)
	AllocateNewLines(int)

	Width() int
	Height() int

	Print(e interface{})
	Println(e interface{})
}

type term struct {
	Out       *bufio.Writer
	Err       *bufio.Writer
	autoFlush bool
	outLock   *sync.Mutex
}

// NewTerminal creates a instance of Terminal
func NewTerminal(autoFlush bool) Terminal {
	return &term{
		Out:       bufio.NewWriter(StdoutWriter),
		Err:       bufio.NewWriter(StderrWriter),
		autoFlush: autoFlush,
		outLock:   &sync.Mutex{},
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
	t.WriteString(fmt.Sprintf("%v", e))
}

func (t *term) Println(e interface{}) {
	t.WriteString(fmt.Sprintf("%v%s", e, TermControlCRLF))
}

func (t *term) WriteString(s string) (n int, err error) {
	t.outLock.Lock()
	defer t.outLock.Unlock()

	if t.autoFlush {
		defer t.Out.Flush()
	}

	return t.Out.WriteString(s)
}

func (t *term) AllocateNewLines(count int) {
	t.Print(strings.Repeat("\n", count)) // allocate 4 lines
	NewCursor(t).Up(count)               // return to start position
}
