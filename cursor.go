package termite

import (
	"fmt"
)

// Cursor represents a terminal cursor
type Cursor interface {
	Up(l int)
	Down(l int)
	Hide()
	Show()
}

type cursor struct {
	t Terminal
}

// NewCursor returns a new cursor for the specified terminal
func NewCursor(t Terminal) Cursor {
	return &cursor{
		t: t,
	}
}

func (c *cursor) Up(lines int) {
	c.t.Print(fmt.Sprintf("\033[%dA", lines))
}

func (c *cursor) Down(lines int) {
	c.t.Print(fmt.Sprintf("\033[%dB", lines))
}

func (c *cursor) Hide() {
	c.t.Print("\033[?25l")
}

func (c *cursor) Show() {
	c.t.Print("\033[?25h")
}
