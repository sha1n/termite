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
	// Start starts to update this matrix in the background.
	// Returns a done channel that closes when the goroutine exits.
	// If the context is already cancelled, returns nil.
	Start(context.Context) <-chan struct{}

	// NewRow allocates and returns a MatrixRow
	NewRow() MatrixRow

	// NewRange allocates and returns the specified n umber of rows
	NewRange(int) []MatrixRow

	// RefreshInterval returns the refresh interval of this matrix
	RefreshInterval() time.Duration

	// GetRow looks up a row by index. Returns an error if none exists
	GetRow(int) (MatrixRow, error)

	// GetRowByID looks up a row an ID. Returns an error if none exists
	GetRowByID(MatrixCellID) (MatrixRow, error)

	// UpdateTerminal updates the terminal immediately.
	//
	// This function can be used as a manual alternative to Start(), which updates the terminal
	// in the background based on the interval specified in the constructor. Combining Start with
	// manual updates can yield unwanted results though.
	UpdateTerminal(resetCursorPosition bool)
}

// MatrixCellID used to identify a Matrix cell internally
type MatrixCellID struct {
	row int
}

// Row returns the row index associated with this ID
func (id MatrixCellID) Row() int {
	return id.row
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
	rows            []*matrixRow
	refreshInterval time.Duration
	writer          io.Writer
	mx              *sync.RWMutex
}

type matrixRow struct {
	id       MatrixCellID
	matrix   *matrixImpl
	value    string
	modified bool
}

// NewMatrix creates a new matrix that writes to the specified writer and refreshes every refreshInterval.
func NewMatrix(writer io.Writer, refreshInterval time.Duration) Matrix {
	return &matrixImpl{
		rows:            []*matrixRow{},
		refreshInterval: refreshInterval,
		writer:          writer,
		mx:              &sync.RWMutex{},
	}
}

func (m *matrixImpl) RefreshInterval() time.Duration {
	return m.refreshInterval
}

// Start starts the matrix update process.
// Returns a done channel that closes when the goroutine exits.
// If the context is already cancelled, returns nil.
func (m *matrixImpl) Start(ctx context.Context) <-chan struct{} {
	if ctx.Err() != nil {
		return nil
	}

	done := make(chan struct{})
	waitStart := &sync.WaitGroup{}
	waitStart.Add(1)

	go func() {
		timer := time.NewTicker(m.refreshInterval)
		// now that we set up, we can release the caller
		waitStart.Done()

		defer func() {
			timer.Stop()
			m.UpdateTerminal(false)
			close(done)
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case <-timer.C:
				m.UpdateTerminal(true)
			}
		}
	}()

	waitStart.Wait()

	return done
}

func (m *matrixImpl) GetRow(index int) (row MatrixRow, err error) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if index < 0 {
		return nil, errors.New("row index cannot be negative")
	}
	if index >= len(m.rows) {
		return nil, errors.New("row index exceeds the matrix range")
	}

	row = m.rows[index]

	return row, err
}

func (m *matrixImpl) GetRowByID(id MatrixCellID) (row MatrixRow, err error) {
	return m.GetRow(id.row)
}

func (m *matrixImpl) UpdateTerminal(resetCursorPosition bool) {
	c := NewCursor(m.writer)
	m.mx.Lock()
	defer m.mx.Unlock()

	if len(m.rows) == 0 {
		return
	}

	for _, row := range m.rows {
		if row.modified {
			_, err := io.WriteString(m.writer, fmt.Sprintf("%s%s\n", TermControlEraseLine, row.value))
			row.modified = err != nil
		} else {
			_, _ = io.WriteString(m.writer, "\n")
		}
	}

	if resetCursorPosition {
		c.Up(len(m.rows))
	}
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
	index := len(m.rows)
	row := &matrixRow{
		id:     MatrixCellID{row: index},
		matrix: m,
	}
	m.rows = append(m.rows, row)

	return row
}

func (r *matrixRow) WriteString(s string) (n int, err error) {
	return r.Write([]byte(s))
}

func (r *matrixRow) Write(b []byte) (n int, err error) {
	r.matrix.mx.Lock()
	defer r.matrix.mx.Unlock()

	row := r.matrix.rows[r.id.row]
	newValue := strings.Trim(string(b), "\n\r")
	row.modified = newValue != row.value
	row.value = newValue

	return len(b), nil
}

func (r *matrixRow) Update(s string) {
	_, _ = r.Write([]byte(s))
}

func (r *matrixRow) ID() MatrixCellID {
	return r.id
}
