// Package orchestrator implements the Loop Engineering state machine.
// It manages the lifecycle: Goal → Plan → Retrieve → Act → Observe → Reflect → Improve → Review → CommitMemory.
package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/1424772/ForgeCrew/internal/i18n"
)

// Step represents one state in the Loop Engineering cycle.
type Step string

// The nine states of Loop Engineering.
const (
	StepGoal         Step = "goal"
	StepPlan         Step = "plan"
	StepRetrieve     Step = "retrieve"
	StepAct          Step = "act"
	StepObserve      Step = "observe"
	StepReflect      Step = "reflect"
	StepImprove      Step = "improve"
	StepReview       Step = "review"
	StepCommitMemory Step = "commit_memory"
	StepDone         Step = "done"
)

// StateSequence is the canonical order of states.
var StateSequence = []Step{
	StepGoal,
	StepPlan,
	StepRetrieve,
	StepAct,
	StepObserve,
	StepReflect,
	StepImprove,
	StepReview,
	StepCommitMemory,
}

// StateMachine manages the Loop Engineering lifecycle.
type StateMachine struct {
	CurrentStep Step
	Iteration   int
	MaxIter     int
	DryRun      bool
	History     []StateRecord
}

// StateRecord records a transition for auditing.
type StateRecord struct {
	Step      Step      `json:"step"`
	Iteration int       `json:"iteration"`
	Timestamp time.Time `json:"timestamp"`
	Note      string    `json:"note,omitempty"`
}

// New creates a new Loop Engineering state machine.
// maxIter is the total number of full cycles (default 3).
// Iterations start at 1, so New(3, ...) runs iterations 1, 2, 3.
func New(maxIter int, dryRun bool) *StateMachine {
	if maxIter <= 0 {
		maxIter = 3
	}
	return &StateMachine{
		CurrentStep: StepGoal,
		Iteration:   1,
		MaxIter:     maxIter,
		DryRun:      dryRun,
	}
}

// Next advances the state machine to the next step.
// Returns true if there are more steps to execute.
func (sm *StateMachine) Next() bool {
	idx := sm.indexOf(sm.CurrentStep)
	if idx < 0 {
		return false
	}

	// Record current step.
	sm.record(sm.CurrentStep, "")

	// Move to next.
	nextIdx := idx + 1
	if nextIdx >= len(StateSequence) {
		// End of cycle; loop if not at max iterations.
		if sm.Iteration < sm.MaxIter {
			sm.Iteration++
			// Restart from Plan (skip Goal on subsequent iterations).
			sm.CurrentStep = StepPlan
			sm.record(StepPlan, fmt.Sprintf("iteration %d restart", sm.Iteration))
			return true
		}
		sm.CurrentStep = StepDone
		sm.record(StepDone, "completed")
		return false
	}
	sm.CurrentStep = StateSequence[nextIdx]
	return true
}

// Execute runs the current step (dry-run: just logs the step).
// It checks ctx.Err() before execution so long-running sequences can be
// cancelled cooperatively.
func (sm *StateMachine) Execute(ctx context.Context, taskID string) (*StepResult, error) {
	// Check for cancellation before executing the step.
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("step %s cancelled: %w", sm.CurrentStep, err)
	}

	result := &StepResult{
		Step: sm.CurrentStep,
		OK:   true,
	}

	if sm.DryRun {
		result.Note = fmt.Sprintf("[dry-run] would execute %s for task %s", sm.CurrentStep, taskID)
		return result, nil
	}

	result.Note = fmt.Sprintf("executed %s for task %s", sm.CurrentStep, taskID)
	return result, nil
}

// StepResult is the output of executing a single step.
type StepResult struct {
	Step   Step   `json:"step"`
	OK     bool   `json:"ok"`
	Note   string `json:"note"`
	Output string `json:"output,omitempty"`
}

// Current returns the current step name.
func (sm *StateMachine) Current() Step {
	return sm.CurrentStep
}

// IsDone returns true if the state machine has completed.
func (sm *StateMachine) IsDone() bool {
	return sm.CurrentStep == StepDone
}

func (sm *StateMachine) indexOf(step Step) int {
	for i, s := range StateSequence {
		if s == step {
			return i
		}
	}
	return -1
}

func (sm *StateMachine) record(step Step, note string) {
	sm.History = append(sm.History, StateRecord{
		Step:      step,
		Iteration: sm.Iteration,
		Timestamp: time.Now(),
		Note:      note,
	})
}

// RunResult is the complete output of a full state machine run.
type RunResult struct {
	Goal   string       `json:"goal"`
	DryRun bool         `json:"dry_run"`
	Steps  []StepResult `json:"steps"`
}

// RunFull executes the complete Loop Engineering cycle for a given goal.
// It advances through all steps until completion and returns the full result.
// The context is checked before each step; if ctx is cancelled the loop
// stops and records the cancellation in the result.
func (sm *StateMachine) RunFull(ctx context.Context, goal string) *RunResult {
	result := &RunResult{
		Goal:   goal,
		DryRun: sm.DryRun,
	}
	for sm.Next() {
		sr, err := sm.Execute(ctx, goal)
		if err != nil {
			// Cancellation or other fatal error — record and stop.
			result.Steps = append(result.Steps, StepResult{
				Step: sm.CurrentStep,
				OK:   false,
				Note: err.Error(),
			})
			return result
		}
		result.Steps = append(result.Steps, *sr)
	}
	return result
}

// FormatText returns a human-readable representation of the run result
// with iteration grouping and dry-run markers. locale must be "zh" or "en".
func (rr *RunResult) FormatText(locale string) string {
	loc := i18n.Locale(locale)
	var b strings.Builder

	// Header.
	dryTag := ""
	if rr.DryRun {
		dryTag = i18n.T("task.dry_run_tag", loc)
	}
	fmt.Fprintf(&b, "%s%q%s\n", i18n.T("task.header", loc), rr.Goal, dryTag)

	if len(rr.Steps) == 0 {
		fmt.Fprint(&b, i18n.T("task.no_steps", loc))
		return b.String()
	}

	// Group steps by iteration. An iteration break happens when we see a
	// step we've already seen in the current cycle.
	seen := map[Step]bool{}
	iter := 1
	printedIterHeader := false

	for _, s := range rr.Steps {
		if seen[s.Step] {
			// New iteration: a step name repeated.
			iter++
			seen = map[Step]bool{}
			printedIterHeader = false
		}
		seen[s.Step] = true

		if !printedIterHeader {
			fmt.Fprintf(&b, "\n%s%d%s:\n", i18n.T("task.iteration", loc), iter, i18n.T("task.iteration_suffix", loc))
			printedIterHeader = true
		}
		if rr.DryRun {
			dryTag := i18n.T("task.dry_run_note", loc)
			desc := fmt.Sprintf(i18n.T("task.dry_run_format", loc), rr.Goal, string(s.Step))
			fmt.Fprintf(&b, "  %-16s %s%s\n", s.Step, dryTag, desc)
		} else {
			fmt.Fprintf(&b, "  %-16s %s\n", s.Step, s.Note)
		}
	}

	return b.String()
}
