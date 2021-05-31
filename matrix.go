package termite

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Matrix is a multiline structure that reflects its state on screen
type Matrix interface {
	// Start starts to update this matrix in the background
	Start() context.CancelFunc

	// NewRow allocates and returns a MatrixRow
	NewRow() MatrixRow

	// NewRange allocates and returns the specified n umber of rows
	NewRange(int) []MatrixRow

	// NewLineStringWriter returns a new string writter to interact with a single matrix line
	//
	// Deprecated: use NewRow instead
	NewLineStringWriter() io.StringWriter // FIXME refactor

	// NewLineWriter returns a new writer interface to interact with a single matrix line
	//
	// Deprecated: use NewRow instead
	NewLineWriter() io.Writer // FIXME refactor

	// RefreshInterval returns the refresh interval of this matrix
	RefreshInterval() time.Duration

	// GetRow looks up a row by index. Returns an error if none exists
	GetRow(int) (MatrixRow, error)

	// GetRowByID looks up a row an ID. Returns an error if none exists
	GetRowByID(MatrixCellID) (MatrixRow, error)
}

// MatrixCellID used to identify a Matrix cell internally
type MatrixCellID struct {
	row int
}

// MatrixRow an accessor to a line in a Matrix structure
// Line feed and return characters are trimmed from written strings to prevent breaking the layout of the matrix.
type MatrixRow interface {
	io.StringWriter
	io.Writer
	ID() MatrixCellID
	Update(string)
}

type matrixImpl struct {
	lines           []string
	refreshInterval time.Duration
	writer          io.Writer
	mx              *sync.RWMutex
}

type matrixRow struct {
	id     MatrixCellID
	matrix *matrixImpl
}

// NewMatrix creates a new matrix that writes to the specified writer and refreshes every refreshInterval.
func NewMatrix(writer io.Writer, refreshInterval time.Duration) Matrix {
	return &matrixImpl{
		lines:           []string{},
		refreshInterval: refreshInterval,
		writer:          writer,
		mx:              &sync.RWMutex{},
	}
}

func (m *matrixImpl) RefreshInterval() time.Duration {
	return m.refreshInterval
}

// Start starts the matrix update process.
// Returns a cancel handle to stop the matrix updates.
func (m *matrixImpl) Start() context.CancelFunc {
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

func (m *matrixImpl) GetRow(index int) (row MatrixRow, err error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if index < 0 {
		return nil, errors.New("row index cannot be negative")
	}
	if index >= len(m.lines) {
		return nil, errors.New("row index exceeds the matrix range")
	}

	row = &matrixRow{
		id:     MatrixCellID{row: index},
		matrix: m,
	}

	return row, err
}

func (m *matrixImpl) GetRowByID(id MatrixCellID) (row MatrixRow, err error) {
	return m.GetRow(id.row)
}

func (m *matrixImpl) updateTerminal(resetCursorPosition bool) {
	c := NewCursor(m.writer)
	m.mx.Lock()
	defer m.mx.Unlock()

	if len(m.lines) == 0 {
		return
	}

	for _, line := range m.lines {
		io.WriteString(m.writer, fmt.Sprintf("%s%s\r\n", TermControlEraseLine, line))
	}

	if resetCursorPosition {
		c.Up(len(m.lines))
	}
}

func (m *matrixImpl) NewLineStringWriter() io.StringWriter {
	return m.NewRow()
}

func (m *matrixImpl) NewLineWriter() io.Writer {
	return m.NewRow()
}

func (m *matrixImpl) NewRange(count int) []MatrixRow {
	m.mx.Lock()
	defer m.mx.Unlock()

	var rows []MatrixRow
	for i := 0; i < count; i++ {
		rows = append(rows, m.newRow())
	}

	return rows
}

func (m *matrixImpl) NewRow() MatrixRow {
	m.mx.Lock()
	defer m.mx.Unlock()

	return m.newRow()
}

func (m *matrixImpl) newRow() MatrixRow {
	index := len(m.lines)
	m.lines = append(m.lines, "")
	return &matrixRow{
		id:     MatrixCellID{row: index},
		matrix: m,
	}
}

func (r *matrixRow) WriteString(s string) (n int, err error) {
	return r.Write([]byte(s))
}

func (r *matrixRow) Write(b []byte) (n int, err error) {
	r.matrix.mx.Lock()
	defer r.matrix.mx.Unlock()

	// we trim line feeds and return characters to prevent breaking the matrix layout
	r.matrix.lines[r.id.row] = strings.Trim(string(b), "\n\r")
	return len(b), nil
}

func (r *matrixRow) Update(s string) {
	_, _ = r.Write([]byte(s))
}

func (r *matrixRow) ID() MatrixCellID {
	return r.id
}
