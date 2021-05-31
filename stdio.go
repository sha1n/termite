package termite

import (
	"bufio"
	"io"
	"os"
)

var (
	// StdoutWriter to be used as standard out
	StdoutWriter *AutoFlushingWriter

	// StderrWriter to be used as standard err
	StderrWriter *AutoFlushingWriter

	// StdinReader to be used as standard in
	StdinReader io.Reader
)

func init() {
	StdoutWriter = NewAutoFlushingWriter(os.Stdout)
	StderrWriter = NewAutoFlushingWriter(os.Stderr)
	StdinReader = os.Stdin
}

// AutoFlushingWriter an implementation of an io.Writer and io.StringWriter with auto-flush semantics.
type AutoFlushingWriter struct {
	Writer *bufio.Writer
}

// NewAutoFlushingWriter creates a new io.Writer that uses a buffer internally and flushes after every write.
// This writer should be used on top of Stdout and Stderr for components that require frequent screen updates.
func NewAutoFlushingWriter(w io.Writer) *AutoFlushingWriter {
	return &AutoFlushingWriter{
		Writer: bufio.NewWriter(w),
	}
}

func (sw *AutoFlushingWriter) Write(b []byte) (int, error) {
	defer sw.Writer.Flush()
	return sw.Writer.Write(b)
}

// WriteString uses io.WriteString to write the specified string to the underlying writer.
func (sw *AutoFlushingWriter) WriteString(s string) (int, error) {
	return sw.Write([]byte(s))
}
