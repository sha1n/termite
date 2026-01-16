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
	c.writeString(fmt.Sprintf(termControlCursorPositionFmt, row, col))
}

func (c cursor) Up(lines int) {
	c.writeString(fmt.Sprintf(termControlCursorUpFmt, lines))
}

func (c cursor) Down(lines int) {
	c.writeString(fmt.Sprintf(termControlCursorDownFmt, lines))
}

func (c cursor) Forward(cols int) {
	c.writeString(fmt.Sprintf(termControlCursorForwardFmt, cols))
}

func (c cursor) Backward(cols int) {
	c.writeString(fmt.Sprintf(termControlCursorBackwardFmt, cols))
}

func (c cursor) Hide() {
	c.writeString(termControlCursorHide)
}

func (c cursor) Show() {
	c.writeString(termControlCursorShow)
}

func (c cursor) SavePosition() {
	c.writeString(termControlCursorSave)
}

func (c cursor) RestorePosition() {
	c.writeString(termControlCursorRestore)
}

func (c cursor) writeString(s string) {
	io.WriteString(c.writer, s)
}
