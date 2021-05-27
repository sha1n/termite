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
	t.WriteString(fmt.Sprintf("%v", e))
}

func (t *fakeTerm) Println(e interface{}) {
	t.WriteString(fmt.Sprintf("%v\r\n", e))
}

func (t *fakeTerm) WriteString(s string) (n int, err error) {
	t.outLock.Lock()
	defer t.outLock.Unlock()

	return t.Out.WriteString(s)
}
