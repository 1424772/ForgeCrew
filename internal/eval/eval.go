// Package eval provides the evaluation harness for measuring
// agent, team, and workflow performance.
package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/1424772/ForgeCrew/internal/config"
)

// Record captures the result of a single task evaluation.
type Record struct {
	ID                     string  `json:"id"`
	TaskID                 string  `json:"task_id"`
	Success                bool    `json:"success"`
	TestPassed             bool    `json:"test_passed"`
	ReviewPassed           bool    `json:"review_passed"`
	Rounds                 int     `json:"rounds"`
	CostEstimate           float64 `json:"cost_estimate"`
	HumanInterventionCount int     `json:"human_intervention_count"`
	Timestamp              string  `json:"timestamp"`
}

// Store manages evaluation records.
type Store struct {
	dir string
}

// NewStore creates a new eval store.
func NewStore() *Store {
	return &Store{dir: config.ConfigDir + "/" + config.EvalsDir}
}

// Write saves an evaluation record as a JSON file.
func (s *Store) Write(record *Record) error {
	if err := config.EnsureDir(s.dir); err != nil {
		return err
	}
	if record.ID == "" {
		record.ID = fmt.Sprintf("eval_%d", time.Now().UnixNano())
	}
	if record.Timestamp == "" {
		record.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	filename := filepath.Join(s.dir, record.ID+".json")
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal eval record: %w", err)
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("write eval record: %w", err)
	}
	return nil
}

// List returns all eval records.
func (s *Store) List() ([]Record, error) {
	if err := config.EnsureDir(s.dir); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read evals dir: %w", err)
	}
	var records []Record
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			continue
		}
		var r Record
		if err := json.Unmarshal(data, &r); err != nil {
			continue
		}
		records = append(records, r)
	}
	return records, nil
}
