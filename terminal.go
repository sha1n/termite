package termite

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/crypto/ssh/terminal"
)

// TermControlEraseLine clears the current line and positions the cursor at the beginning
const TermControlEraseLine = "\r\033[K"

// TermControlClearScreen emulates the bash/sh clear command
const TermControlClearScreen = "\033[H\033[2J"

// TermControlCRLF line feed
const TermControlCRLF = "\r\n"

// Terminal privides terminal related APIs
type Terminal interface {
	StdOut() io.Writer
	StdErr() io.Writer

	Width() int
	Height() int

	Print(e interface{})
	Println(e interface{})
	OverwriteLine(e interface{})
	EraseLine()
	Clear()
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

func (t *term) Width() (width int) {
	if !Tty {
		return 0
	}

	if width, _, err := terminal.GetSize(int(os.Stdin.Fd())); err == nil {
		return width
	}

	// FIXME: we probably need to check whether we have a terminal and handle that earlier.
	panic(errors.New("can't get terminal width"))
}

func (t *term) Height() (height int) {
	if !Tty {
		return 0
	}

	if _, height, err := terminal.GetSize(int(os.Stdin.Fd())); err == nil {
		return height
	}

	// FIXME: we probably need to check whether we have a terminal and handle that earlier.
	panic(errors.New("can't get terminal height"))
}

func (t *term) StdOut() io.Writer {
	return t.Out
}

func (t *term) StdErr() io.Writer {
	return t.Err
}

func (t *term) Print(e interface{}) {
	t.writeString(fmt.Sprintf("%v", e))
}

func (t *term) Println(e interface{}) {
	t.writeString(fmt.Sprintf("%v%s", e, TermControlCRLF))
}

func (t *term) EraseLine() {
	t.writeString(TermControlEraseLine)
}

func (t *term) OverwriteLine(e interface{}) {
	t.Print(fmt.Sprintf("%s%v", TermControlEraseLine, e))
}

func (t *term) Clear() {
	t.writeString(TermControlClearScreen)
}

func (t *term) writeString(s string) (n int) {
	t.outLock.Lock()
	defer t.outLock.Unlock()

	if t.autoFlush {
		defer t.Out.Flush()
	}

	if n, err := t.Out.WriteString(s); err == nil {
		return n
	}
	return 0
}
