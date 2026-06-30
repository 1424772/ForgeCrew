package main

import (
	"fmt"

	"github.com/1424772/ForgeCrew/internal/runs"
	"github.com/spf13/cobra"
)

func runsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runs",
		Short: "Manage execution runs",
		Long:  "List runs and view run status.",
	}
	cmd.AddCommand(runsListCmd())
	cmd.AddCommand(runsStatusCmd())
	return cmd
}

func runsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all runs",
		RunE:  runRunsList,
	}
}

func runRunsList(cmd *cobra.Command, args []string) error {
	m := runs.NewManager()
	list, err := m.ListRuns()
	if err != nil {
		return err
	}

	if len(list) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "没有 run。使用 'forgecrew handoff submit <task_id>' 创建第一个 run。")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%-36s %-16s %-20s %-20s\n", "RUN ID", "STATUS", "CREATED", "UPDATED")
	for _, r := range list {
		fmt.Fprintf(cmd.OutOrStdout(), "%-36s %-16s %-20s %-20s\n",
			r.RunID, r.Status, trimTS(r.CreatedAt), trimTS(r.UpdatedAt))
	}
	return nil
}

func runsStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <run_id>",
		Short: "Show run status and task list",
		Args:  cobra.ExactArgs(1),
		RunE:  runRunsStatus,
	}
}

func runRunsStatus(cmd *cobra.Command, args []string) error {
	runID := args[0]
	m := runs.NewManager()

	run, err := m.GetRun(runID)
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Run: %s\n", run.RunID)
	fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", run.Status)
	fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", run.CreatedAt)
	fmt.Fprintf(cmd.OutOrStdout(), "Updated: %s\n\n", run.UpdatedAt)

	tasks, err := m.ListTasks(runID)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "没有 task。")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-20s %-20s %s\n", "TASK ID", "AGENT", "STATUS", "READY FOR REVIEW")
	for _, t := range tasks {
		ready := "no"
		if t.Handoff.ReadyForReview {
			ready = "yes"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-20s %-20s %s\n",
			t.TaskID, t.AgentID, t.Status, ready)
	}
	return nil
}

// trimTS returns the first 19 characters of an RFC3339 timestamp (YYYY-MM-DDTHH:MM:SS).
func trimTS(ts string) string {
	if len(ts) >= 19 {
		return ts[:19]
	}
	return ts
}
