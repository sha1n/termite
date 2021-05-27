package termite

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// Matrix is a multiline structure that reflects its state on screen
type Matrix interface {
	StringWriter() io.StringWriter
	RefreshInterval() time.Duration
	NewLineStringWriter() io.StringWriter
	NewLineWriter() io.Writer
	Start() context.CancelFunc
}

// MatrixLine an accessor to a line in a Matrix structure
type MatrixLine interface {
	WriteString(s string)
}

type terminalMatrix struct {
	lines           []string
	refreshInterval time.Duration
	writer          io.StringWriter
	mx              *sync.RWMutex
}

type matrixLineWriter struct {
	index  int
	matrix *terminalMatrix
}

// NewMatrix creates a new Matrix for the specified Terminal
func NewMatrix(writer io.StringWriter) Matrix {
	return &terminalMatrix{
		lines:           []string{},
		refreshInterval: time.Millisecond * 100,
		writer:          writer,
		mx:              &sync.RWMutex{},
	}
}

func (m *terminalMatrix) StringWriter() io.StringWriter {
	return m.writer
}

func (m *terminalMatrix) RefreshInterval() time.Duration {
	return m.refreshInterval
}

// Start starts the matrix update process.
// Returns a cancel handle to stop the matrix updates.
func (m *terminalMatrix) Start() context.CancelFunc {
	context, cancel := context.WithCancel(context.Background())

	waitStart := &sync.WaitGroup{}
	waitStart.Add(1)
	var drainWaitGroup *sync.WaitGroup

	go func() {
		timer := time.NewTicker(m.refreshInterval)
		drainWaitGroup = &sync.WaitGroup{}
		drainWaitGroup.Add(1)
		// now that we loaded the drain wait group, we can release the caller
		waitStart.Done()

		for {
			select {
			case <-context.Done():
				timer.Stop()
				m.updateTerminal(false)
				drainWaitGroup.Done()
				return

			case <-timer.C:
				m.updateTerminal(true)
			}
		}
	}()

	waitStart.Wait()

	return func() {
		cancel()
		// Wait for the final update to complete
		drainWaitGroup.Wait()
	}
}

func (m *terminalMatrix) updateTerminal(resetCursorPosition bool) {
	c := NewCursor(m.writer)
	m.mx.Lock()
	defer m.mx.Unlock()

	if len(m.lines) == 0 {
		return
	}

	for _, line := range m.lines {
		m.writer.WriteString(fmt.Sprintf("%s%s\r\n", TermControlEraseLine, line))
	}

	if resetCursorPosition {
		c.Up(len(m.lines))
	}
}

// NewLineStringWriter returns a new string writter to interact with a single matrix line
func (m *terminalMatrix) NewLineStringWriter() io.StringWriter {
	m.mx.Lock()
	defer m.mx.Unlock()

	index := len(m.lines)
	m.lines = append(m.lines, "")
	return &matrixLineWriter{
		index:  index,
		matrix: m,
	}
}

// NewLineWriter returns a new writer interface to interact with a single matrix line.
func (m *terminalMatrix) NewLineWriter() io.Writer {
	return m.NewLineStringWriter().(*matrixLineWriter)
}

func (l *matrixLineWriter) WriteString(s string) (n int, err error) {
	return l.Write([]byte(s))
}

func (l *matrixLineWriter) Write(b []byte) (n int, err error) {
	l.matrix.mx.Lock()
	defer l.matrix.mx.Unlock()

	l.matrix.lines[l.index] = string(b)
	return len(b), nil
}
