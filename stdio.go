package termite

import (
	"bufio"
	"io"
	"os"
)

var (
	// StdoutWriter to be used as standard out
	StdoutWriter io.Writer

	// StderrWriter to be used as standard err
	StderrWriter io.Writer

	// StdinReader to be used as standard in
	StdinReader io.Reader
)


func init() {
	StdoutWriter = NewAutoFlushingWriter(os.Stdout)
	StderrWriter = NewAutoFlushingWriter(os.Stderr)
	StdinReader = os.Stdin
}


type autoFlushingWriter struct {
	writer *bufio.Writer
}

// NewAutoFlushingWriter creates a new io.Writer that uses a buffer internally and flushes after every write.
// This writer should be used on top of Stdout and Stderr for components that require frequent screen updates.
func NewAutoFlushingWriter(w io.Writer) io.Writer {
	return &autoFlushingWriter{
		writer: bufio.NewWriter(w),
	}
}

func (sw *autoFlushingWriter) Write(b []byte) (int, error) {
	sw.writer.Flush()
	return sw.writer.Write(b)
}