package main

import (
	"bytes"
	"testing"

	"github.com/sha1n/termite"
)

func TestSanity(t *testing.T) {
	ctx, teardown := setupContext()
	defer teardown()

	demo(ctx)
}

func setupContext() (c *demoContext, teardown func()) {
	origStdout := termite.StdoutWriter
	emulatedStdout := termite.NewAutoFlushingWriter(new(bytes.Buffer))

	teardown = func() { termite.StdoutWriter = origStdout }

	termite.StdoutWriter = emulatedStdout
	c = &demoContext{
		out:       emulatedStdout,
		termWidth: 80,
	}

	return c, teardown
}
