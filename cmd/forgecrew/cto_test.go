package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/gitops"
	"github.com/1424772/ForgeCrew/internal/runs"
	"github.com/spf13/cobra"
)

// setupTestRun creates a run with a task in a temp directory for review testing.
func setupTestRun(t *testing.T, taskID, status string, changedFiles []string, withHandoff bool, handoffContent string) (string, *runs.Manager) {
	t.Helper()
	chdirTempDir(t)
	config.EnsureDir(".forgecrew")
	m := runs.NewManager()
	run, _ := m.CreateRun()

	nowStr := time.Now().UTC().Format(time.RFC3339)
	task := &runs.Task{
		TaskID:       taskID,
		AgentID:      "backend_engineer",
		Role:         "coding",
		Status:       status,
		StartedAt:    nowStr,
		CompletedAt:  nowStr,
		Goal:         "Test task",
		Summary:      "Test summary",
		ChangedFiles: changedFiles,
		Handoff: runs.HandoffStatus{
			ReadyForReview: status == "ready_for_review",
			Notes:          "Test notes",
		},
	}
	m.SaveTask(run.RunID, task)

	if withHandoff {
		m.SaveHandoff(run.RunID, taskID, handoffContent)
	}

	return run.RunID, m
}

// makeReviewCmd creates a cobra command for testing runCTOReviewWithDeps.
func makeReviewCmd() (*cobra.Command, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	return cmd, buf
}

func TestCTOReviewApproved(t *testing.T) {
	runID, _ := setupTestRun(t, "task-001", "ready_for_review",
		[]string{"pkg/x.go", "pkg/x_test.go"},
		true, "# Handoff\ntest handoff content")

	deps := ctoDeps{
		changedFiles: func() ([]string, error) { return []string{"pkg/x.go", "pkg/x_test.go"}, nil },
		readDiff:     func() (string, error) { return "diff content", nil },
		runTest:      func() (string, error) { return "ok", nil },
		runVet:       func() (string, error) { return "", nil },
	}

	cmd, buf := makeReviewCmd()
	err := runCTOReviewWithDeps(cmd, "task-001", runID, deps)
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "批准") && !strings.Contains(out, "approved") {
		t.Errorf("expected approved status, got: %s", out)
	}
	if !strings.Contains(out, "允许合并: true") && !strings.Contains(out, "MergeAllowed") {
		t.Errorf("expected merge_allowed=true, got: %s", out)
	}

	// Verify review file was written.
	m := runs.NewManager()
	path := filepath.Join(".forgecrew", "runs", runID, "reviews", "task-001-cto.yaml")
	if !config.FileExists(path) {
		t.Errorf("review file not found at %s", path)
	}
	_ = m
}

func TestCTOReviewNotReady(t *testing.T) {
	runID, _ := setupTestRun(t, "task-002", "in_progress",
		[]string{"pkg/x.go"},
		true, "# Handoff\ntest")

	deps := ctoDeps{
		changedFiles: func() ([]string, error) { return []string{"pkg/x.go"}, nil },
		readDiff:     func() (string, error) { return "", nil },
		runTest:      func() (string, error) { return "ok", nil },
		runVet:       func() (string, error) { return "", nil },
	}

	cmd, _ := makeReviewCmd()
	err := runCTOReviewWithDeps(cmd, "task-002", runID, deps)
	if err == nil {
		t.Fatal("expected error for non-ready_for_review task")
	}
	if !strings.Contains(err.Error(), "ready_for_review") {
		t.Errorf("error should mention ready_for_review, got: %v", err)
	}
}

