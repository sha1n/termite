package termite

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	tsize "github.com/kopoli/go-terminal-size"
	"github.com/mattn/go-isatty"
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
		var size tsize.Size
		if size, err = tsize.GetSize(); err == nil {
			width = size.Width
			height = size.Height
		}
	} else {
		err = errors.New("no tty. Terminal dimensions cannot be resolved")
	}

	return
}
