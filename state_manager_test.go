package orchwf

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryStateManager_SaveWorkflow(t *testing.T) {
	sm := NewInMemoryStateManager()
	ctx := context.Background()

	workflow := &WorkflowInstance{
		ID:         "test-workflow",
		WorkflowID: "test",
		Status:     WorkflowStatusPending,
		Input:      map[string]interface{}{"key": "value"},
		StartedAt:  time.Now(),
	}

	err := sm.SaveWorkflow(ctx, workflow)
	if err != nil {
		t.Errorf("SaveWorkflow() error = %v", err)
	}

	// Verify workflow was saved
	saved, err := sm.GetWorkflow(ctx, "test-workflow")
	if err != nil {
		t.Errorf("GetWorkflow() error = %v", err)
	}
	if saved.ID != "test-workflow" {
		t.Errorf("GetWorkflow() = %v, want %v", saved.ID, "test-workflow")
	}
}

func TestInMemoryStateManager_UpdateWorkflowStatus(t *testing.T) {
	sm := NewInMemoryStateManager()
	ctx := context.Background()

	workflow := &WorkflowInstance{
		ID:         "test-workflow",
		WorkflowID: "test",
		Status:     WorkflowStatusPending,
		StartedAt:  time.Now(),
	}

	sm.SaveWorkflow(ctx, workflow)

	err := sm.UpdateWorkflowStatus(ctx, "test-workflow", WorkflowStatusRunning)
	if err != nil {
		t.Errorf("UpdateWorkflowStatus() error = %v", err)
	}

	saved, _ := sm.GetWorkflow(ctx, "test-workflow")
	if saved.Status != WorkflowStatusRunning {
		t.Errorf("UpdateWorkflowStatus() = %v, want %v", saved.Status, WorkflowStatusRunning)
	}
}

func TestInMemoryStateManager_UpdateWorkflowOutput(t *testing.T) {
	sm := NewInMemoryStateManager()
	ctx := context.Background()

	workflow := &WorkflowInstance{
		ID:         "test-workflow",
		WorkflowID: "test",
		Status:     WorkflowStatusPending,
		StartedAt:  time.Now(),
	}

	sm.SaveWorkflow(ctx, workflow)

	output := map[string]interface{}{"result": "success"}
	err := sm.UpdateWorkflowOutput(ctx, "test-workflow", output)
	if err != nil {
		t.Errorf("UpdateWorkflowOutput() error = %v", err)
	}

	saved, _ := sm.GetWorkflow(ctx, "test-workflow")
	if !mapsEqual(saved.Output, output) {
		t.Errorf("UpdateWorkflowOutput() = %v, want %v", saved.Output, output)
	}
}

func TestInMemoryStateManager_UpdateWorkflowError(t *testing.T) {
	sm := NewInMemoryStateManager()
	ctx := context.Background()

	workflow := &WorkflowInstance{
		ID:         "test-workflow",
		WorkflowID: "test",
		Status:     WorkflowStatusPending,
		StartedAt:  time.Now(),
	}

	sm.SaveWorkflow(ctx, workflow)

	err := sm.UpdateWorkflowError(ctx, "test-workflow", &testError{"test error"})
	if err != nil {
		t.Errorf("UpdateWorkflowError() error = %v", err)
	}

	saved, _ := sm.GetWorkflow(ctx, "test-workflow")
	if saved.Status != WorkflowStatusFailed {
		t.Errorf("UpdateWorkflowError() status = %v, want %v", saved.Status, WorkflowStatusFailed)
	}
	if saved.Error == nil || *saved.Error != "test error" {
		t.Errorf("UpdateWorkflowError() error = %v, want %v", saved.Error, "test error")
	}
}

func TestInMemoryStateManager_ListWorkflows(t *testing.T) {
	sm := NewInMemoryStateManager()
	ctx := context.Background()

	// Save multiple workflows
	workflows := []*WorkflowInstance{
		{ID: "wf1", WorkflowID: "test", Status: WorkflowStatusCompleted, StartedAt: time.Now()},
		{ID: "wf2", WorkflowID: "test", Status: WorkflowStatusRunning, StartedAt: time.Now()},
		{ID: "wf3", WorkflowID: "other", Status: WorkflowStatusPending, StartedAt: time.Now()},
	}

	for _, wf := range workflows {
		sm.SaveWorkflow(ctx, wf)
	}

	// Test listing all workflows
	all, total, err := sm.ListWorkflows(ctx, map[string]interface{}{}, 10, 0)
	if err != nil {
		t.Errorf("ListWorkflows() error = %v", err)
	}
	if len(all) != 3 {
		t.Errorf("ListWorkflows() count = %v, want %v", len(all), 3)
	}
	if total != 3 {
		t.Errorf("ListWorkflows() total = %v, want %v", total, 3)
	}

	// Test filtering by workflow_id
	filtered, total, err := sm.ListWorkflows(ctx, map[string]interface{}{"workflow_id": "test"}, 10, 0)
	if err != nil {
		t.Errorf("ListWorkflows() error = %v", err)
	}
	if len(filtered) != 2 {
		t.Errorf("ListWorkflows() filtered count = %v, want %v", len(filtered), 2)
	}
}

