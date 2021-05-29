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
	examples := generateMultiLineExamples(3)

	matrix, cancel := startNewMatrix()
	defer cancel()

	matrix.NewRow().Update(examples[0])
	matrix.NewLineWriter().Write([]byte(examples[1]))
	matrix.NewLineStringWriter().WriteString(examples[2])

	assertEventualSequence(t, matrix, examples)
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

	assertEventualSequence(t, matrix, examples)
}

func TestMatrixRowUpdateTrimsLineFeeds(t *testing.T) {
	expected := generateRandomString()

	matrix, cancel := startNewMatrix()
	defer cancel()

	matrix.NewRow().Update("\r\n" + expected + "\r\n\r\n\r")

	assertEventualSequence(t, matrix, []string{expected})
}

func TestMatrixStructure(t *testing.T) {
	examples := generateMultiLineExamples(3)

	matrix, cancel := startNewMatrix()
	defer cancel()

	matrix.NewRow().Update(examples[0])
	matrix.NewRow().Update(examples[1])
	matrix.NewRow().Update(examples[2])

	assert.Equal(t, examples, matrix.(*matrixImpl).lines)
}

func TestMatrixNewRangeOrder(t *testing.T) {
	examples := generateMultiLineExamples(3)

	matrix, cancel := startNewMatrix()
	defer cancel()

	rows := matrix.NewRange(3)
	for i := 0; i < 3; i++ {
		rows[i].Update(examples[i])
	}

	assert.Equal(t, examples, matrix.(*matrixImpl).lines)
}

func TestWriterLineInterface(t *testing.T) {
	example := generateRandomString()

	matrix1, cancel1 := startNewMatrix()
	defer cancel1()

	matrix2, cancel2 := startNewMatrix()
	defer cancel2()

	matrix1.NewLineStringWriter().WriteString(example)
	matrix2.NewLineWriter().Write([]byte(example))

	assert.Equal(t, matrix1.(*matrixImpl).lines, matrix2.(*matrixImpl).lines)
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

func assertEventualSequence(t *testing.T, matrix Matrix, examples []string) {
	contantsAllExamplesInOrderFn := func() bool {
		return strings.Contains(
			matrix.(*matrixImpl).writer.(*FakeTerminal).Out.String(),
			expectedOutputSequenceFor(examples),
		)
	}

	assert.Eventually(t,
		contantsAllExamplesInOrderFn,
		time.Second*10,
		matrix.RefreshInterval(),
	)

}
func expectedOutputSequenceFor(examples []string) string {
	buf := new(bytes.Buffer)
	for _, e := range examples {
		buf.WriteString(TermControlEraseLine + e + "\r\n")
	}

	return buf.String()
}

func startNewMatrix() (Matrix, context.CancelFunc) {
	term := NewFakeTerminal(80, 80)
	matrix := NewMatrix(term, time.Millisecond)
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
