package communication

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
)

// --- DefaultCadence tests ---

func TestDefaultCadence_HasFiveSteps(t *testing.T) {
	cadence := DefaultCadence()
	if len(cadence.Steps) != 5 {
		t.Errorf("expected 5 cadence steps, got %d", len(cadence.Steps))
	}
}

func TestDefaultCadence_MaxAttempts(t *testing.T) {
	cadence := DefaultCadence()
	if cadence.MaxAttempts != 5 {
		t.Errorf("expected max attempts 5, got %d", cadence.MaxAttempts)
	}
}

func TestDefaultCadence_StepOrder(t *testing.T) {
	cadence := DefaultCadence()

	expectedChannels := []db.Channel{
		db.ChannelEmail,
		db.ChannelGitHub,
		db.ChannelEmail,
		db.ChannelLinkedIn,
		db.ChannelEmail,
	}

	for i, step := range cadence.Steps {
		if step.StepNumber != i+1 {
			t.Errorf("step %d: expected StepNumber %d, got %d", i, i+1, step.StepNumber)
		}
		if step.Channel != expectedChannels[i] {
			t.Errorf("step %d: expected channel %s, got %s", i+1, expectedChannels[i], step.Channel)
		}
	}
}

func TestDefaultCadence_FirstStepIsImmediate(t *testing.T) {
	cadence := DefaultCadence()
	if cadence.Steps[0].DelayAfterPrev != 0 {
		t.Errorf("expected first step delay to be 0, got %v", cadence.Steps[0].DelayAfterPrev)
	}
}

func TestDefaultCadence_SubsequentStepsHaveDelays(t *testing.T) {
	cadence := DefaultCadence()
	for i := 1; i < len(cadence.Steps); i++ {
		if cadence.Steps[i].DelayAfterPrev <= 0 {
			t.Errorf("step %d should have a positive delay, got %v", i+1, cadence.Steps[i].DelayAfterPrev)
		}
	}
}

// --- CadenceTracker.GetNextStep tests ---

