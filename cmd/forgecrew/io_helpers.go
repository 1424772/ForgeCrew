package main

import (
	"os"
)

// osStdin is os.Stdin — wrapped so tests can replace it.
var osStdin = os.Stdin

// osStdout is os.Stdout — wrapped so tests can reference it.
var osStdout = os.Stdout

// osGetwd wraps os.Getwd for consistency.
func osGetwd() (string, error) {
	return os.Getwd()
}
