package termite

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSpinnerCharSequence(t *testing.T) {
	fakeTerminal := NewFakeTerminal(80, 80)

	spinner := NewSpinner(fakeTerminal, 1)
	cancel, err := spinner.Start()
	defer cancel()

	assert.NoError(t, err)
	assert.NotNil(t, cancel)

	assertSpinnerCharSequence(t, fakeTerminal)
}

func TestSpinnerCancellation(t *testing.T) {
	fakeTerminal := NewFakeTerminal(80, 80)

	spin := NewSpinner(fakeTerminal, 1)
	cancel, _ := spin.Start()

	assertSpinnerCharSequence(t, fakeTerminal)

	cancel()
	assertStoppedEventually(t, fakeTerminal, spin.(*spinner))
}

func TestSpinnerStartAlreadyRunning(t *testing.T) {
	fakeTerminal := NewFakeTerminal(80, 80)

	spin := NewSpinner(fakeTerminal, 1)
	cancel, _ := spin.Start()
	defer cancel()

	_, err := spin.Start()
	assert.Error(t, err)
}

func TestSpinnerStopAlreadyStopped(t *testing.T) {
	fakeTerminal := NewFakeTerminal(80, 80)

	spin := NewSpinner(fakeTerminal, 1)
	spin.Start()
	err := spin.Stop("")
	assert.NoError(t, err)

	assert.Error(t, spin.Stop(""), "expected error")
}

func assertStoppedEventually(t *testing.T, fakeTerminal Terminal, spinner *spinner) {
	termOutput := (fakeTerminal.(*fakeTerm)).Out
	startTime := time.Now()

	for spinner.isActive() {
		// guard against infinite loop
		if time.Now().After(startTime.Add(spinner.interval + time.Second)) {
			break
		}
	}

	termOutput.Reset() // clear the buffer
	assert.False(t, spinner.isActive())

	time.Sleep(spinner.interval * 10)
	assert.Error(t, termOutput.UnreadByte())
}

func assertSpinnerCharSequence(t *testing.T, fakeTerminal Terminal) {
	termOutput := (fakeTerminal.(*fakeTerm)).Out
	readChars := make([]string, 4)
	readCharsCount := 0

	readSequence := func() string {
		startTime := time.Now()
		for {
			s, _ := termOutput.ReadString('\n')
			if s == "" {
				continue
			}
			strippedString := strings.TrimSpace(s)
			strippedString = strings.Trim(strippedString, TermControlEraseLine)

			// guard again infinite loop
			if time.Now().After(startTime.Add(time.Second * 30)) {
				return ""
			}

			return strippedString
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
