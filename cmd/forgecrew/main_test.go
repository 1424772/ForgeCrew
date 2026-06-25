package main

import (
	"bytes"
	"strings"
	"testing"
)

// executeCommand runs the root cobra command with the given args and
// returns combined stdout+stderr output. It is the shared harness for
// all CLI integration tests.
func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	// Clean up after execution to avoid state leaks between tests.
	rootCmd.SetArgs(nil)
	return buf.String(), err
}

func TestRootCommandHelp(t *testing.T) {
	out, err := executeCommand("--help")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "ForgeCrew") {
		t.Error("help output should mention ForgeCrew")
	}
	if !strings.Contains(out, "init") {
		t.Error("help output should list init subcommand")
	}
	if !strings.Contains(out, "scan") {
		t.Error("help output should list scan subcommand")
	}
	if !strings.Contains(out, "version") {
		t.Error("help output should list version subcommand")
	}
}

func TestRootCommandAllSubCommandsRegistered(t *testing.T) {
	expectedCmds := []string{"init", "scan", "agents", "models", "team", "diff", "checkpoint", "version", "task", "lang", "language"}
	for _, name := range expectedCmds {
		cmd, _, _ := rootCmd.Find([]string{name})
		if cmd == nil {
			t.Errorf("subcommand %q not registered", name)
		}
	}
}

func TestRootCommandUnknownSubCommand(t *testing.T) {
	_, err := executeCommand("nonexistent-command-xyz")
	if err == nil {
		t.Error("expected error for unknown subcommand")
	}
}
