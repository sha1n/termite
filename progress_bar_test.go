package termite

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/sha1n/clib/pkg/test"
	"github.com/stretchr/testify/assert"
)

const (
	fakeTerminalWidth = 100
)

func TestFullWidthProgressBar(t *testing.T) {
	testProgressBarWith(t, fakeTerminalWidth, fakeTerminalWidth, fakeTerminalWidth)
}

func TestOversizedProgressBar(t *testing.T) {
	testProgressBarWith(t, fakeTerminalWidth*2, fakeTerminalWidth*2, fakeTerminalWidth/2+rand.Intn(fakeTerminalWidth*2))
}

func TestZeroSizedTerminalProgressBar(t *testing.T) {
	testProgressBarWith(t, 0, fakeTerminalWidth*2, fakeTerminalWidth/2+rand.Intn(fakeTerminalWidth*2))
}

func TestTickAnAlreadyDoneProgressBar(t *testing.T) {
	var emulatedStdout = new(bytes.Buffer)
	bar := NewDefaultProgressBar(emulatedStdout, fakeTerminalWidth, 2)

	assert.True(t, bar.Tick())
	assert.False(t, bar.Tick())
	assert.False(t, bar.Tick())
	assert.True(t, bar.IsDone())
}

func TestStart(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)
	bar := NewDefaultProgressBar(emulatedStdout, fakeTerminalWidth, 2)

	tick, cancel, err := bar.Start()

	assert.NoError(t, err)
	assert.NotNil(t, tick)
	assert.NotNil(t, cancel)

	assert.True(t, tick(test.RandomString()))
	assert.False(t, tick(test.RandomString()))
}

func TestStartWithAlreadyStartedBar(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)
	bar := NewDefaultProgressBar(emulatedStdout, fakeTerminalWidth, 2)

	_, _, err := bar.Start()
	assert.NoError(t, err)

	_, _, err = bar.Start()
	assert.Error(t, err)
}

func TestStartCancel(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)
	bar := NewDefaultProgressBar(emulatedStdout, fakeTerminalWidth, 2)

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
	bar := NewDefaultProgressBar(emulatedStdout, fakeTerminalWidth, 2)

	aRandomMessage := test.RandomString()

	assert.True(t, bar.TickMessage(aRandomMessage))
	assert.NotContains(t, emulatedStdout.String(), aRandomMessage)
}

func TestTickMessage(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)

	aRandomMessage := test.RandomString()
	bar := NewProgressBar(emulatedStdout, 2, 100, 100, DefaultProgressBarFormatterWidth(len(aRandomMessage)))

	assert.True(t, bar.TickMessage(aRandomMessage))
	assert.Contains(t, emulatedStdout.String(), aRandomMessage)
}

func testProgressBarWith(t *testing.T, termWidth, width, maxTicks int) {
	emulatedStdout := new(bytes.Buffer)
	bar := NewProgressBar(emulatedStdout, maxTicks, termWidth, width, DefaultProgressBarFormatter())

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
