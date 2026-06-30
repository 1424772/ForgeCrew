package runs

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestCreateRun(t *testing.T) {
	m := NewManagerAt(t.TempDir())
	run, err := m.CreateRun()
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if run.RunID == "" {
		t.Error("run ID should not be empty")
	}
	if !strings.HasPrefix(run.RunID, "run-") {
		t.Errorf("run ID should start with 'run-', got %q", run.RunID)
	}
	if run.Status != "in_progress" {
		t.Errorf("status = %q, want in_progress", run.Status)
	}
	if run.CreatedAt == "" || run.UpdatedAt == "" {
		t.Error("timestamps should not be empty")
	}
}

func TestGetRun(t *testing.T) {
	m := NewManagerAt(t.TempDir())
	run, _ := m.CreateRun()

	got, err := m.GetRun(run.RunID)
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if got.RunID != run.RunID {
		t.Errorf("RunID mismatch: got %q, want %q", got.RunID, run.RunID)
	}
}

func TestGetRunNotFound(t *testing.T) {
	m := NewManagerAt(t.TempDir())
	_, err := m.GetRun("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent run")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestListRuns(t *testing.T) {
	m := NewManagerAt(t.TempDir())

	// No runs initially.
	runs, err := m.ListRuns()
	if err != nil {
		t.Fatalf("ListRuns (empty): %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}

	// Create two runs with a small delay to ensure unique IDs and ordering.
	r1, _ := m.CreateRun()
	time.Sleep(50 * time.Millisecond)
	r2, _ := m.CreateRun()

	runs, err = m.ListRuns()
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}
	// Verify both expected runs are present.
	seen := map[string]bool{runs[0].RunID: true, runs[1].RunID: true}
	if !seen[r1.RunID] {
		t.Errorf("runs list missing %q", r1.RunID)
	}
	if !seen[r2.RunID] {
		t.Errorf("runs list missing %q", r2.RunID)
	}
}

func TestGetLatestRun(t *testing.T) {
	m := NewManagerAt(t.TempDir())

	// No runs.
	latest, err := m.GetLatestRun()
	if err != nil {
		t.Fatalf("GetLatestRun (empty): %v", err)
	}
	if latest != nil {
		t.Error("expected nil when no runs exist")
	}

	r1, _ := m.CreateRun()
	time.Sleep(50 * time.Millisecond)
	r2, _ := m.CreateRun()

	latest, err = m.GetLatestRun()
	if err != nil {
		t.Fatalf("GetLatestRun: %v", err)
	}
	if latest == nil {
		t.Fatal("expected non-nil run")
	}
	if latest.RunID != r2.RunID {
		t.Errorf("latest should be %q, got %q", r2.RunID, latest.RunID)
	}
	_ = r1
}

func TestSaveAndGetTask(t *testing.T) {
	m := NewManagerAt(t.TempDir())
	run, _ := m.CreateRun()

	task := &Task{
		TaskID:       "backend-001",
		AgentID:      "backend_engineer",
		Role:         "coding",
		Status:       "ready_for_review",
		StartedAt:    time.Now().UTC().Format(time.RFC3339),
		Goal:         "Implement feature X",
		Summary:      "Added new package",
		ChangedFiles: []string{"pkg/x.go", "pkg/x_test.go"},
		Handoff: HandoffStatus{
			ReadyForReview: true,
			Notes:          "All tests pass",
		},
	}

	if err := m.SaveTask(run.RunID, task); err != nil {
		t.Fatalf("SaveTask: %v", err)
	}

	got, err := m.GetTask(run.RunID, "backend-001")
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if got.TaskID != "backend-001" {
		t.Errorf("TaskID = %q", got.TaskID)
	}
	if got.AgentID != "backend_engineer" {
		t.Errorf("AgentID = %q", got.AgentID)
	}
	if got.Status != "ready_for_review" {
		t.Errorf("Status = %q", got.Status)
	}
	if len(got.ChangedFiles) != 2 {
		t.Errorf("expected 2 changed files, got %d", len(got.ChangedFiles))
	}
	if !got.Handoff.ReadyForReview {
		t.Error("Handoff.ReadyForReview should be true")
	}
}

func TestGetTaskNotFound(t *testing.T) {
	m := NewManagerAt(t.TempDir())
	run, _ := m.CreateRun()

	_, err := m.GetTask(run.RunID, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent task")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestListTasks(t *testing.T) {
	m := NewManagerAt(t.TempDir())
	run, _ := m.CreateRun()

	m.SaveTask(run.RunID, &Task{
		TaskID: "task-1", AgentID: "agent-a", Status: "in_progress",
		StartedAt: "2026-01-01T00:00:00Z",
	})
	m.SaveTask(run.RunID, &Task{
		TaskID: "task-2", AgentID: "agent-b", Status: "ready_for_review",
		StartedAt: "2026-01-02T00:00:00Z",
	})

	tasks, err := m.ListTasks(run.RunID)
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].TaskID != "task-1" {
		t.Errorf("first task (sorted by started_at) should be task-1, got %s", tasks[0].TaskID)
	}
}

func TestSaveAndReadHandoff(t *testing.T) {
	m := NewManagerAt(t.TempDir())
	run, _ := m.CreateRun()

	content := `# Handoff: backend-001

## Goal
Implement feature X

## Summary
Added new package with tests.

## Changed Files
- pkg/x.go
- pkg/x_test.go
`
	if err := m.SaveHandoff(run.RunID, "backend-001", content); err != nil {
		t.Fatalf("SaveHandoff: %v", err)
	}

	got, err := m.ReadHandoff(run.RunID, "backend-001")
	if err != nil {
		t.Fatalf("ReadHandoff: %v", err)
	}
	if got != content {
		t.Errorf("handoff content mismatch\ngot:\n%s\nwant:\n%s", got, content)
	}
	if !m.HasHandoff(run.RunID, "backend-001") {
		t.Error("HasHandoff should return true")
	}
}

func TestReadHandoffNotFound(t *testing.T) {
	m := NewManagerAt(t.TempDir())
	run, _ := m.CreateRun()

	_, err := m.ReadHandoff(run.RunID, "missing")
	if err == nil {
		t.Fatal("expected error for missing handoff")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestHasHandoffFalse(t *testing.T) {
	m := NewManagerAt(t.TempDir())
	run, _ := m.CreateRun()

	if m.HasHandoff(run.RunID, "missing") {
		t.Error("HasHandoff should return false for missing handoff")
	}
}

func TestSaveReview(t *testing.T) {
	m := NewManagerAt(t.TempDir())
	run, _ := m.CreateRun()

	review := &Review{
		ReviewID: "cto-review-backend-001",
		TaskID:   "backend-001",
		Reviewer: "cto",
		Status:   "approved",
		Findings: []Finding{
			{
				Severity:       "info",
				File:           "pkg/x.go",
				Line:           42,
				Issue:          "Consider adding doc comment",
				Recommendation: "Add a doc comment for the exported function",
			},
		},
		Verification: []VerificationResult{
			{Command: "go test ./...", Status: "passed"},
			{Command: "go vet ./...", Status: "passed"},
		},
		Decision: Decision{
			MergeAllowed: true,
			NextAgent:    "",
		},
	}

	if err := m.SaveReview(run.RunID, "backend-001", review); err != nil {
		t.Fatalf("SaveReview: %v", err)
	}

	// Verify file exists on disk.
	path := m.baseDir + "/" + run.RunID + "/reviews/backend-001-cto.yaml"
	fi, err := os.Stat(path)
	if err != nil {
		t.Errorf("review file not found at %s: %v", path, err)
	} else if fi.IsDir() {
		t.Errorf("expected file, got directory at %s", path)
	}
}

func TestUpdateRunTimestamp(t *testing.T) {
	m := NewManagerAt(t.TempDir())
	run, _ := m.CreateRun()

	original := run.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	task := &Task{TaskID: "t1", AgentID: "a", Status: "ready_for_review"}
	m.SaveTask(run.RunID, task)

	updated, _ := m.GetRun(run.RunID)
	if updated.UpdatedAt == original {
		t.Error("updated_at should change after SaveTask")
	}
}
