package termite

import (
	"fmt"
	"io"
)

// Cursor represents a terminal cursor
type Cursor interface {
	Position(row, col int)
	Up(l int)
	Down(l int)
	Forward(cols int)
	Backward(cols int)
	SavePosition()
	RestorePosition()
	Hide()
	Show()
}

type cursor struct {
	writer io.Writer
}

// NewCursor returns a new cursor for the specified terminal
func NewCursor(writer io.Writer) Cursor {
	return cursor{
		writer: writer,
	}
}

func (c cursor) Position(row, col int) {
	_, _ = c.writeString(fmt.Sprintf(termControlCursorPositionFmt, row, col))
}

func (c cursor) Up(lines int) {
	_, _ = c.writeString(fmt.Sprintf(termControlCursorUpFmt, lines))
}

func (c cursor) Down(lines int) {
	_, _ = c.writeString(fmt.Sprintf(termControlCursorDownFmt, lines))
}

func (c cursor) Forward(cols int) {
	_, _ = c.writeString(fmt.Sprintf(termControlCursorForwardFmt, cols))
}

func (c cursor) Backward(cols int) {
	_, _ = c.writeString(fmt.Sprintf(termControlCursorBackwardFmt, cols))
}

func (c cursor) Hide() {
	_, _ = c.writeString(termControlCursorHide)
}

func (c cursor) Show() {
	_, _ = c.writeString(termControlCursorShow)
}

func (c cursor) SavePosition() {
	_, _ = c.writeString(termControlCursorSave)
}

func (c cursor) RestorePosition() {
	_, _ = c.writeString(termControlCursorRestore)
}

func (c cursor) writeString(s string) (int, error) {
	return io.WriteString(c.writer, s)
}
