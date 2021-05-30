package termite

import (
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
	fakeTerminal := NewFakeTerminal(80, 80)

	spinner := NewSpinner(fakeTerminal.Out, "", interval)
	cancel, err := spinner.Start()
	defer cancel()

	assert.NoError(t, err)
	assert.NotNil(t, cancel)

	assertSpinnerCharSequence(t, fakeTerminal)
}

func TestSpinnerCancellation(t *testing.T) {
	fakeTerminal := NewFakeTerminal(80, 80)

	spin := NewSpinner(fakeTerminal.Out, "", interval)
	cancel, _ := spin.Start()

	assertSpinnerCharSequence(t, fakeTerminal)

	cancel()
	assertStoppedEventually(t, fakeTerminal, spin.(*spinner))
}

func TestSpinnerStartAlreadyRunning(t *testing.T) {
	fakeTerminal := NewFakeTerminal(80, 80)

	spin := NewSpinner(fakeTerminal.Out, "", interval)
	cancel, _ := spin.Start()
	defer cancel()

	_, err := spin.Start()
	assert.Error(t, err)
}

func TestSpinnerStopAlreadyStopped(t *testing.T) {
	fakeTerminal := NewFakeTerminal(80, 80)

	spin := NewSpinner(fakeTerminal.Out, "", interval)
	spin.Start()
	err := spin.Stop("")
	assert.NoError(t, err)

	assert.Error(t, spin.Stop(""), "expected error")
}

func TestSpinnerStopMessage(t *testing.T) {
	expectedStopMessage := generateRandomString()
	fakeTerminal := NewFakeTerminal(80, 80)

	spin := NewSpinner(fakeTerminal.Out, "", interval)
	spin.Start()
	err := spin.Stop(expectedStopMessage)
	assert.NoError(t, err)

	assertBufferEventuallyContains(t, fakeTerminal, expectedStopMessage)
	assert.NotContains(t, fakeTerminal.Out.String(), "\n", "line feed is expected!")
}

func TestSpinnerTitle(t *testing.T) {
	expectedTitle := generateRandomString()
	fakeTerminal := NewFakeTerminal(80, 80)

	spin := NewSpinner(fakeTerminal.Out, expectedTitle, interval)
	cancel, _ := spin.Start()
	defer cancel()

	assertBufferEventuallyContains(t, fakeTerminal, expectedTitle)
}

func TestSpinnerSetTitle(t *testing.T) {
	expectedInitialTitle := generateRandomString()
	expectedUpdatedTitle := generateRandomString()
	fakeTerminal := NewFakeTerminal(80, 80)

	spin := NewSpinner(fakeTerminal.Out, expectedInitialTitle, interval)
	cancel, _ := spin.Start()
	defer cancel()

	assertBufferEventuallyContains(t, fakeTerminal, expectedInitialTitle)

	spin.SetTitle(expectedUpdatedTitle)

	assertBufferEventuallyContains(t, fakeTerminal, expectedUpdatedTitle)
}

func assertBufferEventuallyContains(t *testing.T, fakeTerminal *FakeTerminal, expected string) {
	assert.Eventually(
		t,
		bufferContains(fakeTerminal, expected),
		timeout,
		interval,
	)
}

func bufferContains(fakeTerminal *FakeTerminal, expected string) func() bool {
	return func() bool {
		return strings.Contains(fakeTerminal.Out.String(), expected)
	}
}

func assertStoppedEventually(t *testing.T, fakeTerminal *FakeTerminal, spinner *spinner) {
	termOutput := fakeTerminal.Out

	assert.Eventually(
		t,
		func() bool { return !spinner.isActiveSafe() },
		timeout,
		interval,
	)

	termOutput.Reset() // clear the buffer

	assert.Eventually(
		t,
		func() bool { return termOutput.UnreadByte() != nil },
		timeout,
		spinner.interval,
	)
}

// TODO can this be simplified?
func assertSpinnerCharSequence(t *testing.T, fakeTerminal *FakeTerminal) {
	termOutput := fakeTerminal.Out
	readChars := make([]string, 4)
	readCharsCount := 0

	readSequence := func() string {
		startTime := time.Now()
		for {
			s, _ := termOutput.ReadString(TermControlEraseLine[len(TermControlEraseLine)-1]) // read everything you got
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
		if strippedString != "" && strippedString == defaultSpinnerCharacters[0] {
			readChars[0] = strippedString
			break
		}
		// guard against infinite loop caused by bugs
		readCharsCount++
		if readCharsCount > 8 {
			assert.Fail(t, "something went wrong...")
		}
	}

	readChars[1] = readSequence()
	readChars[2] = readSequence()
	readChars[3] = readSequence()

	assert.Equal(t, defaultSpinnerCharacters, readChars)

}
