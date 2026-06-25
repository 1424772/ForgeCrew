// Package orchestrator implements the Loop Engineering state machine.
// It manages the lifecycle: Goal → Plan → Retrieve → Act → Observe → Reflect → Improve → Review → CommitMemory.
package orchestrator

import (
	"fmt"
	"time"
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
// maxIter is the maximum number of full cycles (default 3).
func New(maxIter int, dryRun bool) *StateMachine {
	if maxIter <= 0 {
		maxIter = 3
	}
	return &StateMachine{
		CurrentStep: StepGoal,
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
func (sm *StateMachine) Execute(taskID string) (*StepResult, error) {
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
