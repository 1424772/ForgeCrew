package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStore(t *testing.T) {
	s := NewStore()
	if s == nil {
		t.Fatal("NewStore returned nil")
	}
}

func TestCreateAndList(t *testing.T) {
	tmp := t.TempDir()
	s := NewStoreAt(tmp)

	// Create a checkpoint.
	ckpt, err := s.Create("task_001", "backend_engineer", []string{"main.go", "config.go"})
	if err != nil {
		t.Fatal(err)
	}
	if ckpt.ID == "" {
		t.Error("checkpoint ID should not be empty")
	}
	if ckpt.TaskID != "task_001" {
		t.Errorf("TaskID = %q, want task_001", ckpt.TaskID)
	}
	if ckpt.AgentID != "backend_engineer" {
		t.Errorf("AgentID = %q, want backend_engineer", ckpt.AgentID)
	}
	if len(ckpt.ChangedFiles) != 2 {
		t.Errorf("expected 2 changed files, got %d", len(ckpt.ChangedFiles))
	}

	// Verify the JSON file exists.
	files, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 checkpoint file, got %d", len(files))
	}

	// List.
	ckpts, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(ckpts) != 1 {
		t.Fatalf("expected 1 checkpoint in list, got %d", len(ckpts))
	}
	if ckpts[0].ID != ckpt.ID {
		t.Errorf("listed ID = %q, want %q", ckpts[0].ID, ckpt.ID)
	}
}

func TestListEmpty(t *testing.T) {
	tmp := t.TempDir()
	s := NewStoreAt(tmp)

	ckpts, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(ckpts) != 0 {
		t.Errorf("expected 0 checkpoints, got %d", len(ckpts))
	}
}

func TestCreateCheckpointFileFormat(t *testing.T) {
	tmp := t.TempDir()
	s := NewStoreAt(tmp)

	ckpt, err := s.Create("task_002", "qa_engineer", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Read the file back and verify JSON structure.
	data, err := os.ReadFile(filepath.Join(tmp, ckpt.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("checkpoint file should not be empty")
	}
	// Verify key fields are present.
	content := string(data)
	for _, field := range []string{"id", "timestamp", "task_id", "agent_id", "changed_files"} {
		if !contains(content, field) {
			t.Errorf("checkpoint JSON missing field: %s", field)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