func TestCTOReviewMissingHandoff(t *testing.T) {
	runID, _ := setupTestRun(t, "task-003", "ready_for_review",
		[]string{"pkg/x.go"},
		false, "") // no handoff

	deps := ctoDeps{
		changedFiles: func() ([]string, error) { return []string{"pkg/x.go"}, nil },
		readDiff:     func() (string, error) { return "", nil },
		runTest:      func() (string, error) { return "", fmt.Errorf("should not be called") },
		runVet:       func() (string, error) { return "", nil },
	}

	cmd, buf := makeReviewCmd()
	err := runCTOReviewWithDeps(cmd, "task-003", runID, deps)
	if err != nil {
		t.Fatalf("review should not return error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "blocked") && !strings.Contains(out, "缺少 handoff") {
		t.Errorf("expected blocked status due to missing handoff, got: %s", out)
	}
}

func TestCTOReviewNoChangedFiles(t *testing.T) {
	runID, _ := setupTestRun(t, "task-004", "ready_for_review",
		nil, // no changed files
		true, "# Handoff\ntest")

	deps := ctoDeps{
		changedFiles: func() ([]string, error) { return nil, nil },
		readDiff:     func() (string, error) { return "", nil },
		runTest:      func() (string, error) { return "", fmt.Errorf("should not be called") },
		runVet:       func() (string, error) { return "", nil },
	}

	cmd, buf := makeReviewCmd()
	err := runCTOReviewWithDeps(cmd, "task-004", runID, deps)
	if err != nil {
		t.Fatalf("review should not return error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "blocked") && !strings.Contains(out, "changed_files") {
		t.Errorf("expected blocked status due to no changed files, got: %s", out)
	}
}

func TestCTOReviewTestFailure(t *testing.T) {
	runID, _ := setupTestRun(t, "task-005", "ready_for_review",
		[]string{"pkg/x.go"},
		true, "# Handoff\ntest")

	deps := ctoDeps{
		changedFiles: func() ([]string, error) { return []string{"pkg/x.go"}, nil },
		readDiff:     func() (string, error) { return "diff", nil },
		runTest:      func() (string, error) { return "", fmt.Errorf("tests failed") },
		runVet:       func() (string, error) { return "", nil },
	}

	cmd, buf := makeReviewCmd()
	err := runCTOReviewWithDeps(cmd, "task-005", runID, deps)
	if err != nil {
		t.Fatalf("review should not return error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "changes_requested") {
		t.Errorf("expected changes_requested status, got: %s", out)
	}
}

func TestRunsListEmpty(t *testing.T) {
	chdirTempDir(t)
	os.MkdirAll(".forgecrew", 0755)

	out, err := executeCommand("runs", "list")
	if err != nil {
		t.Fatalf("runs list failed: %v", err)
	}
	if out == "" {
		t.Error("runs list should output a message even when empty")
	}
	if !strings.Contains(out, "没有") && !strings.Contains(out, "run") {
		t.Errorf("runs list (empty) should show hint, got: %s", out)
	}
}

func TestRunsStatusNotFound(t *testing.T) {
	chdirTempDir(t)

	_, err := executeCommand("runs", "status", "nonexistent-run")
	if err == nil {
		t.Fatal("expected error for nonexistent run")
	}
}

func TestHandoffSubmitCreatesRunAndTask(t *testing.T) {
	chdirTempDir(t)

	out, err := executeCommand("handoff", "submit", "test-task",
		"--agent", "backend_engineer",
		"--goal", "Test goal",
		"--summary", "Test summary",
		"--notes", "Test notes",
	)
	if err != nil {
		t.Fatalf("handoff submit failed: %v", err)
	}
	if !strings.Contains(out, "提交 handoff") {
		t.Errorf("expected handoff submission message, got: %s", out)
	}
	if !strings.Contains(out, "test-task") {
		t.Errorf("output should contain task ID, got: %s", out)
	}

	// Verify run and task were created on disk.
	m := runs.NewManager()
	runsList, err := m.ListRuns()
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runsList) == 0 {
		t.Fatal("expected at least one run to be created")
	}

	runID := runsList[0].RunID
	task, err := m.GetTask(runID, "test-task")
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if task.AgentID != "backend_engineer" {
		t.Errorf("agent = %q, want backend_engineer", task.AgentID)
	}
	if task.Status != "ready_for_review" {
		t.Errorf("status = %q, want ready_for_review", task.Status)
	}

	// Verify runs list now shows the run.
	out, err = executeCommand("runs", "list")
	if err != nil {
		t.Fatalf("runs list after submit: %v", err)
	}
	if !strings.Contains(out, runID) {
		t.Errorf("runs list should contain run ID %s, got: %s", runID, out)
	}
}

func TestHandoffSubmitWithoutForcePreservesHandoff(t *testing.T) {
	chdirTempDir(t)

	// First submission.
	_, err := executeCommand("handoff", "submit", "test-task",
		"--agent", "backend_engineer",
		"--goal", "Original goal",
		"--summary", "Original summary",
		"--notes", "Original notes",
	)
	if err != nil {
		t.Fatalf("first submit: %v", err)
	}

	// Read the handoff content after first submission.
	m := runs.NewManager()
	runsList, _ := m.ListRuns()
	runID := runsList[0].RunID
	originalHandoff, _ := m.ReadHandoff(runID, "test-task")

	// Second submission without --force.
	_, err = executeCommand("handoff", "submit", "test-task",
		"--agent", "backend_engineer",
		"--goal", "New goal",
		"--summary", "New summary",
	)
	if err != nil {
		t.Fatalf("second submit: %v", err)
	}

	// Handoff content should be preserved (no --force).
	currentHandoff, _ := m.ReadHandoff(runID, "test-task")
	if currentHandoff != originalHandoff {
		t.Error("handoff content was overwritten without --force flag")
	}

	// But task fields should be updated.
	task, _ := m.GetTask(runID, "test-task")
	if task.Status != "ready_for_review" {
		t.Errorf("task status should be updated, got %q", task.Status)
	}
}

func TestHandoffSubmitWithForceOverwrites(t *testing.T) {
	chdirTempDir(t)

	// First submission.
	_, err := executeCommand("handoff", "submit", "test-task",
		"--agent", "backend_engineer",
		"--goal", "Original goal",
		"--summary", "Original summary",
	)
	if err != nil {
		t.Fatalf("first submit: %v", err)
	}

	// Second submission with --force.
	_, err = executeCommand("handoff", "submit", "test-task",
		"--agent", "qa_engineer",
		"--goal", "New goal",
		"--summary", "New summary",
		"--force",
	)
	if err != nil {
		t.Fatalf("second submit with --force: %v", err)
	}

	// Task should reflect new values.
	m := runs.NewManager()
	runsList, _ := m.ListRuns()
	runID := runsList[0].RunID
	task, _ := m.GetTask(runID, "test-task")
	if task.AgentID != "qa_engineer" {
		t.Errorf("agent should be qa_engineer after --force, got %q", task.AgentID)
	}
	if task.Goal != "New goal" {
		t.Errorf("goal should be 'New goal' after --force, got %q", task.Goal)
	}
}

func TestCTOReviewChangedFilesMismatch(t *testing.T) {
	runID, _ := setupTestRun(t, "task-mismatch", "ready_for_review",
		[]string{"pkg/a.go"},
		true, "# Handoff\ntest")

	deps := ctoDeps{
		changedFiles: func() ([]string, error) { return []string{"pkg/b.go"}, nil },
		readDiff:     func() (string, error) { return "", nil },
		runTest:      func() (string, error) { return "", fmt.Errorf("should not be called") },
		runVet:       func() (string, error) { return "", nil },
	}

	cmd, buf := makeReviewCmd()
	err := runCTOReviewWithDeps(cmd, "task-mismatch", runID, deps)
	if err != nil {
		t.Fatalf("review should not return error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "changes_requested") {
		t.Errorf("expected changes_requested for file mismatch, got: %s", out)
	}
	if !strings.Contains(out, "不一致") {
		t.Errorf("expected '不一致' finding, got: %s", out)
	}
}

func TestCTOReviewCurrentGitEmpty(t *testing.T) {
	runID, _ := setupTestRun(t, "task-empty-git", "ready_for_review",
		[]string{"pkg/a.go"},
		true, "# Handoff\ntest")

	deps := ctoDeps{
		changedFiles: func() ([]string, error) { return nil, nil },
		readDiff:     func() (string, error) { return "", nil },
		runTest:      func() (string, error) { return "", fmt.Errorf("should not be called") },
		runVet:       func() (string, error) { return "", nil },
	}

	cmd, buf := makeReviewCmd()
	err := runCTOReviewWithDeps(cmd, "task-empty-git", runID, deps)
	if err != nil {
		t.Fatalf("review should not return error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "blocked") && !strings.Contains(out, "没有") {
		t.Errorf("expected blocked for empty git, got: %s", out)
	}
}

func TestHandoffReSubmitSetsReadyForReview(t *testing.T) {
	chdirTempDir(t)

	// First submission via CLI.
	_, err := executeCommand("handoff", "submit", "task-rehandoff",
		"--agent", "backend_engineer",
		"--goal", "Test goal",
		"--summary", "Test summary",
	)
	if err != nil {
		t.Fatalf("first submit failed: %v", err)
	}

	// Directly set task status to changes_requested (simulating CTO review rejection).
	m := runs.NewManager()
	runsList, _ := m.ListRuns()
	if len(runsList) == 0 {
		t.Fatal("expected at least one run")
	}
	runID := runsList[0].RunID
	task, _ := m.GetTask(runID, "task-rehandoff")
	task.Status = "changes_requested"
	task.Handoff.ReadyForReview = false
	m.SaveTask(runID, task)

	// Re-submit the same task via CLI.
	_, err = executeCommand("handoff", "submit", "task-rehandoff",
		"--agent", "backend_engineer",
		"--goal", "Updated goal",
	)
	if err != nil {
		t.Fatalf("re-submit handoff failed: %v", err)
	}

	// Verify task status changed to ready_for_review regardless of create vs update path.
	updated, _ := m.GetTask(runID, "task-rehandoff")
	if updated.Status != "ready_for_review" {
		t.Errorf("status should be ready_for_review after re-handoff, got %q", updated.Status)
	}
	if !updated.Handoff.ReadyForReview {
		t.Error("handoff.ready_for_review should be true")
	}
}

func TestChangedFilesIncludesUntracked(t *testing.T) {
	chdirTempDir(t)

	initCmd := execCommand("git", "init")
	if err := initCmd.Run(); err != nil {
		t.Skipf("git init failed, skipping: %v", err)
	}
	execCommand("git", "config", "user.email", "test@test").Run()
	execCommand("git", "config", "user.name", "Test").Run()

	os.WriteFile("tracked.go", []byte("package main"), 0644)
	execCommand("git", "add", "tracked.go").Run()
	execCommand("git", "commit", "-m", "initial").Run()
	os.WriteFile("tracked.go", []byte("package main\n// changed"), 0644)
	os.WriteFile("untracked.txt", []byte("new file"), 0644)

	files, err := gitops.ChangedFiles(".")
	if err != nil {
		t.Fatalf("ChangedFiles: %v", err)
	}

	foundTracked := false
	foundUntracked := false
	for _, f := range files {
		if f == "tracked.go" {
			foundTracked = true
		}
		if f == "untracked.txt" {
			foundUntracked = true
		}
	}
	if !foundTracked {
		t.Error("ChangedFiles should include tracked changed file 'tracked.go'")
	}
	if !foundUntracked {
		t.Errorf("ChangedFiles should include untracked file 'untracked.txt', got: %v", files)
	}
}

var execCommand = func(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func TestRunsStatusShowsTasks(t *testing.T) {
	chdirTempDir(t)

	// Submit a task first.
	_, err := executeCommand("handoff", "submit", "test-task",
		"--agent", "backend_engineer",
		"--goal", "Test goal",
		"--summary", "Test summary",
	)
	if err != nil {
		t.Fatalf("handoff submit: %v", err)
	}

	// Find the run ID.
	m := runs.NewManager()
	runsList, _ := m.ListRuns()
	runID := runsList[0].RunID

	// Check status.
	out, err := executeCommand("runs", "status", runID)
	if err != nil {
		t.Fatalf("runs status: %v", err)
	}
	if !strings.Contains(out, "test-task") {
		t.Errorf("runs status should show task ID, got: %s", out)
	}
	if !strings.Contains(out, "backend_engineer") {
		t.Errorf("runs status should show agent, got: %s", out)
	}
}
