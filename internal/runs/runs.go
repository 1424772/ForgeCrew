// Package runs manages run/task state for ForgeCrew agent workflows.
// It provides a Manager for creating runs, tracking task status,
// writing handoff markdown files, and saving structured CTO reviews.
//
// All state is stored under .forgecrew/runs/<run_id>/ as YAML and markdown files.
package runs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/1424772/ForgeCrew/internal/config"
)

// ── Data types ──

// Run represents a ForgeCrew execution run grouping multiple tasks.
type Run struct {
	RunID     string `yaml:"run_id" json:"run_id"`
	Status    string `yaml:"status" json:"status"`
	CreatedAt string `yaml:"created_at" json:"created_at"`
	UpdatedAt string `yaml:"updated_at" json:"updated_at"`
}

// Task represents an agent's work item within a run.
type Task struct {
	TaskID       string        `yaml:"task_id" json:"task_id"`
	AgentID      string        `yaml:"agent_id" json:"agent_id"`
	Role         string        `yaml:"role" json:"role"`
	Status       string        `yaml:"status" json:"status"`
	StartedAt    string        `yaml:"started_at" json:"started_at"`
	CompletedAt  string        `yaml:"completed_at" json:"completed_at"`
	Goal         string        `yaml:"goal" json:"goal"`
	Summary      string        `yaml:"summary" json:"summary"`
	ChangedFiles []string      `yaml:"changed_files" json:"changed_files"`
	Verification Verification  `yaml:"verification" json:"verification"`
	Handoff      HandoffStatus `yaml:"handoff" json:"handoff"`
}

// Verification holds command verification results.
type Verification struct {
	Commands []VerificationResult `yaml:"commands" json:"commands"`
}

// VerificationResult is the outcome of running a verification command.
type VerificationResult struct {
	Command string `yaml:"command" json:"command"`
	Status  string `yaml:"status" json:"status"`
}

// HandoffStatus indicates whether a task is ready for review.
type HandoffStatus struct {
	ReadyForReview bool   `yaml:"ready_for_review" json:"ready_for_review"`
	Notes          string `yaml:"notes" json:"notes"`
}

// Review represents a CTO review of a task.
type Review struct {
	ReviewID     string               `yaml:"review_id" json:"review_id"`
	TaskID       string               `yaml:"task_id" json:"task_id"`
	Reviewer     string               `yaml:"reviewer" json:"reviewer"`
	Status       string               `yaml:"status" json:"status"`
	Findings     []Finding            `yaml:"findings" json:"findings"`
	Verification []VerificationResult `yaml:"verification" json:"verification"`
	Decision     Decision             `yaml:"decision" json:"decision"`
}

// Finding is a single issue discovered during review.
type Finding struct {
	Severity       string `yaml:"severity" json:"severity"`
	File           string `yaml:"file" json:"file"`
	Line           int    `yaml:"line" json:"line"`
	Issue          string `yaml:"issue" json:"issue"`
	Recommendation string `yaml:"recommendation" json:"recommendation"`
}

// Decision is the CTO's verdict on a task.
type Decision struct {
	MergeAllowed bool   `yaml:"merge_allowed" json:"merge_allowed"`
	NextAgent    string `yaml:"next_agent" json:"next_agent"`
}

// ── Manager ──

// Manager manages runs, tasks, handoffs, and reviews on disk.
type Manager struct {
	baseDir string
}

// NewManager creates a Manager rooted at .forgecrew/runs.
func NewManager() *Manager {
	return &Manager{baseDir: filepath.Join(config.ConfigDir, "runs")}
}

// NewManagerAt creates a Manager rooted at a custom directory (for testing).
func NewManagerAt(dir string) *Manager {
	return &Manager{baseDir: dir}
}

// BaseDir returns the manager's base directory.
func (m *Manager) BaseDir() string { return m.baseDir }

// ── Run operations ──

// CreateRun creates a new run with an auto-generated ID.
func (m *Manager) CreateRun() (*Run, error) {
	now := time.Now().UTC()
	runID := fmt.Sprintf("run-%s", now.Format("20060102-150405.000"))

	run := &Run{
		RunID:     runID,
		Status:    "in_progress",
		CreatedAt: now.Format(time.RFC3339Nano),
		UpdatedAt: now.Format(time.RFC3339Nano),
	}

	runDir := filepath.Join(m.baseDir, runID)
	if err := os.MkdirAll(runDir, 0755); err != nil {
		return nil, fmt.Errorf("create run dir: %w", err)
	}

	if err := m.saveRun(run); err != nil {
		return nil, err
	}

	// Ensure subdirectories exist.
	for _, sub := range []string{"tasks", "handoffs", "reviews"} {
		if err := os.MkdirAll(filepath.Join(runDir, sub), 0755); err != nil {
			return nil, fmt.Errorf("create %s dir: %w", sub, err)
		}
	}

	return run, nil
}

