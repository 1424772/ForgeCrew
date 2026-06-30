package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/1424772/ForgeCrew/internal/config"
	"github.com/1424772/ForgeCrew/internal/i18n"
	"github.com/1424772/ForgeCrew/internal/models"
	"github.com/1424772/ForgeCrew/internal/orchestrator"
	"github.com/1424772/ForgeCrew/internal/provider"
	"github.com/spf13/cobra"
)

var taskDryRun bool
var taskExecute bool
var taskModel string

func taskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task <goal>",
		Short: "Execute a task through the Loop Engineering state machine",
		Long: `Execute a task through the full Loop Engineering cycle.

Without --execute, runs in dry-run mode and prints the state sequence
without making any changes.

With --execute, calls the configured LLM provider to generate an
execution plan. No files are written, no shell commands are executed.
Use --model to specify which model to use (required in execute mode).

The Loop Engineering cycle runs: Goal -> Plan -> Retrieve -> Act ->
Observe -> Reflect -> Improve -> Review -> CommitMemory`,
		Args: cobra.ExactArgs(1),
		RunE: runTask,
	}
	cmd.Flags().BoolVar(&taskDryRun, "dry-run", false, "Run in dry-run mode (no real execution)")
	cmd.Flags().BoolVar(&taskExecute, "execute", false, "Execute with real LLM provider (generates plan, no file writes)")
	cmd.Flags().StringVar(&taskModel, "model", "", "Model ID to use in execute mode (required with --execute)")
	return cmd
}

func runTask(cmd *cobra.Command, args []string) error {
	goal := args[0]

	if taskExecute {
		return runTaskExecute(cmd, goal)
	}

	// Dry-run path: use orchestrator to print state sequence.
	sm := orchestrator.New(3, true)
	result := sm.RunFull(goal)

	locale, err := loadLocale()
	if err != nil {
		return err
	}

	fmt.Fprint(cmd.OutOrStdout(), result.FormatText(locale))
	return nil
}

// runTaskExecute calls the LLM provider to generate a plan.
// Safety: no file writes, no shell execution — only text generation.
func runTaskExecute(cmd *cobra.Command, goal string) error {
	locale, err := loadLocale()
	if err != nil {
		return err
	}
	loc := i18n.Locale(locale)

	// Load models registry.
	registry := models.NewRegistry()
	modelsPath := filepath.Join(config.ConfigDir, config.ModelsYAML)
	if regErr := registry.Load(modelsPath); regErr != nil {
		return fmt.Errorf("%s", i18n.T("task.execute_missing_config", loc))
	}

	if taskModel == "" {
		return fmt.Errorf("--execute 需要指定 --model <model_id>。可用模型: %s", modelIDsString(registry))
	}

	def, getErr := registry.Get(taskModel)
	if getErr != nil {
		return fmt.Errorf("模型 %q 未在 models.yaml 中找到。可用模型: %s", taskModel, modelIDsString(registry))
	}

	// Create provider client.
	client, clientErr := provider.NewClient(*def)
	if clientErr != nil {
		return fmt.Errorf("%s: %w", i18n.T("task.execute_provider_error", loc), clientErr)
	}

	req := provider.Request{
		System: "You are an AI software engineer. Generate a detailed implementation plan for the following task. Do NOT write any code or modify files — only produce a written plan with steps, considerations, and estimated effort.",
		Prompt: goal,
		Model:  def.Model,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, respErr := client.Complete(ctx, req)
	if respErr != nil {
		return fmt.Errorf("%s: %w", i18n.T("task.execute_provider_error", loc), respErr)
	}

	// Output.
	dryTag := i18n.T("task.execute_mode", loc)
	fmt.Fprintf(cmd.OutOrStdout(), "%s%q%s\n", i18n.T("task.header", loc), goal, dryTag)
	fmt.Fprint(cmd.OutOrStdout(), i18n.T("task.execute_plan_header", loc))
	fmt.Fprintln(cmd.OutOrStdout(), resp.Text)

	return nil
}

// modelIDsString returns a comma-separated list of model IDs from the registry.
func modelIDsString(r *models.Registry) string {
	ids := r.ModelIDs()
	if len(ids) == 0 {
		return "(无)"
	}
	var parts []string
	for id := range ids {
		parts = append(parts, id)
	}
	return strings.Join(parts, ", ")
}
