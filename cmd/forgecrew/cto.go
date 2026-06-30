package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/1424772/ForgeCrew/internal/gitops"
	"github.com/1424772/ForgeCrew/internal/runs"
	"github.com/spf13/cobra"
)

var ctoRun string

// ctoDeps holds injectable dependencies for CTO review, enabling testing
// without real git or go commands.
type ctoDeps struct {
	changedFiles func() ([]string, error)
	readDiff     func() (string, error)
	runTest      func() (string, error)
	runVet       func() (string, error)
}

func ctoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cto",
		Short: "CTO review operations",
		Long:  "Review agent task results as CTO (rule-based audit, no LLM).",
	}
	cmd.AddCommand(ctoReviewCmd())
	return cmd
}

func ctoReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review <task_id>",
		Short: "Review a task as CTO",
		Long: `Run a rule-based CTO review of a task.

The review checks:
  - Task status is ready_for_review
  - Handoff file exists
  - Changed files are present
  - go test ./... passes
  - go vet ./... passes

The review is rule-based (no LLM). Results are saved to
.forgecrew/runs/<run_id>/reviews/<task_id>-cto.yaml.`,
		Args: cobra.ExactArgs(1),
		RunE: runCTOReview,
	}
	cmd.Flags().StringVar(&ctoRun, "run", "", "Run ID (default: latest run)")
	return cmd
}

func runCTOReview(cmd *cobra.Command, args []string) error {
	deps := ctoDeps{
		changedFiles: func() ([]string, error) { return gitops.ChangedFiles(".") },
		readDiff:     func() (string, error) { return gitops.ReadDiffFull(".") },
		runTest:      func() (string, error) { return runGoCmd("test", "./...") },
		runVet:       func() (string, error) { return runGoCmd("vet", "./...") },
	}
	return runCTOReviewWithDeps(cmd, args[0], ctoRun, deps)
}

// runCTOReviewWithDeps is the testable core of CTO review.
func runCTOReviewWithDeps(cmd *cobra.Command, taskID, runID string, deps ctoDeps) error {
	m := runs.NewManager()

	// Determine run.
	var run *runs.Run
	var err error
	if runID != "" {
		run, err = m.GetRun(runID)
		if err != nil {
			return err
		}
	} else {
		run, err = m.GetLatestRun()
		if err != nil {
			return err
		}
		if run == nil {
			return fmt.Errorf("没有找到 run。请先运行 'forgecrew handoff submit %s'", taskID)
		}
	}
	runID = run.RunID

	// Load task.
	task, err := m.GetTask(runID, taskID)
	if err != nil {
		return err
	}

	// Must be ready_for_review.
	if task.Status != "ready_for_review" {
		return fmt.Errorf("task %s 状态为 %q，不能 review（需要 ready_for_review）", taskID, task.Status)
	}

	// Build review.
	review := &runs.Review{
		ReviewID: fmt.Sprintf("cto-review-%s", taskID),
		TaskID:   taskID,
		Reviewer: "cto",
	}

	var findings []runs.Finding
	var allPassed bool

	// Check 1: Has handoff.
	if !m.HasHandoff(runID, taskID) {
		findings = append(findings, runs.Finding{
			Severity:       "blocking",
			Issue:          "缺少 handoff 文件",
			Recommendation: fmt.Sprintf("运行 'forgecrew handoff submit %s' 创建 handoff", taskID),
		})
		review.Status = "blocked"
		review.Decision.MergeAllowed = false
	} else {
		allPassed = true
	}

	// Check 2: Current git changed files (real, not just task YAML history).
	currentFiles, currentFilesErr := deps.changedFiles()
	if currentFilesErr != nil {
		findings = append(findings, runs.Finding{
			Severity:       "blocking",
			Issue:          fmt.Sprintf("无法读取当前 git 变更: %s", currentFilesErr.Error()),
			Recommendation: "确认当前目录是 git 仓库",
		})
		review.Status = "blocked"
		review.Decision.MergeAllowed = false
		allPassed = false
	} else if len(currentFiles) == 0 {
		findings = append(findings, runs.Finding{
			Severity:       "blocking",
			Issue:          "当前没有 git 变更（git diff 为空，无未跟踪文件）",
			Recommendation: "确保在提交 handoff 前有实际代码变更",
		})
		review.Status = "blocked"
		review.Decision.MergeAllowed = false
		allPassed = false
	} else {
		// Check 2a: task YAML changed_files must not be stale / empty.
		if len(task.ChangedFiles) == 0 {
			findings = append(findings, runs.Finding{
				Severity:       "blocking",
				Issue:          "task.ChangedFiles 为空，但当前有 git 变更",
				Recommendation: fmt.Sprintf("运行 'forgecrew handoff submit %s' 更新 task", taskID),
			})
			review.Status = "blocked"
			review.Decision.MergeAllowed = false
			allPassed = false
		} else if !stringSetsEqual(currentFiles, task.ChangedFiles) {
			// Check 2b: current files differ from task.ChangedFiles.
			diff := diffStringSets(currentFiles, task.ChangedFiles)
			findings = append(findings, runs.Finding{
				Severity:       "required",
				Issue:          fmt.Sprintf("当前 git 变更与 task.ChangedFiles 不一致: %s", diff),
				Recommendation: fmt.Sprintf("运行 'forgecrew handoff submit %s' 更新 changed_files", taskID),
			})
			review.Status = "changes_requested"
			review.Decision.MergeAllowed = false
			allPassed = false
		}
	}

	// Run verification commands.
	var verificationResults []runs.VerificationResult

	// Only run verification if no blocking issues so far.
	if allPassed {
		// Run go test.
		testOut, testErr := deps.runTest()
		testStatus := "passed"
		if testErr != nil {
			testStatus = "failed"
		}
		verificationResults = append(verificationResults, runs.VerificationResult{
			Command: "go test ./...",
			Status:  testStatus,
		})
		if testErr != nil {
			findings = append(findings, runs.Finding{
				Severity:       "required",
				File:           "",
				Issue:          fmt.Sprintf("go test 失败: %s", testErr.Error()),
				Recommendation: "修复测试失败后重新提交 handoff",
			})
			review.Status = "changes_requested"
			review.Decision.MergeAllowed = false
			allPassed = false
		}
		_ = testOut
	}

	if allPassed {
		// Run go vet.
		_, vetErr := deps.runVet()
		vetStatus := "passed"
		if vetErr != nil {
			vetStatus = "failed"
		}
		verificationResults = append(verificationResults, runs.VerificationResult{
			Command: "go vet ./...",
			Status:  vetStatus,
		})
		if vetErr != nil {
			findings = append(findings, runs.Finding{
				Severity:       "required",
				File:           "",
				Issue:          fmt.Sprintf("go vet 失败: %s", vetErr.Error()),
				Recommendation: "修复 vet 警告后重新提交 handoff",
			})
			review.Status = "changes_requested"
			review.Decision.MergeAllowed = false
			allPassed = false
		}
	}

	// All clear.
	if allPassed && review.Status == "" {
		review.Status = "approved"
		review.Decision.MergeAllowed = true
	}

	review.Findings = findings
	review.Verification = verificationResults

	// Read diff for the record.
	if diff, err := deps.readDiff(); err == nil && diff != "" {
		// Diff is captured for informational purposes; not stored in review
		// since review YAML doesn't have a diff field, but we can add it
		// as context in findings if needed.
		_ = diff
	}

	// Save review.
	if err := m.SaveReview(runID, taskID, review); err != nil {
		return fmt.Errorf("保存 review 失败: %w", err)
	}

	// Print summary.
	printReviewSummary(cmd, review)

	return nil
}

