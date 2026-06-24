// Package checkpoint provides checkpoint creation, listing, and rollback.
// All Agent modifications should be checkpointed for audit and rollback.
package checkpoint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/gitops"
)

// Checkpoint represents a saved state before/after agent modifications.
type Checkpoint struct {
	ID           string   `json:"id"`
	Timestamp    string   `json:"timestamp"`
	TaskID       string   `json:"task_id"`
	AgentID      string   `json:"agent_id"`
	ModelID      string   `json:"model_id,omitempty"`
	ChangedFiles []string `json:"changed_files"`
	GitHashBefore string  `json:"git_hash_before"`
}

// Store manages checkpoints on disk.
type Store struct {
	dir string
}

// NewStore creates a new checkpoint store using the default .forgecrew/checkpoints directory.
func NewStore() *Store {
	return &Store{dir: config.ConfigDir + "/" + config.CheckpointsDir}
}

// NewStoreAt creates a checkpoint store at a specific path.
func NewStoreAt(dir string) *Store {
	return &Store{dir: dir}
}

// Create writes a new checkpoint JSON file.
func (s *Store) Create(taskID, agentID string, changedFiles []string) (*Checkpoint, error) {
	if err := config.EnsureDir(s.dir); err != nil {
		return nil, err
	}

	hash, _ := gitops.GetCurrentHash(".")
	id := fmt.Sprintf("ckpt_%d", time.Now().UnixNano())

	ckpt := &Checkpoint{
		ID:            id,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		TaskID:        taskID,
		AgentID:       agentID,
		ChangedFiles:  changedFiles,
		GitHashBefore: hash,
	}
	if ckpt.ChangedFiles == nil {
		ckpt.ChangedFiles = []string{}
	}

	filename := filepath.Join(s.dir, id+".json")
	data, err := json.MarshalIndent(ckpt, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal checkpoint: %w", err)
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return nil, fmt.Errorf("write checkpoint: %w", err)
	}
	return ckpt, nil
}

// List returns all checkpoints sorted by timestamp (newest first).
func (s *Store) List() ([]Checkpoint, error) {
	if err := config.EnsureDir(s.dir); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read checkpoints dir: %w", err)
	}

	var ckpts []Checkpoint
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			continue
		}
		var c Checkpoint
		if err := json.Unmarshal(data, &c); err != nil {
			continue
		}
		ckpts = append(ckpts, c)
	}

	sort.Slice(ckpts, func(i, j int) bool {
		return ckpts[i].Timestamp > ckpts[j].Timestamp
	})
	return ckpts, nil
}
