package termite

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMatrixWritesToTerminalOutput(t *testing.T) {
	example := generateRandomString()

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
	examples[1] = generateRandomString()
	matrix.NewRow().Update(examples[2])
	row2.WriteString(examples[1])

	assertEventualSequence(t, matrix, expectedRewriteSequenceFor(examples))
}

func TestMatrixRowUpdateTrimsLineFeeds(t *testing.T) {
	expected := generateRandomString()

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
	cancel := matrix.Start()

	return matrix, cancel
}

func generateRandomString() string {
	return fmt.Sprintf("[%d]", rand.Intn(time.Now().Nanosecond()))
}

func generateMultiLineExamples(count int) []string {
	examples := []string{}

	for i := 0; i < count; i++ {
		examples = append(examples, generateRandomString())
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
