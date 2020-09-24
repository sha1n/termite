package termite

import (
	"github.com/mattn/go-isatty"
	"io"
	"os"
)

// StdoutWriter to be used as standard out
var StdoutWriter io.Writer

// StderrWriter to be used as standard err
var StderrWriter io.Writer

// StdinReader to be used as standard in
var StdinReader io.Reader

// Tty whether or not we have a terminal
var Tty bool

func init() {
	StdoutWriter = os.Stdout
	StderrWriter = os.Stderr
	StdinReader = os.Stdin

	Tty = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}
