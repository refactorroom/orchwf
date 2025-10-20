package orchwf

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewOrchestrator(t *testing.T) {
	sm := NewInMemoryStateManager()
	orchestrator := NewOrchestrator(sm)

	if orchestrator == nil {
		t.Errorf("NewOrchestrator() returned nil")
	}
	if orchestrator.stateManager != sm {
		t.Errorf("NewOrchestrator() stateManager = %v, want %v", orchestrator.stateManager, sm)
	}
}

func TestNewOrchestratorWithAsyncWorkers(t *testing.T) {
	sm := NewInMemoryStateManager()
	orchestrator := NewOrchestratorWithAsyncWorkers(sm, 5)

	if orchestrator == nil {
		t.Errorf("NewOrchestratorWithAsyncWorkers() returned nil")
	}
	if orchestrator.asyncWorkers != 5 {
		t.Errorf("NewOrchestratorWithAsyncWorkers() asyncWorkers = %v, want %v", orchestrator.asyncWorkers, 5)
	}
}

func TestOrchestrator_RegisterWorkflow(t *testing.T) {
	orchestrator := NewOrchestrator(NewInMemoryStateManager())

	// Test valid workflow
	workflow := &WorkflowDefinition{
		ID:   "test-workflow",
		Name: "Test Workflow",
		Steps: []*StepDefinition{
			{ID: "step1", Name: "Step 1", Executor: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
				return map[string]interface{}{"result": "success"}, nil
			}},
		},
	}

	err := orchestrator.RegisterWorkflow(workflow)
	if err != nil {
		t.Errorf("RegisterWorkflow() error = %v", err)
	}

	// Test nil workflow
	err = orchestrator.RegisterWorkflow(nil)
	if err == nil {
		t.Errorf("RegisterWorkflow() with nil workflow should return error")
	}

	// Test empty ID
	workflow.ID = ""
	err = orchestrator.RegisterWorkflow(workflow)
	if err == nil {
		t.Errorf("RegisterWorkflow() with empty ID should return error")
	}
}

func TestOrchestrator_GetWorkflow(t *testing.T) {
	orchestrator := NewOrchestrator(NewInMemoryStateManager())

	workflow := &WorkflowDefinition{
		ID:   "test-workflow",
		Name: "Test Workflow",
		Steps: []*StepDefinition{
			{ID: "step1", Name: "Step 1", Executor: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
				return map[string]interface{}{"result": "success"}, nil
			}},
		},
	}

	orchestrator.RegisterWorkflow(workflow)

	// Test existing workflow
	got, err := orchestrator.GetWorkflow("test-workflow")
	if err != nil {
		t.Errorf("GetWorkflow() error = %v", err)
	}
	if got.ID != "test-workflow" {
		t.Errorf("GetWorkflow() = %v, want %v", got.ID, "test-workflow")
	}

	// Test non-existing workflow
	_, err = orchestrator.GetWorkflow("non-existing")
	if err == nil {
		t.Errorf("GetWorkflow() with non-existing workflow should return error")
	}
}

func TestOrchestrator_StartWorkflow(t *testing.T) {
	orchestrator := NewOrchestrator(NewInMemoryStateManager())

	// Create a simple workflow
	step, _ := NewStepBuilder("step1", "Test Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "success"}, nil
	}).Build()

	workflow, _ := NewWorkflowBuilder("test-workflow", "Test Workflow").
		AddStep(step).
		Build()

	orchestrator.RegisterWorkflow(workflow)

	// Test successful execution
	result, err := orchestrator.StartWorkflow(context.Background(), "test-workflow",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Errorf("StartWorkflow() error = %v", err)
	}
	if !result.Success {
		t.Errorf("StartWorkflow() success = %v, want %v", result.Success, true)
	}
	if result.WorkflowInst.Status != WorkflowStatusCompleted {
		t.Errorf("StartWorkflow() status = %v, want %v", result.WorkflowInst.Status, WorkflowStatusCompleted)
	}

	// Test non-existing workflow
	_, err = orchestrator.StartWorkflow(context.Background(), "non-existing",
		map[string]interface{}{"data": "test"}, nil)
	if err == nil {
		t.Errorf("StartWorkflow() with non-existing workflow should return error")
	}
}

