package termite

import (
	"bytes"
)

// FakeTerminal a fake Terminal implementation for testing purposes.
type FakeTerminal struct {
	Terminal
	width  int
	height int
}

// NewFakeTerminal ...
func NewFakeTerminal(width, height int) *FakeTerminal {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	t := NewTerminal()
	(t.(*term)).Out = NewAutoFlushingWriter(outBuf)
	(t.(*term)).Err = NewAutoFlushingWriter(errBuf)

	return &FakeTerminal{
		width:    width,
		height:   height,
		Terminal: t,
	}
}

// Width return a pre-set fake width
func (t *FakeTerminal) Width() (width int) {
	return t.width
}

// Height return a pre-set fake height
func (t *FakeTerminal) Height() (height int) {
	return t.height
}
