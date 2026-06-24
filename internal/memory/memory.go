// Package memory provides task and project memory persistence using JSONL.
// MVP uses flat JSONL files; no SQLite dependency yet.
package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/1424772/ForgeCrew/internal/config"
)

// EntryType classifies a memory record.
type EntryType string

const (
	TypeTask    EntryType = "task"
	TypeProject EntryType = "project"
)

// Entry is a single memory record.
type Entry struct {
	ID        string    `json:"id"`
	Type      EntryType `json:"type"`
	Scope     string    `json:"scope"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags,omitempty"`
	Timestamp string    `json:"timestamp"`
}

// Store manages memory persistence.
type Store struct {
	dir string
}

// NewStore creates a new memory store using the default .forgecrew/memory directory.
func NewStore() *Store {
	return &Store{dir: config.ConfigDir + "/" + config.MemoryDir}
}

// Write appends a memory entry to the appropriate JSONL file.
func (s *Store) Write(entry *Entry) error {
	if err := config.EnsureDir(s.dir); err != nil {
		return err
	}
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("mem_%d", time.Now().UnixNano())
	}
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	filename := filepath.Join(s.dir, string(entry.Type)+".jsonl")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open memory file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal memory entry: %w", err)
	}
	data = append(data, '\n')
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write memory entry: %w", err)
	}
	return nil
}

// ReadAll reads all memory entries of a given type.
func (s *Store) ReadAll(t EntryType) ([]Entry, error) {
	filename := filepath.Join(s.dir, string(t)+".jsonl")
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read memory file: %w", err)
	}

	var entries []Entry
	lines := splitLines(string(data))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, ch := range s {
		if ch == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
