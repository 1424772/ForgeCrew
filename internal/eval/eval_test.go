package eval

import (
	"testing"
)

func TestNewStore(t *testing.T) {
	s := NewStore()
	if s == nil {
		t.Fatal("NewStore returned nil")
	}
}

func TestWriteAndList(t *testing.T) {
	tmp := t.TempDir()
	s := &Store{dir: tmp}

	record := &Record{
		TaskID:                "task_001",
		Success:               true,
		TestPassed:            true,
		ReviewPassed:          true,
		Rounds:                3,
		CostEstimate:          0.15,
		HumanInterventionCount: 1,
	}
	if err := s.Write(record); err != nil {
		t.Fatal(err)
	}

	records, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if !records[0].Success {
		t.Error("success should be true")
	}
	if records[0].Rounds != 3 {
		t.Errorf("rounds = %d, want 3", records[0].Rounds)
	}
	if records[0].CostEstimate != 0.15 {
		t.Errorf("cost_estimate = %f, want 0.15", records[0].CostEstimate)
	}
}

func TestListEmpty(t *testing.T) {
	tmp := t.TempDir()
	s := &Store{dir: tmp}

	records, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestRecordFields(t *testing.T) {
	tmp := t.TempDir()
	s := &Store{dir: tmp}

	record := &Record{
		TaskID:                "task_002",
		Success:               false,
		TestPassed:            false,
		ReviewPassed:          false,
		Rounds:                5,
		CostEstimate:          0.50,
		HumanInterventionCount: 3,
	}
	if err := s.Write(record); err != nil {
		t.Fatal(err)
	}
	if record.ID == "" {
		t.Error("ID should be auto-generated")
	}
	if record.Timestamp == "" {
		t.Error("Timestamp should be auto-generated")
	}
}
