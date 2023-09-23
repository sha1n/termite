package termite

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/sha1n/gommons/pkg/io"
	"github.com/stretchr/testify/assert"
)

const (
	timeout  = time.Second * 10
	interval = time.Nanosecond
)

func TestSpinnerCharSequence(t *testing.T) {
	probedWriter := io.NewUnlimitedProbedWriter(new(bytes.Buffer))

	spinner := NewSpinner(probedWriter, "", interval, DefaultSpinnerFormatter())
	cancel, err := spinner.Start()
	defer cancel()

	assert.NoError(t, err)
	assert.NotNil(t, cancel)

	assertSpinnerCharSequence(t, probedWriter)
}

func TestSpinnerCancellation(t *testing.T) {
	probedWriter := io.NewUnlimitedProbedWriter(new(bytes.Buffer))

	spin := NewSpinner(probedWriter, "", interval, DefaultSpinnerFormatter())
	cancel, err := spin.Start()

	assert.NoError(t, err)
	assert.NotNil(t, cancel)
	assertSpinnerCharSequence(t, probedWriter)

	cancel()
	assertStoppedEventually(t, probedWriter, spin.(*spinner))
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

func assertStoppedEventually(t *testing.T, probedWriter *io.ProbedWriter, spinner *spinner) {
	assert.Eventually(
		t,
		func() bool { return !spinner.isActiveSafe() },
		timeout,
		interval,
		"expected spinner to deactivate",
	)

	assert.Eventually(
		t,
		func() bool {
			probedWriter.Reset()
			time.Sleep(spinner.interval * 2)
			return len(probedWriter.Bytes()) == 0

		},
		timeout,
		spinner.interval,
		"expected no more output from spinner",
	)
}

func assertSpinnerCharSequence(t *testing.T, probedWriter *io.ProbedWriter) {
	charSeq := DefaultSpinnerCharSeq()
	expectedCharSequence := strings.Join(charSeq, "")
	var read string = ""

	for {
		read = stripControlCharacters(probedWriter.String())
		if len(read) >= len(expectedCharSequence)*2 {
			break
		}
	}

	assert.Contains(t, read, expectedCharSequence)
}

func stripControlCharacters(input string) string {
	controlCharsRegex := regexp.MustCompile(`[[:cntrl:]]|\[|K`)

	return controlCharsRegex.ReplaceAllString(input, "")
}
