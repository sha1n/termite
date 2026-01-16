package termite

import (
	"bytes"
	"testing"

	"github.com/sha1n/gommons/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestNewAutoFlushingWriter(t *testing.T) {
	buf := new(bytes.Buffer)
	expected := randomBytes()

	writer := NewAutoFlushingWriter(buf)
	writer.Write(expected)

	assert.Equal(t, expected, buf.Bytes())
}

// The purpose of this test is to ensure that WriteString also flushes the buffer
// and has been introduced to reproduce and solve a bug.
func TestWriteString(t *testing.T) {
	buf := new(bytes.Buffer)
	example := test.RandomString()
	expected := []byte(example)

	writer := NewAutoFlushingWriter(buf)
	writer.WriteString(example)

	assert.Equal(t, expected, buf.Bytes())
}

func randomBytes() []byte {
	return []byte(test.RandomString())
}
