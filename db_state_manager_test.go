package orchwf

import (
	"testing"
	"time"
)

func TestDBStateManager_NewDBStateManager(t *testing.T) {
	// Test with nil database (should not panic)
	manager := NewDBStateManager(nil)

	if manager == nil {
		t.Errorf("NewDBStateManager() returned nil")
	}
}

func TestDBStateManager_InterfaceCompliance(t *testing.T) {
	// Test that DBStateManager implements StateManager interface
	var _ StateManager = (*DBStateManager)(nil)
}

func TestDBStateManager_JSONBConversion(t *testing.T) {
	// Test JSONB conversion functions
	input := map[string]interface{}{"key": "value", "number": 123}
	jsonb := JSONB(input)

	// Test Value method
	value, err := jsonb.Value()
	if err != nil {
		t.Errorf("JSONB.Value() error = %v", err)
	}

	// Test Scan method
	var newJSONB JSONB
	err = newJSONB.Scan(value)
	if err != nil {
		t.Errorf("JSONB.Scan() error = %v", err)
	}

	// Verify data integrity - JSON marshaling converts numbers to float64
	if len(newJSONB) != len(input) {
		t.Errorf("JSONB conversion failed: got length %d, want %d", len(newJSONB), len(input))
	}

	// Check string value
	if newJSONB["key"] != "value" {
		t.Errorf("JSONB conversion failed for key 'key': got %v, want %v", newJSONB["key"], "value")
	}

	// Check number value (JSON converts int to float64)
	if newJSONB["number"] != float64(123) {
		t.Errorf("JSONB conversion failed for key 'number': got %v, want %v", newJSONB["number"], float64(123))
	}
}

func TestDBStateManager_ModelConversions(t *testing.T) {
	// Test workflow instance conversion
	now := time.Now()
	workflow := &WorkflowInstance{
		ID:            "test-workflow",
		WorkflowID:    "test",
		Status:        WorkflowStatusCompleted,
		Input:         map[string]interface{}{"key": "value"},
		Output:        map[string]interface{}{"result": "success"},
		Context:       map[string]interface{}{"ctx": "data"},
		CurrentStepID: "step1",
		StartedAt:     now,
		CompletedAt:   &now,
		Error:         stringPtr("test error"),
		RetryCount:    2,
		LastRetryAt:   &now,
		Metadata:      map[string]interface{}{"meta": "data"},
		TraceID:       "trace123",
		CorrelationID: "corr123",
		BusinessID:    "biz123",
		Steps: []*StepInstance{
			{ID: "step1", StepID: "s1", WorkflowInstID: "test-workflow", Status: StepStatusCompleted},
		},
	}

	// Test conversion to model
	model := workflowInstanceToModel(workflow)
	if model.ID != "test-workflow" {
		t.Errorf("workflowInstanceToModel() ID = %v, want %v", model.ID, "test-workflow")
	}

	// Test conversion back to workflow
	convertedWorkflow, err := modelToWorkflowInstance(model)
	if err != nil {
		t.Errorf("modelToWorkflowInstance() error = %v", err)
	}
	if convertedWorkflow.ID != "test-workflow" {
		t.Errorf("modelToWorkflowInstance() ID = %v, want %v", convertedWorkflow.ID, "test-workflow")
	}
}

func TestDBStateManager_StepConversions(t *testing.T) {
	// Test step instance conversion
	now := time.Now()
	step := &StepInstance{
		ID:             "test-step",
		StepID:         "step1",
		WorkflowInstID: "test-workflow",
		Status:         StepStatusCompleted,
		Input:          map[string]interface{}{"key": "value"},
		Output:         map[string]interface{}{"result": "success"},
		StartedAt:      &now,
		CompletedAt:    &now,
		Error:          stringPtr("test error"),
		RetryCount:     2,
		LastRetryAt:    &now,
		DurationMs:     1000,
		ExecutionOrder: 1,
	}

	// Test conversion to model
	model := stepInstanceToModel(step)
	if model.ID != "test-step" {
		t.Errorf("stepInstanceToModel() ID = %v, want %v", model.ID, "test-step")
	}

	// Test conversion back to step
	convertedStep, err := modelToStepInstance(model)
	if err != nil {
		t.Errorf("modelToStepInstance() error = %v", err)
	}
	if convertedStep.ID != "test-step" {
		t.Errorf("modelToStepInstance() ID = %v, want %v", convertedStep.ID, "test-step")
	}
}

func TestDBStateManager_EventConversions(t *testing.T) {
	// Test workflow event conversion
	now := time.Now()
	event := &WorkflowEvent{
		ID:             "test-event",
		WorkflowInstID: "test-workflow",
		StepInstID:     stringPtr("test-step"),
		EventType:      "test.event",
		EventData:      map[string]interface{}{"key": "value"},
		Timestamp:      now,
	}

	// Test conversion to model
	model := workflowEventToModel(event)
	if model.ID != "test-event" {
		t.Errorf("workflowEventToModel() ID = %v, want %v", model.ID, "test-event")
	}

	// Test conversion back to event
	convertedEvent, err := modelToWorkflowEvent(model)
	if err != nil {
		t.Errorf("modelToWorkflowEvent() error = %v", err)
	}
	if convertedEvent.ID != "test-event" {
		t.Errorf("modelToWorkflowEvent() ID = %v, want %v", convertedEvent.ID, "test-event")
	}
}

func TestDBStateManager_WithTransaction(t *testing.T) {
	// Test transaction wrapper - skip this test as it requires a real database
	// The WithTransaction method will panic with nil db, so we'll test the interface compliance instead
	t.Skip("Skipping WithTransaction test as it requires a real database connection")
}
