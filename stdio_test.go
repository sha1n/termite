package termite

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAutoFlushingWriter(t *testing.T) {
	buf := new(bytes.Buffer)
	expected := randomBytes()
	
	writer := NewAutoFlushingWriter(buf)
	writer.Write(expected)

	assert.Equal(t, expected, buf.Bytes())
}

func randomBytes() []byte {
	return []byte(generateRandomString())
}