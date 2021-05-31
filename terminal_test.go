package termite

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTerminalDimensionsReturnsErrorWhenThereIsNoTeleTypewriter(t *testing.T) {
	if Tty {
		t.Skipf("This test cannot run with TTY")
	}
	_, _, expectedErr := GetTerminalDimensions()

	assert.Error(t, expectedErr)
}
