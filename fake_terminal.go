package termite

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

type fakeTerm struct {
	Out     *bytes.Buffer
	Err     *bytes.Buffer
	outLock *sync.Mutex
	width   int
	height  int
}

// NewFakeTerminal ...
func NewFakeTerminal(width, height int) Terminal {
	return &fakeTerm{
		Out:     new(bytes.Buffer),
		Err:     new(bytes.Buffer),
		outLock: &sync.Mutex{},
		width:   width,
		height:  height,
	}
}

func (t *fakeTerm) Width() (width int) {
	return t.width
}

func (t *fakeTerm) Height() (height int) {
	return t.height
}

func (t *fakeTerm) StdOut() io.Writer {
	return t.Out
}

func (t *fakeTerm) StdErr() io.Writer {
	return t.Err
}

func (t *fakeTerm) Print(e interface{}) {
	t.writeString(fmt.Sprintf("%v", e))
}

func (t *fakeTerm) Println(e interface{}) {
	t.writeString(fmt.Sprintf("%v\r\n", e))
}

func (t *fakeTerm) EraseLine() {
	t.writeString(TermControlEraseLine)
}

func (t *fakeTerm) OverwriteLine(e interface{}) {
	t.Print(fmt.Sprintf("%s%v", TermControlEraseLine, e))
}

func (t *fakeTerm) Clear() {
	t.Out.Reset()
}

func (t *fakeTerm) writeString(s string) (n int) {
	t.outLock.Lock()
	defer t.outLock.Unlock()

	if n, err := t.Out.WriteString(s); err == nil {
		return n
	}

	return 0
}
