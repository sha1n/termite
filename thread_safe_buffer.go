package termite

import (
    "bytes"
    "sync"
)

type ThreadSafeBufferWriter struct {
    buf    *bytes.Buffer
    mutex  sync.Mutex
}

func NewThreadSafeBuffer() *ThreadSafeBufferWriter {
    return &ThreadSafeBufferWriter{
        buf: new(bytes.Buffer),
    }
}

func (b *ThreadSafeBufferWriter) Write(p []byte) (n int, err error) {
    b.mutex.Lock()
    defer b.mutex.Unlock()
    return b.buf.Write(p)
}

func (b *ThreadSafeBufferWriter) String() string {
    b.mutex.Lock()
    defer b.mutex.Unlock()
    return b.buf.String()
}

func (b *ThreadSafeBufferWriter) Len() int {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.buf.Len()
}

func (b *ThreadSafeBufferWriter) Reset() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.buf.Reset()
}
