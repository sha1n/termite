package termite

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sha1n/gommons/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestMatrixWritesToTerminalOutput(t *testing.T) {
	example := test.RandomString()

	matrix, cancel := startNewMatrix()
	defer cancel()

	matrix.NewRow().Update(example)

	assertEventualSequence(t, matrix, example)
}

func TestMatrixUpdatesTerminalOutput(t *testing.T) {
	examples := generateMultiLineExamples(3)

	matrix, cancel := startNewMatrix()
	defer cancel()

	matrix.NewRow().Update(examples[0])
	row2 := matrix.NewRow()
	row2.WriteString(examples[1])
	examples[1] = test.RandomString()
	matrix.NewRow().Update(examples[2])
	row2.WriteString(examples[1])

	assertEventualSequence(t, matrix, expectedRewriteSequenceFor(examples))
}

func TestMatrixRowUpdateTrimsLineFeeds(t *testing.T) {
	expected := test.RandomString()

	matrix, cancel := startNewMatrix()
	defer cancel()

	matrix.NewRow().Update("\n" + expected + "\n\n\r")

	assertEventualSequence(t, matrix, expectedRewriteSequenceFor([]string{expected}))
}

func TestNoRewritesWhenNothingChanges(t *testing.T) {
	matrix, cancel := startNewMatrix()
	defer cancel()

	matrix.NewRange(4)

	assertEventualSequence(t, matrix, expectedSkipRewriteSequenceFor(4))
}

func TestMatrixStructure(t *testing.T) {
	examples := generateMultiLineExamples(3)

	matrix, cancel := startNewMatrix()
	defer cancel()

	matrix.NewRow().Update(examples[0])
	matrix.NewRow().Update(examples[1])
	matrix.NewRow().Update(examples[2])

	assert.Equal(t, examples, linesOf(matrix))
}

func TestMatrixNewRangeOrder(t *testing.T) {
	examples := generateMultiLineExamples(3)

	matrix, cancel := startNewMatrix()
	defer cancel()

	rows := matrix.NewRange(3)
	for i := 0; i < 3; i++ {
		rows[i].Update(examples[i])
	}

	assert.Equal(t, examples, linesOf(matrix))
}

func TestMatrixGetRowByID(t *testing.T) {
	matrix, cancel := startNewMatrix()
	defer cancel()

	matrix.NewRow()
	aRow := matrix.NewRow()
	matrix.NewRow()
	fetchedRow, err := matrix.GetRowByID(aRow.ID())

	assert.NoError(t, err)
	assert.Equal(t, aRow, fetchedRow)
}

func TestMatrixGetRowByIdWithInvalidRange(t *testing.T) {
	matrix, cancel := startNewMatrix()
	defer cancel()

	_, err := matrix.GetRow(0)

	assert.Error(t, err)
}

func TestMatrixGetRowByIdWithIllegalValue(t *testing.T) {
	matrix, cancel := startNewMatrix()
	defer cancel()

	_, err := matrix.GetRow(-1)

	assert.Error(t, err)
}

func TestMatrixConcurrentStress(t *testing.T) {
	matrix, cancel := startNewMatrix()
	defer cancel()

	count := 100
	rows := matrix.NewRange(count)
	startC := make(chan struct{})
	doneC := make(chan struct{})

	for i := 0; i < count; i++ {
		go func(row MatrixRow) {
			<-startC
			for j := 0; j < 100; j++ {
				row.Update(test.RandomString())
			}
			doneC <- struct{}{}
		}(rows[i])
	}

	close(startC)
	for i := 0; i < count; i++ {
		<-doneC
	}
}

func assertEventualSequence(t *testing.T, matrix Matrix, expected string) {
	contantsAllExamplesInOrderFn := func() bool {
		return strings.Contains(
			matrix.(*matrixImpl).writer.(*bytes.Buffer).String(),
			expected,
		)
	}

	assert.Eventually(t,
		contantsAllExamplesInOrderFn,
		time.Second*10,
		matrix.RefreshInterval(),
	)
}

func expectedSkipRewriteSequenceFor(count int) string {
	return strings.Repeat("\n", count)
}

func expectedRewriteSequenceFor(examples []string) string {
	buf := new(bytes.Buffer)
	for _, e := range examples {
		buf.WriteString(TermControlEraseLine + e + "\n")
	}

	return buf.String()
}

func startNewMatrix() (Matrix, context.CancelFunc) {
	emulatedOutput := new(bytes.Buffer)
	matrix := NewMatrix(emulatedOutput, time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	_ = matrix.Start(ctx)

	return matrix, cancel
}

func generateMultiLineExamples(count int) []string {
	examples := []string{}

	for i := 0; i < count; i++ {
		examples = append(examples, test.RandomString())
	}

	return examples
}

func linesOf(m Matrix) []string {
	rows := m.(*matrixImpl).rows
	lines := make([]string, len(rows))
	for i, row := range rows {
		lines[i] = row.value
	}

	return lines
}

func TestMatrixStartWithCancelledContext(t *testing.T) {
	emulatedOutput := new(bytes.Buffer)
	matrix := NewMatrix(emulatedOutput, time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before starting

	done := matrix.Start(ctx)
	assert.Nil(t, done, "expected nil done channel when context is already cancelled")
}

func TestMatrixStopsWritingAfterCancel(t *testing.T) {
	emulatedOutput := new(bytes.Buffer)
	refreshInterval := time.Millisecond * 10
	matrix := NewMatrix(emulatedOutput, refreshInterval)

	ctx, cancel := context.WithCancel(context.Background())
	_ = matrix.Start(ctx)

	row := matrix.NewRow()
	row.Update("test content")

	time.Sleep(refreshInterval * 3)
	assert.True(t, emulatedOutput.Len() > 0, "expected some output")

	cancel()
	time.Sleep(refreshInterval * 3)
	outputLenAfterCancel := emulatedOutput.Len()

	time.Sleep(refreshInterval * 5)
	assert.Equal(t, outputLenAfterCancel, emulatedOutput.Len(),
		"matrix continued writing after context was cancelled")
}

// TestMatrixCancelWaitsForCompletion verifies that waiting on the done channel
// guarantees no more output will be written - prevents output interleaving.
func TestMatrixCancelWaitsForCompletion(t *testing.T) {
	emulatedOutput := new(bytes.Buffer)
	refreshInterval := time.Millisecond * 10
	matrix := NewMatrix(emulatedOutput, refreshInterval)

	ctx, cancel := context.WithCancel(context.Background())
	done := matrix.Start(ctx)
	assert.NotNil(t, done)

	row := matrix.NewRow()
	row.Update("test content")

	time.Sleep(refreshInterval * 2)
	initialLen := emulatedOutput.Len()
	assert.True(t, initialLen > 0)

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("done channel was not closed within timeout")
	}

	// Verify channel is closed
	select {
	case _, ok := <-done:
		assert.False(t, ok)
	default:
		t.Fatal("done channel should be closed")
	}

	outputLenAfterWait := emulatedOutput.Len()
	assert.GreaterOrEqual(t, outputLenAfterWait, initialLen)

	time.Sleep(refreshInterval * 5)
	assert.Equal(t, outputLenAfterWait, emulatedOutput.Len(),
		"matrix wrote output after done channel closed")
}