// GetRun reads a run by ID.
func (m *Manager) GetRun(runID string) (*Run, error) {
	path := filepath.Join(m.baseDir, runID, "run.yaml")
	if !config.FileExists(path) {
		return nil, fmt.Errorf("run not found: %s", runID)
	}
	var run Run
	if err := config.LoadYAML(path, &run); err != nil {
		return nil, err
	}
	return &run, nil
}

// ListRuns returns all runs sorted by creation time descending.
func (m *Manager) ListRuns() ([]Run, error) {
	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read runs dir: %w", err)
	}

	var runs []Run
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		runPath := filepath.Join(m.baseDir, e.Name(), "run.yaml")
		if !config.FileExists(runPath) {
			continue
		}
		var run Run
		if err := config.LoadYAML(runPath, &run); err != nil {
			continue
		}
		runs = append(runs, run)
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].CreatedAt > runs[j].CreatedAt
	})

	return runs, nil
}

// GetLatestRun returns the most recently created run, or nil if none exist.
func (m *Manager) GetLatestRun() (*Run, error) {
	runs, err := m.ListRuns()
	if err != nil {
		return nil, err
	}
	if len(runs) == 0 {
		return nil, nil
	}
	return &runs[0], nil
}

// saveRun writes a run's YAML file to disk.
func (m *Manager) saveRun(run *Run) error {
	path := filepath.Join(m.baseDir, run.RunID, "run.yaml")
	return config.SaveYAML(path, run)
}

// updateRunTimestamp updates the run's updated_at field.
func (m *Manager) updateRunTimestamp(runID string) error {
	run, err := m.GetRun(runID)
	if err != nil {
		return err
	}
	run.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return m.saveRun(run)
}

// ── Task operations ──

// SaveTask writes a task YAML file under the given run.
func (m *Manager) SaveTask(runID string, task *Task) error {
	tasksDir := filepath.Join(m.baseDir, runID, "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		return fmt.Errorf("create tasks dir: %w", err)
	}

	path := filepath.Join(tasksDir, task.TaskID+".yaml")
	if err := config.SaveYAML(path, task); err != nil {
		return err
	}

	return m.updateRunTimestamp(runID)
}

// GetTask reads a task YAML file from the given run.
func (m *Manager) GetTask(runID, taskID string) (*Task, error) {
	path := filepath.Join(m.baseDir, runID, "tasks", taskID+".yaml")
	if !config.FileExists(path) {
		return nil, fmt.Errorf("task not found: %s (run: %s)", taskID, runID)
	}
	var task Task
	if err := config.LoadYAML(path, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks returns all tasks in a run, sorted by started_at.
func (m *Manager) ListTasks(runID string) ([]Task, error) {
	tasksDir := filepath.Join(m.baseDir, runID, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read tasks dir: %w", err)
	}

	var tasks []Task
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(tasksDir, e.Name())
		var task Task
		if err := config.LoadYAML(path, &task); err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].StartedAt < tasks[j].StartedAt
	})

	return tasks, nil
}

// ── Handoff operations ──

// SaveHandoff writes a handoff markdown file for a task.
func (m *Manager) SaveHandoff(runID, taskID, content string) error {
	handoffDir := filepath.Join(m.baseDir, runID, "handoffs")
	if err := os.MkdirAll(handoffDir, 0755); err != nil {
		return fmt.Errorf("create handoffs dir: %w", err)
	}

	path := filepath.Join(handoffDir, taskID+".md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write handoff: %w", err)
	}

	return m.updateRunTimestamp(runID)
}

// ReadHandoff reads a handoff markdown file for a task.
func (m *Manager) ReadHandoff(runID, taskID string) (string, error) {
	path := filepath.Join(m.baseDir, runID, "handoffs", taskID+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("handoff not found for task %s (run: %s)", taskID, runID)
		}
		return "", fmt.Errorf("read handoff: %w", err)
	}
	return string(data), nil
}

// HasHandoff returns true if a handoff file exists for the task.
func (m *Manager) HasHandoff(runID, taskID string) bool {
	path := filepath.Join(m.baseDir, runID, "handoffs", taskID+".md")
	return config.FileExists(path)
}

// ── Review operations ──

// SaveReview writes a CTO review YAML file for a task.
func (m *Manager) SaveReview(runID, taskID string, review *Review) error {
	reviewDir := filepath.Join(m.baseDir, runID, "reviews")
	if err := os.MkdirAll(reviewDir, 0755); err != nil {
		return fmt.Errorf("create reviews dir: %w", err)
	}

	path := filepath.Join(reviewDir, taskID+"-cto.yaml")
	if err := config.SaveYAML(path, review); err != nil {
		return err
	}

	return m.updateRunTimestamp(runID)
}