func TestCadenceTracker_GetNextStep_NoInteractions(t *testing.T) {
	tracker := &CadenceTracker{db: nil}
	schedule := DefaultCadence()

	step, progress, err := tracker.GetNextStep(context.Background(), 1, schedule, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if step == nil {
		t.Fatal("expected a step for deal with no interactions")
	}
	if step.StepNumber != 1 {
		t.Errorf("expected first step (1), got %d", step.StepNumber)
	}
	if progress.IsExhausted {
		t.Error("should not be exhausted with no interactions")
	}
	if progress.TotalAttempts != 0 {
		t.Errorf("expected 0 total attempts, got %d", progress.TotalAttempts)
	}
}

func TestCadenceTracker_GetNextStep_ExhaustedAfterMaxAttempts(t *testing.T) {
	tracker := &CadenceTracker{db: nil}
	schedule := DefaultCadence()

	// Create 5 outbound interactions (max attempts)
	now := time.Now()
	interactions := make([]db.Interaction, 5)
	for i := 0; i < 5; i++ {
		interactions[i] = db.Interaction{
			Direction: "Outbound",
			Channel:   "email",
			CreatedAt: now.Add(time.Duration(i) * time.Hour),
		}
	}

	step, progress, err := tracker.GetNextStep(context.Background(), 1, schedule, interactions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if step != nil {
		t.Error("expected nil step when exhausted")
	}
	if !progress.IsExhausted {
		t.Error("expected IsExhausted to be true after max attempts")
	}
	if progress.TotalAttempts != 5 {
		t.Errorf("expected 5 total attempts, got %d", progress.TotalAttempts)
	}
}

func TestCadenceTracker_GetNextStep_AdvancesAfterOutbound(t *testing.T) {
	tracker := &CadenceTracker{db: nil}
	schedule := DefaultCadence()

	// One outbound email interaction
	interactions := []db.Interaction{
		{
			Direction: "Outbound",
			Channel:   "email",
			CreatedAt: time.Now().Add(-72 * time.Hour), // 3 days ago
		},
	}

	step, progress, err := tracker.GetNextStep(context.Background(), 1, schedule, interactions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if step == nil {
		t.Fatal("expected a next step")
	}
	if step.StepNumber < 2 {
		t.Errorf("expected step >= 2 after first outbound, got %d", step.StepNumber)
	}
	if progress.TotalAttempts != 1 {
		t.Errorf("expected 1 total attempt, got %d", progress.TotalAttempts)
	}
}

func TestCadenceTracker_GetNextStep_InboundsDoNotCount(t *testing.T) {
	tracker := &CadenceTracker{db: nil}
	schedule := DefaultCadence()

	// Only inbound interactions — should still be on step 1
	interactions := []db.Interaction{
		{Direction: "Inbound", Channel: "email", CreatedAt: time.Now()},
		{Direction: "Inbound", Channel: "email", CreatedAt: time.Now()},
	}

	step, progress, err := tracker.GetNextStep(context.Background(), 1, schedule, interactions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if step == nil {
		t.Fatal("expected a step")
	}
	if step.StepNumber != 1 {
		t.Errorf("expected step 1 (inbounds don't count), got %d", step.StepNumber)
	}
	if progress.TotalAttempts != 0 {
		t.Errorf("expected 0 outbound attempts, got %d", progress.TotalAttempts)
	}
}

// --- CadenceTracker.IsTimeForNextStep tests ---

func TestCadenceTracker_IsTimeForNextStep_Exhausted(t *testing.T) {
	tracker := &CadenceTracker{db: nil}
	progress := CadenceProgress{IsExhausted: true}

	if tracker.IsTimeForNextStep(progress) {
		t.Error("should return false when exhausted")
	}
}

func TestCadenceTracker_IsTimeForNextStep_PastScheduledTime(t *testing.T) {
	tracker := &CadenceTracker{db: nil}
	progress := CadenceProgress{
		IsExhausted:       false,
		NextScheduledTime: time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	if !tracker.IsTimeForNextStep(progress) {
		t.Error("should return true when past scheduled time")
	}
}

func TestCadenceTracker_IsTimeForNextStep_FutureScheduledTime(t *testing.T) {
	tracker := &CadenceTracker{db: nil}
	progress := CadenceProgress{
		IsExhausted:       false,
		NextScheduledTime: time.Now().Add(24 * time.Hour), // tomorrow
	}

	if tracker.IsTimeForNextStep(progress) {
		t.Error("should return false when scheduled time is in the future")
	}
}

// --- CadenceProgress struct tests ---

func TestCadenceProgress_Fields(t *testing.T) {
	now := time.Now()
	progress := CadenceProgress{
		DealID:            42,
		ScheduleName:      "Test Schedule",
		NextStepNumber:    3,
		LastTouchTime:     now,
		NextScheduledTime: now.Add(48 * time.Hour),
		TotalAttempts:     2,
		IsExhausted:       false,
	}

	if progress.DealID != 42 {
		t.Errorf("unexpected DealID: %d", progress.DealID)
	}
	if progress.NextStepNumber != 3 {
		t.Errorf("unexpected NextStepNumber: %d", progress.NextStepNumber)
	}
	if progress.TotalAttempts != 2 {
		t.Errorf("unexpected TotalAttempts: %d", progress.TotalAttempts)
	}
}

// --- CadenceStep struct tests ---

func TestCadenceStep_Fields(t *testing.T) {
	step := CadenceStep{
		StepNumber:     1,
		Channel:        db.ChannelEmail,
		DelayAfterPrev: 0,
		TemplateID:     "intro-email",
		Subject:        "Hello %s",
	}

	if step.StepNumber != 1 {
		t.Errorf("unexpected StepNumber: %d", step.StepNumber)
	}
	if step.Channel != db.ChannelEmail {
		t.Errorf("unexpected Channel: %s", step.Channel)
	}
	if step.TemplateID != "intro-email" {
		t.Errorf("unexpected TemplateID: %s", step.TemplateID)
	}
}

// --- CadenceAwareManager construction tests ---

func TestNewCadenceAwareManager_NilDB(t *testing.T) {
	mgr := NewManager(nil, nil, nil, nil, nil, nil)
	cam := NewCadenceAwareManager(mgr, nil)

	if cam == nil {
		t.Fatal("expected non-nil CadenceAwareManager")
	}
	if cam.Manager != mgr {
		t.Error("expected Manager to be set")
	}
}

// --- CadenceAwareManager.checkCadence nil-safety test ---

func TestCadenceAwareManager_CheckCadence_NilDB(t *testing.T) {
	mgr := NewManager(nil, nil, nil, nil, nil, nil)
	cam := NewCadenceAwareManager(mgr, nil)

	// Should not panic with nil DB
	cam.checkCadence(context.Background())
}
