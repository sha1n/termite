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

	spinner := NewSpinner(emulatedStdout, "", interval)
	cancel, err := spinner.Start()
	defer cancel()

	assert.NoError(t, err)
	assert.NotNil(t, cancel)

	assertSpinnerCharSequence(t, emulatedStdout)
}

func TestSpinnerCancellation(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, "", interval)
	cancel, _ := spin.Start()

	assertSpinnerCharSequence(t, emulatedStdout)

	cancel()
	assertStoppedEventually(t, emulatedStdout, spin.(*spinner))
}

func TestSpinnerStartAlreadyRunning(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, "", interval)
	cancel, _ := spin.Start()
	defer cancel()

	_, err := spin.Start()
	assert.Error(t, err)
}

func TestSpinnerStopAlreadyStopped(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, "", interval)
	spin.Start()
	err := spin.Stop("")
	assert.NoError(t, err)

	assert.Error(t, spin.Stop(""), "expected error")
}

func TestSpinnerStopMessage(t *testing.T) {
	expectedStopMessage := generateRandomString()
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, "", interval)
	spin.Start()
	err := spin.Stop(expectedStopMessage)
	assert.NoError(t, err)

	assertBufferEventuallyContains(t, emulatedStdout, expectedStopMessage)
	assert.NotContains(t, emulatedStdout.String(), "\n", "line feed is expected!")
}

func TestSpinnerTitle(t *testing.T) {
	expectedTitle := generateRandomString()
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, expectedTitle, interval)
	cancel, _ := spin.Start()
	defer cancel()

	assertBufferEventuallyContains(t, emulatedStdout, expectedTitle)
}

func TestSpinnerSetTitle(t *testing.T) {
	expectedInitialTitle := generateRandomString()
	expectedUpdatedTitle := generateRandomString()
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, expectedInitialTitle, interval)
	cancel, _ := spin.Start()
	defer cancel()

	assertBufferEventuallyContains(t, emulatedStdout, expectedInitialTitle)

	spin.SetTitle(expectedUpdatedTitle)

	assertBufferEventuallyContains(t, emulatedStdout, expectedUpdatedTitle)
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

// TODO can this be simplified?
func assertSpinnerCharSequence(t *testing.T, outBuffer *bytes.Buffer) {
	readChars := make([]string, len(defaultSpinnerCharSeq))
	readCharsCount := 0

	readSequence := func() string {
		startTime := time.Now()
		for {
			s, _ := outBuffer.ReadString(TermControlEraseLine[len(TermControlEraseLine)-1]) // read everything you got
			if strippedString := strings.Trim(s, TermControlEraseLine); strippedString != "" {
				return strippedString
			}

			// guard again infinite loop
			if time.Now().After(startTime.Add(time.Second * 30)) {
				return ""
			}
		}
	}

	// find the first character in the spinner sequence, so we can validate order properly
	for {
		strippedString := readSequence()
		if strippedString != "" && strippedString == defaultSpinnerCharSeq[0] {
			readChars[0] = strippedString
			break
		}
		// guard against infinite loop caused by bugs
		readCharsCount++
		if readCharsCount > len(defaultSpinnerCharSeq)*2 {
			assert.Fail(t, "something went wrong...")
		}
	}

	for i := 1; i < len(defaultSpinnerCharSeq); i++ {
		readChars[i] = readSequence()
	}

	assert.Equal(t, defaultSpinnerCharSeq, readChars)

}