func TestOrchestrator_StartWorkflowAsync(t *testing.T) {
	orchestrator := NewOrchestrator(NewInMemoryStateManager())

	// Create a simple workflow
	step, _ := NewStepBuilder("step1", "Test Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return map[string]interface{}{"result": "success"}, nil
	}).Build()

	workflow, _ := NewWorkflowBuilder("test-workflow", "Test Workflow").
		AddStep(step).
		Build()

	orchestrator.RegisterWorkflow(workflow)

	// Test async execution
	workflowID, err := orchestrator.StartWorkflowAsync(context.Background(), "test-workflow",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Errorf("StartWorkflowAsync() error = %v", err)
	}
	if workflowID == "" {
		t.Errorf("StartWorkflowAsync() returned empty workflow ID")
	}

	// Wait a bit for async execution
	time.Sleep(200 * time.Millisecond)

	// Check workflow status
	status, err := orchestrator.GetWorkflowStatus(context.Background(), workflowID)
	if err != nil {
		t.Errorf("GetWorkflowStatus() error = %v", err)
	}
	if status.Status != WorkflowStatusCompleted {
		t.Errorf("GetWorkflowStatus() status = %v, want %v", status.Status, WorkflowStatusCompleted)
	}
}

func TestOrchestrator_ResumeWorkflow(t *testing.T) {
	orchestrator := NewOrchestrator(NewInMemoryStateManager())

	// Create a workflow that fails
	step, _ := NewStepBuilder("step1", "Failing Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return nil, errors.New("step failed")
	}).Build()

	workflow, _ := NewWorkflowBuilder("test-workflow", "Test Workflow").
		AddStep(step).
		Build()

	orchestrator.RegisterWorkflow(workflow)

	// Start workflow (it will fail)
	_, err := orchestrator.StartWorkflow(context.Background(), "test-workflow",
		map[string]interface{}{"data": "test"}, nil)

	if err == nil {
		t.Errorf("StartWorkflow() should have failed")
	}

	// Get the workflow instance ID from the error
	// For this test, we'll create a workflow instance manually
	instance := &WorkflowInstance{
		ID:         uuid.New().String(),
		WorkflowID: "test-workflow",
		Status:     WorkflowStatusFailed,
		StartedAt:  time.Now(),
		Steps: []*StepInstance{
			{ID: uuid.New().String(), StepID: "step1", WorkflowInstID: "", Status: StepStatusFailed},
		},
	}

	orchestrator.stateManager.SaveWorkflow(context.Background(), instance)

	// Resume workflow
	result, err := orchestrator.ResumeWorkflow(context.Background(), instance.ID)
	if err != nil {
		t.Errorf("ResumeWorkflow() error = %v", err)
	}
	if result.Success {
		t.Errorf("ResumeWorkflow() success = %v, want %v", result.Success, false)
	}
}

func TestOrchestrator_ListWorkflows(t *testing.T) {
	orchestrator := NewOrchestrator(NewInMemoryStateManager())

	// Create and start a workflow
	step, _ := NewStepBuilder("step1", "Test Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "success"}, nil
	}).Build()

	workflow, _ := NewWorkflowBuilder("test-workflow", "Test Workflow").
		AddStep(step).
		Build()

	orchestrator.RegisterWorkflow(workflow)
	orchestrator.StartWorkflow(context.Background(), "test-workflow",
		map[string]interface{}{"data": "test"}, nil)

	// List workflows
	workflows, total, err := orchestrator.ListWorkflows(context.Background(), map[string]interface{}{}, 10, 0)
	if err != nil {
		t.Errorf("ListWorkflows() error = %v", err)
	}
	if len(workflows) != 1 {
		t.Errorf("ListWorkflows() count = %v, want %v", len(workflows), 1)
	}
	if total != 1 {
		t.Errorf("ListWorkflows() total = %v, want %v", total, 1)
	}
}

