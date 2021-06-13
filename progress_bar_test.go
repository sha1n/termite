package termite

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/sha1n/clib/pkg/test"
	"github.com/stretchr/testify/assert"
)

var (
	fakeTerminalWidth   = 100
	fakeTerminalWidthFn = func() int { return 100 }
)

func TestFullWidthProgressBar(t *testing.T) {
	testProgressBarWith(t, fakeTerminalWidthFn, fakeTerminalWidth, fakeTerminalWidth)
}

func TestOversizedProgressBar(t *testing.T) {
	testProgressBarWith(t, fakeTerminalWidthFn, fakeTerminalWidth*2, fakeTerminalWidth/2+rand.Intn(fakeTerminalWidth*2))
}

func TestZeroSizedTerminalProgressBar(t *testing.T) {
	testProgressBarWith(t, func() int { return 0 }, fakeTerminalWidth*2, fakeTerminalWidth/2+rand.Intn(fakeTerminalWidth*2))
}

func TestTickAnAlreadyDoneProgressBar(t *testing.T) {
	var emulatedStdout = new(bytes.Buffer)
	bar := NewDefaultProgressBar(emulatedStdout, 2, fakeTerminalWidthFn)

	assert.True(t, bar.Tick())
	assert.False(t, bar.Tick())
	assert.False(t, bar.Tick())
	assert.True(t, bar.IsDone())
}

func TestStart(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)
	bar := NewDefaultProgressBar(emulatedStdout, 2, fakeTerminalWidthFn)

	tick, cancel, err := bar.Start()

	assert.NoError(t, err)
	assert.NotNil(t, tick)
	assert.NotNil(t, cancel)

	assert.True(t, tick(test.RandomString()))
	assert.False(t, tick(test.RandomString()))
}

func TestStartWithAlreadyStartedBar(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)
	bar := NewDefaultProgressBar(emulatedStdout, 2, fakeTerminalWidthFn)

	_, _, err := bar.Start()
	assert.NoError(t, err)

	_, _, err = bar.Start()
	assert.Error(t, err)
}

func TestStartCancel(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)
	bar := NewDefaultProgressBar(emulatedStdout, 2, fakeTerminalWidthFn)

	tick, cancel, err := bar.Start()

	assert.NoError(t, err)
	assert.NotNil(t, tick)
	assert.NotNil(t, cancel)

	assert.True(t, tick(test.RandomString()))
	cancel()
	assert.False(t, tick(test.RandomString()))
}

func TestTickMessageNotDisplayedIfWidthIsZero(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)
	bar := NewDefaultProgressBar(emulatedStdout, fakeTerminalWidth, fakeTerminalWidthFn)

	aRandomMessage := test.RandomString()

	assert.True(t, bar.TickMessage(aRandomMessage))
	assert.NotContains(t, emulatedStdout.String(), aRandomMessage)
}

func TestTickMessage(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)

	aRandomMessage := test.RandomString()
	bar := NewProgressBar(emulatedStdout, 2, fakeTerminalWidthFn, 100, DefaultProgressBarFormatterWidth(len(aRandomMessage)))

	assert.True(t, bar.TickMessage(aRandomMessage))
	assert.Contains(t, emulatedStdout.String(), aRandomMessage)
}

func testProgressBarWith(t *testing.T, termWidthFn func() int, width, maxTicks int) {
	emulatedStdout := new(bytes.Buffer)
	bar := NewProgressBar(emulatedStdout, maxTicks, termWidthFn, width, DefaultProgressBarFormatter())

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
