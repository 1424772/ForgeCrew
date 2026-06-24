package memory

import (
	"testing"
)

func TestNewStore(t *testing.T) {
	s := NewStore()
	if s == nil {
		t.Fatal("NewStore returned nil")
	}
}

func TestWriteAndReadTask(t *testing.T) {
	tmp := t.TempDir()
	s := &Store{dir: tmp}

	entry := &Entry{
		ID:      "mem_001",
		Type:    TypeTask,
		Scope:   "task_001",
		Content: "Fixed login bug by adding rate limiting",
		Tags:    []string{"bug", "auth"},
	}
	if err := s.Write(entry); err != nil {
		t.Fatal(err)
	}

	entries, err := s.ReadAll(TypeTask)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Content != "Fixed login bug by adding rate limiting" {
		t.Errorf("content mismatch: %q", entries[0].Content)
	}
	if len(entries[0].Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(entries[0].Tags))
	}
}

func TestWriteProjectMemory(t *testing.T) {
	tmp := t.TempDir()
	s := &Store{dir: tmp}

	entry := &Entry{
		Type:    TypeProject,
		Scope:   "project",
		Content: "This project uses Go + Cobra for CLI",
	}
	if err := s.Write(entry); err != nil {
		t.Fatal(err)
	}

	entries, err := s.ReadAll(TypeProject)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Type != TypeProject {
		t.Errorf("type = %q, want project", entries[0].Type)
	}
}

func TestReadEmpty(t *testing.T) {
	tmp := t.TempDir()
	s := &Store{dir: tmp}

	entries, err := s.ReadAll(TypeTask)
	if err != nil {
		t.Fatal(err)
	}
	if entries != nil && len(entries) > 0 {
		t.Error("expected nil or empty for non-existent file")
	}
}

func TestAutoGenerateID(t *testing.T) {
	tmp := t.TempDir()
	s := &Store{dir: tmp}

	entry := &Entry{
		Type:    TypeTask,
		Content: "test",
	}
	if err := s.Write(entry); err != nil {
		t.Fatal(err)
	}
	if entry.ID == "" {
		t.Error("ID should be auto-generated")
	}
}
