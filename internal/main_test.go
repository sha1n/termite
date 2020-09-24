package main

import (
	"github.com/sha1n/termite"
	"testing"
)

func TestSanity(t *testing.T) {
	demo(termite.NewFakeTerminal(272, 72))
}
