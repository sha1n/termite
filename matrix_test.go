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

	matrix, cancel := startMatrix()
	defer cancel()

	matrix.NewLine().WriteString(examples[0])
	matrix.NewLine().WriteString(examples[1])
	matrix.NewLine().WriteString(examples[2])

	assertEventualSequence(t, matrix, examples)
}

func TestMatrixUpdatesTerminalOutput(t *testing.T) {
	examples := generateMultiLineExamples(3)

	matrix, cancel := startMatrix()
	defer cancel()

	matrix.NewLine().WriteString(examples[0])
	line2 := matrix.NewLine()
	line2.WriteString(examples[1])
	examples[1] = generateRandomString()
	matrix.NewLine().WriteString(examples[2])
	line2.WriteString(examples[1])

	assertEventualSequence(t, matrix, examples)
}

func TestMatrixStructure(t *testing.T) {
	examples := generateMultiLineExamples(3)

	matrix, cancel := startMatrix()
	defer cancel()

	matrix.NewLine().WriteString(examples[0])
	matrix.NewLine().WriteString(examples[1])
	matrix.NewLine().WriteString(examples[2])

	assert.Equal(t, examples, matrix.(*terminalMatrix).lines)
}

func assertEventualSequence(t *testing.T, matrix Matrix, examples []string) {
	contantsAllExamplesInOrderFn := func() bool {
		return strings.Contains(
			matrix.Terminal().(*fakeTerm).Out.String(),
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

func startMatrix() (Matrix, context.CancelFunc) {
	term := NewFakeTerminal(80, 80)
	matrix := NewMatrix(term)
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
