package termite

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	Tty = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

var (
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

// // Terminal privides terminal related APIs
// type Terminal interface {
// 	StdOut() io.Writer
// 	StdErr() io.Writer
// 	AllocateNewLines(int)

// 	Width() int
// 	Height() int

// 	Print(e interface{})
// 	Println(e interface{})
// }

// type term struct {
// 	Out io.Writer
// 	Err io.Writer
// }

// // NewTerminal creates a instance of Terminal
// func NewTerminal() Terminal {
// 	return &term{
// 		Out: StdoutWriter,
// 		Err: StderrWriter,
// 	}
// }

// // Width returns the current terminal width.
// // If no TTY returns 0
// // If it fails to get the width it panics with an error.
// func (t *term) Width() (width int) {
// 	width, _ = queryTermDimensions()
// 	return width
// }

// // Height returns the current terminal height.
// // If no TTY returns 0
// // If it fails to get the width it panics with an error.
// func (t *term) Height() (height int) {
// 	_, height = queryTermDimensions()
// 	return height
// }

// func (t *term) StdOut() io.Writer {
// 	return t.Out
// }

// func (t *term) StdErr() io.Writer {
// 	return t.Err
// }

// func (t *term) Print(e interface{}) {
// 	io.WriteString(t.Out, fmt.Sprintf("%v", e))
// }

// func (t *term) Println(e interface{}) {
// 	io.WriteString(t.Out, fmt.Sprintf("%v%s", e, TermControlCRLF))
// }

// func (t *term) AllocateNewLines(count int) {
// 	io.WriteString(t.Out, strings.Repeat("\n", count)) // allocate 4 lines
// 	NewCursor(t.Out).Up(count)                         // return to start position
// }

// func queryTermDimensions() (width int, height int) {
// 	if Tty {
// 		var err error
// 		width, height, err = terminal.GetSize(int(os.Stdin.Fd()))

// 		if err != nil {
// 			println("failed to resolve terminal dimensions")
// 		}
// 	}

// 	return width, height
// }

// Print utility function for printing an object into StdoutWritter
func Print(e interface{}) {
	io.WriteString(StdoutWriter, fmt.Sprintf("%v", e))
}

// Println utility function for printing a new line into StdoutWritter
func Println(e interface{}) {
	io.WriteString(StdoutWriter, fmt.Sprintf("%v%s", e, TermControlCRLF))
}

// AllocateNewLines utility function for starting a number of new empty lines on StdoutWriter
func AllocateNewLines(count int) {
	io.WriteString(StdoutWriter, strings.Repeat("\n", count))
	NewCursor(StdoutWriter).Up(count) // return to start position
}

// GetTerminalDimensions attempts to get the current terminal dimensions
// If no TTY returns 0, 0 and an error
func GetTerminalDimensions() (width int, height int, err error) {
	if Tty {
		width, height, err = terminal.GetSize(int(os.Stdin.Fd()))
	} else {
		err = errors.New("no tty. Terminal dimensions cannot be resolved")
	}

	return width, height, err
}