func printReviewSummary(cmd *cobra.Command, review *runs.Review) {
	fmt.Fprintf(cmd.OutOrStdout(), "\n=== CTO Review: %s ===\n", review.TaskID)
	fmt.Fprintf(cmd.OutOrStdout(), "状态: %s\n", review.Status)
	fmt.Fprintf(cmd.OutOrStdout(), "允许合并: %v\n", review.Decision.MergeAllowed)

	if len(review.Findings) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "\n发现 %d 个问题:\n", len(review.Findings))
		for _, f := range review.Findings {
			fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", f.Severity, f.Issue)
			if f.Recommendation != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "        建议: %s\n", f.Recommendation)
			}
		}
	}

	if len(review.Verification) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "\n验证命令:\n")
		for _, v := range review.Verification {
			mark := "✓"
			if v.Status != "passed" {
				mark = "✗"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  %s %s: %s\n", mark, v.Command, v.Status)
		}
	}
	fmt.Fprintln(cmd.OutOrStdout())
}

// runGoCmd runs a go subcommand and returns combined stdout/stderr.
func runGoCmd(subcmd, args string) (string, error) {
	cmd := exec.Command("go", append([]string{subcmd}, strings.Fields(args)...)...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// stringSetsEqual returns true if two string slices have the same elements (order-insensitive).
func stringSetsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	am := make(map[string]bool, len(a))
	for _, s := range a {
		am[s] = true
	}
	for _, s := range b {
		if !am[s] {
			return false
		}
	}
	return true
}

// diffStringSets returns a human-readable diff description between two sets.
func diffStringSets(current, stored []string) string {
	cm := make(map[string]bool, len(current))
	for _, s := range current {
		cm[s] = true
	}
	sm := make(map[string]bool, len(stored))
	for _, s := range stored {
		sm[s] = true
	}

	var added, removed []string
	for _, s := range current {
		if !sm[s] {
			added = append(added, s)
		}
	}
	for _, s := range stored {
		if !cm[s] {
			removed = append(removed, s)
		}
	}

	parts := []string{}
	if len(added) > 0 {
		parts = append(parts, fmt.Sprintf("新增: %s", strings.Join(added, ", ")))
	}
	if len(removed) > 0 {
		parts = append(parts, fmt.Sprintf("移除: %s", strings.Join(removed, ", ")))
	}
	return strings.Join(parts, "; ")
}
