package termite

import (
	"bufio"
	"bytes"
	// "io"
)

type fakeTerm struct {
	Terminal
	Out    *bytes.Buffer
	Err    *bytes.Buffer
	width  int
	height int
}

// NewFakeTerminal ...
func NewFakeTerminal(width, height int) *fakeTerm {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	t := NewTerminal(true)
	(t.(*term)).Out = bufio.NewWriter(outBuf)
	(t.(*term)).Err = bufio.NewWriter(errBuf)

	return &fakeTerm{
		Out:      outBuf,
		Err:      errBuf,
		width:    width,
		height:   height,
		Terminal: t,
	}
}

func (t *fakeTerm) Width() (width int) {
	return t.width
}

func (t *fakeTerm) Height() (height int) {
	return t.height
}
