package main

import (
	"bytes"
	"strings"
	"testing"
)

// executeTaskCmd creates a fresh task command (not via rootCmd) to avoid
// state pollution from cobra's internal arg/flag handling across tests.
func executeTaskCmd(args ...string) (string, error) {
	taskDryRun = false
	cmd := taskCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestTaskDryRun(t *testing.T) {
	out, err := executeTaskCmd("test-goal", "--dry-run")
	if err != nil {
		t.Fatalf("task --dry-run failed: %v", err)
	}
	// Default locale is zh (no settings.yaml in test temp dir).
	if !strings.Contains(out, "test-goal") {
		t.Error("output should contain the goal")
	}
	if !strings.Contains(out, "[演习]") {
		t.Error("output should mark [演习] (zh)")
	}
	// In zh: "第 1 轮:"
	if !strings.Contains(out, "第") || !strings.Contains(out, "轮") {
		t.Error("output should show Chinese iteration headers")
	}
	// zh output must NOT contain English placeholder text.
	if strings.Contains(out, "would execute") {
		t.Error("zh output should not contain 'would execute'")
	}
}

func TestTaskRequiresArg(t *testing.T) {
	_, err := executeTaskCmd()
	if err == nil {
		t.Error("task without goal should fail")
	}
}

func TestTaskHelp(t *testing.T) {
	out, err := executeTaskCmd("--help")
	if err != nil {
		t.Fatalf("task --help failed: %v", err)
	}
	if !strings.Contains(out, "--dry-run") {
		t.Error("help should document --dry-run flag")
	}
	if !strings.Contains(out, "goal") {
		t.Error("help should mention goal argument")
	}
}

func TestTaskAllStates(t *testing.T) {
	out, err := executeTaskCmd("verify-states", "--dry-run")
	if err != nil {
		t.Fatalf("task --dry-run failed: %v", err)
	}
	if out == "" {
		t.Fatal("empty output from task --dry-run")
	}
	// All executed states should appear (plan through commit_memory).
	expectedStates := []string{"plan", "retrieve", "act", "observe", "reflect", "improve", "review", "commit_memory"}
	for _, state := range expectedStates {
		if !strings.Contains(out, state) {
			t.Errorf("output should contain state %q", state)
		}
	}
}
