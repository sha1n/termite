package termite

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Matrix is a multiline structure that reflects its state on screen
type Matrix interface {
	Terminal() Terminal
	RefreshInterval() time.Duration
	NewLine() MatrixLine
	Start() context.CancelFunc
}

// MatrixLine an accessor to a line in a Matrix structure
type MatrixLine interface {
	WriteString(s string)
}

type terminalMatrix struct {
	lines           []string
	refreshInterval time.Duration
	terminal        Terminal
	mx              *sync.RWMutex
}

type terminalLine struct {
	index  int
	matrix *terminalMatrix
}

// NewMatrix creates a new Matrix for the specified Terminal
func NewMatrix(t Terminal) Matrix {
	return &terminalMatrix{
		lines:           []string{},
		refreshInterval: time.Millisecond * 100,
		terminal:        t,
		mx:              &sync.RWMutex{},
	}
}

func (m *terminalMatrix) Terminal() Terminal {
	return m.terminal
}

func (m *terminalMatrix) RefreshInterval() time.Duration {
	return m.refreshInterval
}

// Start starts the matrix update process.
// Returns a cancel handle to stop the matrix updates.
func (m *terminalMatrix) Start() context.CancelFunc {
	c := NewCursor(m.terminal)
	context, cancel := context.WithCancel(context.Background())

	go func() {
		timer := time.NewTicker(m.refreshInterval)
		for {
			select {
			case <-context.Done():
				timer.Stop()

			case <-timer.C:
				if len(m.lines) == 0 {
					continue
				}

				m.mx.Lock()
				for _, line := range m.lines {
					m.terminal.OverwriteLine(fmt.Sprintf("%s\r\n", line))
				}
				c.Up(len(m.lines))
				m.mx.Unlock()
			}
		}
	}()

	return cancel
}

// NewRow creates a new matrix row
func (m *terminalMatrix) NewLine() MatrixLine {
	m.mx.Lock()
	defer m.mx.Unlock()

	index := len(m.lines)
	m.lines = append(m.lines, "")
	return &terminalLine{
		index:  index,
		matrix: m,
	}
}

func (l *terminalLine) WriteString(s string) {
	l.matrix.mx.Lock()
	defer l.matrix.mx.Unlock()

	l.matrix.lines[l.index] = s
}
