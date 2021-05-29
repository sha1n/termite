package termite

import (
	"bufio"
	"bytes"
)

// FakeTerminal a fake Terminal implementation for testing purposes.
type FakeTerminal struct {
	Terminal
	Out    *bytes.Buffer
	Err    *bytes.Buffer
	width  int
	height int
}

// NewFakeTerminal ...
func NewFakeTerminal(width, height int) *FakeTerminal {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	t := NewTerminal(true)
	(t.(*term)).Out = bufio.NewWriter(outBuf)
	(t.(*term)).Err = bufio.NewWriter(errBuf)

	return &FakeTerminal{
		Out:      outBuf,
		Err:      errBuf,
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
