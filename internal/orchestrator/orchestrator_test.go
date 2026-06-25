package orchestrator

import (
	"strings"
	"testing"
)

func TestNewStateMachine(t *testing.T) {
	sm := New(3, true)
	if sm.CurrentStep != StepGoal {
		t.Errorf("initial step = %q, want goal", sm.CurrentStep)
	}
	if sm.MaxIter != 3 {
		t.Errorf("maxIter = %d, want 3", sm.MaxIter)
	}
	if !sm.DryRun {
		t.Error("dryRun should be true")
	}
}

func TestDefaultMaxIter(t *testing.T) {
	sm := New(0, false)
	if sm.MaxIter != 3 {
		t.Errorf("default maxIter = %d, want 3", sm.MaxIter)
	}
}

func TestStateSequence(t *testing.T) {
	expectedOrder := []Step{
		StepGoal, StepPlan, StepRetrieve, StepAct, StepObserve,
		StepReflect, StepImprove, StepReview, StepCommitMemory,
	}
	for i, s := range StateSequence {
		if s != expectedOrder[i] {
			t.Errorf("StateSequence[%d] = %q, want %q", i, s, expectedOrder[i])
		}
	}
	if len(StateSequence) != 9 {
		t.Errorf("StateSequence length = %d, want 9", len(StateSequence))
	}
}

func TestSingleCycle(t *testing.T) {
	sm := New(1, true)
	steps := []Step{}
	for sm.Next() {
		steps = append(steps, sm.CurrentStep)
		_, err := sm.Execute("task_001")
		if err != nil {
			t.Fatal(err)
		}
	}
	steps = append(steps, sm.CurrentStep)

	if sm.CurrentStep != StepDone {
		t.Errorf("final step = %q, want done", sm.CurrentStep)
	}
}

func TestMultiCycle(t *testing.T) {
	sm := New(3, true)
	count := 0
	for sm.Next() {
		count++
	}
	// New(3, true) → 3 iterations (1, 2, 3).
	// Iteration 1: 9 Next() calls (Goal→Plan + 7 more transitions + restart).
	// Iteration 2: 8 Next() calls + restart.
	// Iteration 3: 7 Next() calls (last CommitMemory→Done returns false).
	// Total: 9 + 8 + 7 = 24.
	const want = 24
	if count != want {
		t.Errorf("expected exactly %d Next() calls across 3 iterations, got %d", want, count)
	}
	if sm.CurrentStep != StepDone {
		t.Errorf("final step = %q, want done", sm.CurrentStep)
	}
}

func TestExecuteDryRun(t *testing.T) {
	sm := New(1, true)
	result, err := sm.Execute("task_001")
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK {
		t.Error("dry-run result should be OK")
	}
	if result.Note == "" {
		t.Error("dry-run result should have a note")
	}
}

func TestHistoryRecording(t *testing.T) {
	sm := New(1, true)
	// Execute one full cycle.
	for sm.Next() {
		sm.Execute("task_001")
	}
	if len(sm.History) == 0 {
		t.Error("history should contain records")
	}
}

func TestIsDone(t *testing.T) {
	sm := New(1, true)
	if sm.IsDone() {
		t.Error("should not be done initially")
	}
	for sm.Next() {
	}
	if !sm.IsDone() {
		t.Error("should be done after completion")
	}
}

func TestRunFullDryRun(t *testing.T) {
	sm := New(1, true)
	result := sm.RunFull("test goal")

	if result.Goal != "test goal" {
		t.Errorf("goal = %q, want %q", result.Goal, "test goal")
	}
	if !result.DryRun {
		t.Error("DryRun should be true")
	}
	// New(1, true): 8 steps (Plan through CommitMemory, Goal not executed).
	const wantSteps = 8
	if len(result.Steps) != wantSteps {
		t.Errorf("expected exactly %d steps with 1 iteration, got %d", wantSteps, len(result.Steps))
	}
	if !sm.IsDone() {
		t.Error("state machine should be done after RunFull")
	}
}

func TestRunFullMultiIteration(t *testing.T) {
	sm := New(2, true)
	result := sm.RunFull("multi iter")

	// With 2 iterations: 8 steps per iteration = 16 exactly.
	const want = 16
	if len(result.Steps) != want {
		t.Errorf("expected exactly %d steps across 2 iterations, got %d", want, len(result.Steps))
	}
}

func TestFormatTextDryRun(t *testing.T) {
	sm := New(1, true)
	result := sm.RunFull("fix login bug")

	out := result.FormatText("en")
	if !strings.Contains(out, "fix login bug") {
		t.Error("FormatText should contain the goal")
	}
	if !strings.Contains(out, "[dry-run]") {
		t.Error("FormatText should mark dry-run")
	}
	if !strings.Contains(out, "Iteration 1") {
		t.Error("FormatText should show iteration headers starting at 1")
	}
}

func TestFormatTextContainsStates(t *testing.T) {
	sm := New(1, true)
	result := sm.RunFull("test states")

	out := result.FormatText("en")
	// All states except Goal should appear (Goal is recorded but never
	// executed by the current state machine design).
	for _, s := range StateSequence {
		if s == StepGoal {
			continue
		}
		if !strings.Contains(out, string(s)) {
			t.Errorf("FormatText should contain step %q", s)
		}
	}
}

func TestFormatTextMultiIteration(t *testing.T) {
	sm := New(2, true)
	result := sm.RunFull("multi")

	out := result.FormatText("en")
	// Iteration 1 is the first cycle; Iteration 2 is the second.
	if !strings.Contains(out, "Iteration 2") {
		t.Error("FormatText should show Iteration 2 for second iteration")
	}
}
