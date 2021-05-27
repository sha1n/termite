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
	writer io.StringWriter
}

// NewCursor returns a new cursor for the specified terminal
func NewCursor(writer io.StringWriter) Cursor {
	return &cursor{
		writer: writer,
	}
}

func (c *cursor) Position(row, col int) {
	c.writer.WriteString(fmt.Sprintf("\033[%d;%dH", row, col))
}

func (c *cursor) Up(lines int) {
	c.writer.WriteString(fmt.Sprintf("\033[%dA", lines))
}

func (c *cursor) Down(lines int) {
	c.writer.WriteString(fmt.Sprintf("\033[%dB", lines))
}

func (c *cursor) Forward(cols int) {
	c.writer.WriteString(fmt.Sprintf("\033[%dC", cols))
}

func (c *cursor) Backward(cols int) {
	c.writer.WriteString(fmt.Sprintf("\033[%dD", cols))
}

func (c *cursor) Hide() {
	c.writer.WriteString("\033[?25l")
}

func (c *cursor) Show() {
	c.writer.WriteString("\033[?25h")
}

func (c *cursor) SavePosition() {
	c.writer.WriteString("\033[s")
}

func (c *cursor) RestorePosition() {
	c.writer.WriteString("\033[u")
}
