package orchestrator

import (
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
	// After first pass (Goal→...→CommitMemory) = 9 steps minus goal which already passed = 8 steps.
	// Then restart from Plan: 2 more iterations × 8 steps = 16 steps.
	// Total = 8 + 16 = 24, plus final next returns false.
	if count < 20 {
		t.Errorf("expected at least 20 steps across 3 iterations, got %d", count)
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