func TestInMemoryStateManager_SaveStep(t *testing.T) {
	sm := NewInMemoryStateManager()
	ctx := context.Background()

	step := &StepInstance{
		ID:             "test-step",
		StepID:         "step1",
		WorkflowInstID: "test-workflow",
		Status:         StepStatusPending,
		Input:          map[string]interface{}{"key": "value"},
		ExecutionOrder: 1,
	}

	err := sm.SaveStep(ctx, step)
	if err != nil {
		t.Errorf("SaveStep() error = %v", err)
	}

	// Verify step was saved
	saved, err := sm.GetStep(ctx, "test-step")
	if err != nil {
		t.Errorf("GetStep() error = %v", err)
	}
	if saved.ID != "test-step" {
		t.Errorf("GetStep() = %v, want %v", saved.ID, "test-step")
	}
}

func TestInMemoryStateManager_GetWorkflowSteps(t *testing.T) {
	sm := NewInMemoryStateManager()
	ctx := context.Background()

	workflowID := "test-workflow"
	steps := []*StepInstance{
		{ID: "step1", StepID: "s1", WorkflowInstID: workflowID, ExecutionOrder: 1},
		{ID: "step2", StepID: "s2", WorkflowInstID: workflowID, ExecutionOrder: 2},
		{ID: "step3", StepID: "s3", WorkflowInstID: "other-workflow", ExecutionOrder: 1},
	}

	for _, step := range steps {
		sm.SaveStep(ctx, step)
	}

	workflowSteps, err := sm.GetWorkflowSteps(ctx, workflowID)
	if err != nil {
		t.Errorf("GetWorkflowSteps() error = %v", err)
	}
	if len(workflowSteps) != 2 {
		t.Errorf("GetWorkflowSteps() count = %v, want %v", len(workflowSteps), 2)
	}
}

func TestInMemoryStateManager_UpdateStepStatus(t *testing.T) {
	sm := NewInMemoryStateManager()
	ctx := context.Background()

	step := &StepInstance{
		ID:             "test-step",
		StepID:         "step1",
		WorkflowInstID: "test-workflow",
		Status:         StepStatusPending,
		ExecutionOrder: 1,
	}

	sm.SaveStep(ctx, step)

	err := sm.UpdateStepStatus(ctx, "test-step", StepStatusRunning)
	if err != nil {
		t.Errorf("UpdateStepStatus() error = %v", err)
	}

	saved, _ := sm.GetStep(ctx, "test-step")
	if saved.Status != StepStatusRunning {
		t.Errorf("UpdateStepStatus() = %v, want %v", saved.Status, StepStatusRunning)
	}
}

func TestInMemoryStateManager_SaveEvent(t *testing.T) {
	sm := NewInMemoryStateManager()
	ctx := context.Background()

	event := &WorkflowEvent{
		ID:             "test-event",
		WorkflowInstID: "test-workflow",
		EventType:      "test.event",
		EventData:      map[string]interface{}{"key": "value"},
		Timestamp:      time.Now(),
	}

	err := sm.SaveEvent(ctx, event)
	if err != nil {
		t.Errorf("SaveEvent() error = %v", err)
	}

	// Verify event was saved
	events, err := sm.GetWorkflowEvents(ctx, "test-workflow")
	if err != nil {
		t.Errorf("GetWorkflowEvents() error = %v", err)
	}
	if len(events) != 1 {
		t.Errorf("GetWorkflowEvents() count = %v, want %v", len(events), 1)
	}
	if events[0].ID != "test-event" {
		t.Errorf("GetWorkflowEvents() = %v, want %v", events[0].ID, "test-event")
	}
}

func TestInMemoryStateManager_WithTransaction(t *testing.T) {
	sm := NewInMemoryStateManager()
	ctx := context.Background()

	err := sm.WithTransaction(ctx, func(txCtx context.Context) error {
		workflow := &WorkflowInstance{
			ID:         "test-workflow",
			WorkflowID: "test",
			Status:     WorkflowStatusPending,
			StartedAt:  time.Now(),
		}
		return sm.SaveWorkflow(txCtx, workflow)
	})

	if err != nil {
		t.Errorf("WithTransaction() error = %v", err)
	}

	// Verify workflow was saved
	_, err = sm.GetWorkflow(ctx, "test-workflow")
	if err != nil {
		t.Errorf("GetWorkflow() after transaction error = %v", err)
	}
}

func TestInMemoryStateManager_DeepCopy(t *testing.T) {
	sm := NewInMemoryStateManager()
	ctx := context.Background()

	// Test deep copy of workflow
	original := &WorkflowInstance{
		ID:         "test-workflow",
		WorkflowID: "test",
		Status:     WorkflowStatusPending,
		Input:      map[string]interface{}{"key": "value"},
		Context:    map[string]interface{}{"ctx": "data"},
		StartedAt:  time.Now(),
	}

	sm.SaveWorkflow(ctx, original)
	copied, _ := sm.GetWorkflow(ctx, "test-workflow")

	// Modify original
	original.Input["key"] = "modified"
	original.Context["ctx"] = "modified"

	// Copied should not be affected
	if copied.Input["key"] == "modified" {
		t.Errorf("Deep copy failed - input was modified")
	}
	if copied.Context["ctx"] == "modified" {
		t.Errorf("Deep copy failed - context was modified")
	}
}
