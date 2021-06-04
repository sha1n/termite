package termite

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	timeout  = time.Second * 10
	interval = time.Nanosecond
)

func TestSpinnerCharSequence(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)

	spinner := NewSpinner(emulatedStdout, "", interval, DefaultSpinnerFormatter())
	cancel, err := spinner.Start()
	defer cancel()

	assert.NoError(t, err)
	assert.NotNil(t, cancel)

	assertSpinnerCharSequence(t, emulatedStdout)
}

func TestSpinnerCancellation(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, "", interval, DefaultSpinnerFormatter())
	cancel, _ := spin.Start()

	assertSpinnerCharSequence(t, emulatedStdout)

	cancel()
	assertStoppedEventually(t, emulatedStdout, spin.(*spinner))
}

func TestSpinnerStartAlreadyRunning(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, "", interval, DefaultSpinnerFormatter())
	cancel, _ := spin.Start()
	defer cancel()

	_, err := spin.Start()
	assert.Error(t, err)
}

func TestSpinnerStopAlreadyStopped(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, "", interval, DefaultSpinnerFormatter())
	_, _ = spin.Start()
	err := spin.Stop("")
	assert.NoError(t, err)

	assert.Error(t, spin.Stop(""), "expected error")
}

func TestSpinnerStopMessage(t *testing.T) {
	expectedStopMessage := generateRandomString()
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, "", interval, DefaultSpinnerFormatter())
	_, err := spin.Start()
	assert.NoError(t, err)

	err = spin.Stop(expectedStopMessage)
	assert.NoError(t, err)

	assertBufferEventuallyContains(t, emulatedStdout, expectedStopMessage)
	assert.NotContains(t, emulatedStdout.String(), "\n", "line feed is expected!")
}

func TestSpinnerTitle(t *testing.T) {
	expectedTitle := generateRandomString()
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, expectedTitle, interval, DefaultSpinnerFormatter())
	cancel, _ := spin.Start()
	defer cancel()

	assertBufferEventuallyContains(t, emulatedStdout, expectedTitle)
}

func TestSpinnerSetTitle(t *testing.T) {
	expectedInitialTitle := generateRandomString()
	expectedUpdatedTitle := generateRandomString()
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, expectedInitialTitle, interval, DefaultSpinnerFormatter())
	cancel, _ := spin.Start()
	defer cancel()

	assertBufferEventuallyContains(t, emulatedStdout, expectedInitialTitle)

	assert.NoError(t, spin.SetTitle(expectedUpdatedTitle))

	assertBufferEventuallyContains(t, emulatedStdout, expectedUpdatedTitle)
}

func TestSpinnerSetTitleOnStoppedSpinner(t *testing.T) {
	expectedInitialTitle := generateRandomString()
	expectedUpdatedTitle := generateRandomString()
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, expectedInitialTitle, interval, DefaultSpinnerFormatter())
	_, _ = spin.Start()

	assertBufferEventuallyContains(t, emulatedStdout, expectedInitialTitle)

	assert.NoError(t, spin.Stop(""))
	assert.Error(t, spin.SetTitle(expectedUpdatedTitle))
}

func assertBufferEventuallyContains(t *testing.T, outBuffer *bytes.Buffer, expected string) {
	assert.Eventually(
		t,
		bufferContains(outBuffer, expected),
		timeout,
		interval,
	)
}

func bufferContains(outBuffer *bytes.Buffer, expected string) func() bool {
	return func() bool {
		return strings.Contains(outBuffer.String(), expected)
	}
}

func assertStoppedEventually(t *testing.T, outBuffer *bytes.Buffer, spinner *spinner) {
	assert.Eventually(
		t,
		func() bool { return !spinner.isActiveSafe() },
		timeout,
		interval,
	)

	outBuffer.Reset() // clear the buffer

	assert.Eventually(
		t,
		func() bool { return outBuffer.UnreadByte() != nil },
		timeout,
		spinner.interval,
	)
}

func assertSpinnerCharSequence(t *testing.T, outBuffer *bytes.Buffer) {
	charSeq := DefaultSpinnerCharSeq()
	readChars := []string{}

	scan := func() {
		for {
			r, _, e := outBuffer.ReadRune()
			print(string(r), ",")
			if e != nil {
				continue
			}
			readChar := string(r)
			if len(readChars) == 0 && readChar == charSeq[0] {
				readChars = append(readChars, readChar)
			} else if len(readChars) > 0 {
				for _, ch := range charSeq {
					if ch == readChar {
						readChars = append(readChars, ch)
					}

					if len(readChars) == len(charSeq) {
						return
					}
				}
			}
		}
	}

	scan()

	assert.Equal(t, charSeq, readChars)
}
