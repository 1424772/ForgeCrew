package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/1424772/ForgeCrew/internal/gitops"
	"github.com/1424772/ForgeCrew/internal/runs"
	"github.com/spf13/cobra"
)

var (
	handoffRun     string
	handoffAgent   string
	handoffRole    string
	handoffGoal    string
	handoffSummary string
	handoffNotes   string
	handoffForce   bool
)

func handoffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handoff",
		Short: "Manage task handoffs",
		Long:  "Submit and view task handoffs.",
	}
	cmd.AddCommand(handoffSubmitCmd())
	return cmd
}

func handoffSubmitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit <task_id>",
		Short: "Submit a task handoff for review",
		Long: `Create or update a task handoff, readying it for CTO review.

Automatically collects changed files from git diff and creates
a handoff markdown file with the task goal, summary, and notes.

If no run exists, a new run is created automatically.`,
		Args: cobra.ExactArgs(1),
		RunE: runHandoffSubmit,
	}
	cmd.Flags().StringVar(&handoffRun, "run", "", "Run ID (default: latest run)")
	cmd.Flags().StringVar(&handoffAgent, "agent", "backend_engineer", "Agent ID")
	cmd.Flags().StringVar(&handoffRole, "role", "coding", "Agent role")
	cmd.Flags().StringVar(&handoffGoal, "goal", "", "Task goal")
	cmd.Flags().StringVar(&handoffSummary, "summary", "", "Task summary")
	cmd.Flags().StringVar(&handoffNotes, "notes", "", "Handoff notes")
	cmd.Flags().BoolVar(&handoffForce, "force", false, "Overwrite existing handoff content")
	return cmd
}

func runHandoffSubmit(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	m := runs.NewManager()

	// Determine run.
	var run *runs.Run
	var err error
	if handoffRun != "" {
		run, err = m.GetRun(handoffRun)
		if err != nil {
			return err
		}
	} else {
		run, err = m.GetLatestRun()
		if err != nil {
			return err
		}
		if run == nil {
			run, err = m.CreateRun()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "创建新 run: %s\n", run.RunID)
		}
	}

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	// Check if task already exists.
	existingTask, taskExists := loadExistingTask(m, run.RunID, taskID)

	var changedFiles []string
	if gitops.IsGitRepo(".") {
		changedFiles, _ = gitops.ChangedFiles(".")
	}

	if taskExists && !handoffForce {
		// Update existing task: set ready_for_review, refresh changed_files.
		// Preserve handoff markdown content unless --force is given.
		existingTask.Status = "ready_for_review"
		existingTask.ChangedFiles = changedFiles
		existingTask.CompletedAt = nowStr
		existingTask.Handoff.ReadyForReview = true
		existingTask.Verification = runs.Verification{
			Commands: buildVerification(),
		}
		if handoffNotes != "" {
			existingTask.Handoff.Notes = handoffNotes
		}

		if err := m.SaveTask(run.RunID, existingTask); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "更新 task: %s (保留已有 handoff)\n", taskID)
		return nil
	}

	// Create new task.
	task := &runs.Task{
		TaskID:       taskID,
		AgentID:      handoffAgent,
		Role:         handoffRole,
		Status:       "ready_for_review",
		StartedAt:    nowStr,
		CompletedAt:  nowStr,
		Goal:         handoffGoal,
		Summary:      handoffSummary,
		ChangedFiles: changedFiles,
		Verification: runs.Verification{
			Commands: buildVerification(),
		},
		Handoff: runs.HandoffStatus{
			ReadyForReview: true,
			Notes:          handoffNotes,
		},
	}

	if err := m.SaveTask(run.RunID, task); err != nil {
		return err
	}

	// Write handoff markdown.
	handoffMD := buildHandoffMD(taskID, handoffAgent, handoffGoal, handoffSummary, handoffNotes, changedFiles)
	if err := m.SaveHandoff(run.RunID, taskID, handoffMD); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "提交 handoff: %s (run: %s)\n", taskID, run.RunID)
	fmt.Fprintf(cmd.OutOrStdout(), "  Agent: %s\n", handoffAgent)
	if len(changedFiles) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Changed files: %s\n", strings.Join(changedFiles, ", "))
	}
	return nil
}

func loadExistingTask(m *runs.Manager, runID, taskID string) (*runs.Task, bool) {
	task, err := m.GetTask(runID, taskID)
	if err != nil {
		return nil, false
	}
	return task, true
}

func buildVerification() []runs.VerificationResult {
	return []runs.VerificationResult{
		{Command: "go test ./...", Status: "pending"},
		{Command: "go vet ./...", Status: "pending"},
	}
}

func buildHandoffMD(taskID, agent, goal, summary, notes string, changedFiles []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Handoff: %s\n\n", taskID))
	b.WriteString(fmt.Sprintf("**Agent:** %s\n\n", agent))
	if goal != "" {
		b.WriteString(fmt.Sprintf("## Goal\n%s\n\n", goal))
	}
	if summary != "" {
		b.WriteString(fmt.Sprintf("## Summary\n%s\n\n", summary))
	}
	if notes != "" {
		b.WriteString(fmt.Sprintf("## Notes\n%s\n\n", notes))
	}
	if len(changedFiles) > 0 {
		b.WriteString("## Changed Files\n")
		for _, f := range changedFiles {
			b.WriteString(fmt.Sprintf("- %s\n", f))
		}
		b.WriteString("\n")
	}
	return b.String()
}