func TestOrchestrator_StepDependencies(t *testing.T) {
	orchestrator := NewOrchestrator(NewInMemoryStateManager())

	// Create workflow with dependencies
	step1, _ := NewStepBuilder("step1", "Step 1", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"step1_result": "success"}, nil
	}).Build()

	step2, _ := NewStepBuilder("step2", "Step 2", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		// Check that step1 output is available
		if input["step1_result"] != "success" {
			return nil, errors.New("step1 output not available")
		}
		return map[string]interface{}{"step2_result": "success"}, nil
	}).WithDependencies("step1").Build()

	workflow, _ := NewWorkflowBuilder("test-workflow", "Test Workflow").
		AddStep(step1).
		AddStep(step2).
		Build()

	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "test-workflow",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Errorf("StartWorkflow() with dependencies error = %v", err)
	}
	if !result.Success {
		t.Errorf("StartWorkflow() with dependencies success = %v, want %v", result.Success, true)
	}
}

func TestOrchestrator_AsyncSteps(t *testing.T) {
	orchestrator := NewOrchestrator(NewInMemoryStateManager())

	// Create workflow with async steps
	step1, _ := NewStepBuilder("step1", "Step 1", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(50 * time.Millisecond)
		return map[string]interface{}{"step1_result": "success"}, nil
	}).WithAsync(true).Build()

	step2, _ := NewStepBuilder("step2", "Step 2", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(50 * time.Millisecond)
		return map[string]interface{}{"step2_result": "success"}, nil
	}).WithAsync(true).Build()

	step3, _ := NewStepBuilder("step3", "Step 3", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		// Both async steps should have completed
		if input["step1_result"] != "success" || input["step2_result"] != "success" {
			return nil, errors.New("async steps not completed")
		}
		return map[string]interface{}{"step3_result": "success"}, nil
	}).WithDependencies("step1", "step2").Build()

	workflow, _ := NewWorkflowBuilder("test-workflow", "Test Workflow").
		AddStep(step1).
		AddStep(step2).
		AddStep(step3).
		Build()

	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "test-workflow",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Errorf("StartWorkflow() with async steps error = %v", err)
	}
	if !result.Success {
		t.Errorf("StartWorkflow() with async steps success = %v, want %v", result.Success, true)
	}
}

func TestOrchestrator_RetryPolicy(t *testing.T) {
	orchestrator := NewOrchestrator(NewInMemoryStateManager())

	retryCount := 0
	step, _ := NewStepBuilder("step1", "Failing Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		retryCount++
		if retryCount < 3 {
			return nil, errors.New("temporary failure")
		}
		return map[string]interface{}{"result": "success"}, nil
	}).WithRetryPolicy(NewRetryPolicyBuilder().
		WithMaxAttempts(3).
		WithInitialInterval(1 * time.Millisecond).
		Build()).
		Build()

	workflow, _ := NewWorkflowBuilder("test-workflow", "Test Workflow").
		AddStep(step).
		Build()

	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "test-workflow",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Errorf("StartWorkflow() with retry error = %v", err)
	}
	if !result.Success {
		t.Errorf("StartWorkflow() with retry success = %v, want %v", result.Success, true)
	}
	if retryCount != 3 {
		t.Errorf("Retry count = %v, want %v", retryCount, 3)
	}
}

func TestOrchestrator_OptionalSteps(t *testing.T) {
	orchestrator := NewOrchestrator(NewInMemoryStateManager())

	// Create workflow with optional failing step
	step1, _ := NewStepBuilder("step1", "Step 1", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"step1_result": "success"}, nil
	}).Build()

	step2, _ := NewStepBuilder("step2", "Failing Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return nil, errors.New("step failed")
	}).WithRequired(false).Build()

	workflow, _ := NewWorkflowBuilder("test-workflow", "Test Workflow").
		AddStep(step1).
		AddStep(step2).
		Build()

	orchestrator.RegisterWorkflow(workflow)

	result, err := orchestrator.StartWorkflow(context.Background(), "test-workflow",
		map[string]interface{}{"data": "test"}, nil)

	if err != nil {
		t.Errorf("StartWorkflow() with optional step error = %v", err)
	}
	if !result.Success {
		t.Errorf("StartWorkflow() with optional step success = %v, want %v", result.Success, true)
	}
}
