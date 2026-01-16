package termite

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/sha1n/gommons/pkg/io"
	"github.com/sha1n/gommons/pkg/test"
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

func TestSpinnerTitles(t *testing.T) {
	t.Run("InitialTitle", func(t *testing.T) {
		expectedTitle := test.RandomString()
		emulatedStdout := new(bytes.Buffer)
		spin := NewSpinner(emulatedStdout, expectedTitle, interval, DefaultSpinnerFormatter())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_ = spin.Start(ctx)

		assertBufferEventuallyContains(t, emulatedStdout, expectedTitle)
	})

	t.Run("SetTitle", func(t *testing.T) {
		expectedInitialTitle := test.RandomString()
		expectedUpdatedTitle := test.RandomString()
		emulatedStdout := new(bytes.Buffer)
		spin := NewSpinner(emulatedStdout, expectedInitialTitle, interval, DefaultSpinnerFormatter())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_ = spin.Start(ctx)

		assertBufferEventuallyContains(t, emulatedStdout, expectedInitialTitle)
		assert.NoError(t, spin.SetTitle(expectedUpdatedTitle))
		assertBufferEventuallyContains(t, emulatedStdout, expectedUpdatedTitle)
	})

	t.Run("SetTitleOnStoppedSpinner", func(t *testing.T) {
		spin := NewSpinner(new(bytes.Buffer), "title", interval, DefaultSpinnerFormatter())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_ = spin.Start(ctx)
		_ = spin.Stop("")

		assert.Error(t, spin.SetTitle("new title"))
	})
}

func TestSpinnerErrors(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "StartAlreadyRunning",
			run: func(t *testing.T) {
				spin := NewSpinner(new(bytes.Buffer), "", interval, DefaultSpinnerFormatter())
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				_ = spin.Start(ctx)
				assert.Error(t, spin.Start(ctx))
			},
		},
		{
			name: "StopAlreadyStopped",
			run: func(t *testing.T) {
				spin := NewSpinner(new(bytes.Buffer), "", interval, DefaultSpinnerFormatter())
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				_ = spin.Start(ctx)
				assert.NoError(t, spin.Stop(""))
				assert.Error(t, spin.Stop(""))
			},
		},
		{
			name: "StartWithCancelledContext",
			run: func(t *testing.T) {
				spin := NewSpinner(new(bytes.Buffer), "", interval, DefaultSpinnerFormatter())
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				err := spin.Start(ctx)
				assert.Error(t, err)
				assert.Equal(t, context.Canceled, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSpinnerStopMessage(t *testing.T) {
	expectedStopMessage := test.RandomString()
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
