package termite

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	fakeTerminalWidth  = 100
	fakeTerminalHeight = 100
)

var fakeTerminal = NewFakeTerminal(fakeTerminalWidth, fakeTerminalHeight)

func TestFullWidthProgressBar(t *testing.T) {
	testProgressBarWith(t, fakeTerminalWidth, fakeTerminalWidth)
}

func TestOversizedProgressBar(t *testing.T) {
	testProgressBarWith(t, fakeTerminalWidth*2, fakeTerminalWidth/2+rand.Intn(fakeTerminalWidth*2))
}

func TestTickAnAlreadyDoneProgressBar(t *testing.T) {
	bar := NewDefaultProgressBar(fakeTerminal, 2)

	assert.True(t, bar.Tick())
	assert.False(t, bar.Tick())
	assert.False(t, bar.Tick())
	assert.True(t, bar.IsDone())
}

func TestStart(t *testing.T) {
	bar := NewDefaultProgressBar(fakeTerminal, 2)

	tick, cancel, err := bar.Start()

	assert.NoError(t, err)
	assert.NotNil(t, tick)
	assert.NotNil(t, cancel)

	assert.True(t, tick())
	assert.False(t, tick())
}

func TestStartWithAlreadyStartedBar(t *testing.T) {
	bar := NewDefaultProgressBar(fakeTerminal, 2)

	_, _, err := bar.Start()
	assert.NoError(t, err)

	_, _, err = bar.Start()
	assert.Error(t, err)
}

func TestStartCancel(t *testing.T) {
	bar := NewDefaultProgressBar(fakeTerminal, 2)

	tick, cancel, err := bar.Start()

	assert.NoError(t, err)
	assert.NotNil(t, tick)
	assert.NotNil(t, cancel)

	assert.True(t, tick())
	cancel()
	assert.False(t, tick())
}

func testProgressBarWith(t *testing.T, width, maxTicks int) {
	bar := NewProgressBar(fakeTerminal, maxTicks, width, width, '|', '-', '|')

	var count = 0
	for {
		if !bar.Tick() {
			break
		}
		count++
	}

	assert.True(t, bar.IsDone())
	assert.Equal(t, maxTicks-1, count)
}
