package termite

import (
	"bytes"
	"context"
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
	ctx, cancel := context.WithCancel(context.Background())
	err := spinner.Start(ctx)
	defer cancel()

	assert.NoError(t, err)
	assert.NotNil(t, cancel)

	assertSpinnerCharSequence(t, probedWriter)
}

func TestSpinnerCancellation(t *testing.T) {
	probedWriter := io.NewUnlimitedProbedWriter(new(bytes.Buffer))

	spin := NewSpinner(probedWriter, "", interval, DefaultSpinnerFormatter())
	ctx, cancel := context.WithCancel(context.Background())
	err := spin.Start(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, cancel)
	assertSpinnerCharSequence(t, probedWriter)

	cancel()
	assertStoppedEventually(t, probedWriter, spin.(*spinner))
}

func TestSpinnerStartAlreadyRunning(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, "", interval, DefaultSpinnerFormatter())
	ctx, cancel := context.WithCancel(context.Background())
	_ = spin.Start(ctx)
	defer cancel()

	err := spin.Start(ctx)
	assert.Error(t, err)
}

func TestSpinnerStopAlreadyStopped(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, "", interval, DefaultSpinnerFormatter())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = spin.Start(ctx)

	err := spin.Stop("")
	assert.NoError(t, err)

	assert.Error(t, spin.Stop(""), "expected error on second stop")
}

func TestSpinnerStopMessage(t *testing.T) {
	expectedStopMessage := generateRandomString()
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, "", interval, DefaultSpinnerFormatter())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := spin.Start(ctx)
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = spin.Start(ctx)

	assertBufferEventuallyContains(t, emulatedStdout, expectedTitle)
}

func TestSpinnerSetTitle(t *testing.T) {
	expectedInitialTitle := generateRandomString()
	expectedUpdatedTitle := generateRandomString()
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, expectedInitialTitle, interval, DefaultSpinnerFormatter())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = spin.Start(ctx)

	assertBufferEventuallyContains(t, emulatedStdout, expectedInitialTitle)

	assert.NoError(t, spin.SetTitle(expectedUpdatedTitle))

	assertBufferEventuallyContains(t, emulatedStdout, expectedUpdatedTitle)
}

func TestSpinnerSetTitleOnStoppedSpinner(t *testing.T) {
	expectedInitialTitle := generateRandomString()
	expectedUpdatedTitle := generateRandomString()
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, expectedInitialTitle, interval, DefaultSpinnerFormatter())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = spin.Start(ctx)

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

func assertStoppedEventually(t *testing.T, probedWriter *io.ProbedWriter, spin *spinner) {
	// Verify the spinner has stopped by checking no more output is produced
	assert.Eventually(
		t,
		func() bool {
			probedWriter.Reset()
			time.Sleep(spin.interval * 2)
			return len(probedWriter.Bytes()) == 0

		},
		timeout,
		spin.interval,
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

func TestSpinnerStartWithCancelledContext(t *testing.T) {
	emulatedStdout := new(bytes.Buffer)

	spin := NewSpinner(emulatedStdout, "", interval, DefaultSpinnerFormatter())
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before starting

	err := spin.Start(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
