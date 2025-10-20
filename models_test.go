package orchwf

import (
	"database/sql/driver"
	"testing"
	"time"
)

func TestJSONB_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    JSONB
		wantErr bool
	}{
		{
			name:    "nil value",
			value:   nil,
			want:    JSONB{},
			wantErr: false,
		},
		{
			name:    "byte slice",
			value:   []byte(`{"key": "value"}`),
			want:    JSONB{"key": "value"},
			wantErr: false,
		},
		{
			name:    "string",
			value:   `{"key": "value"}`,
			want:    JSONB{"key": "value"},
			wantErr: false,
		},
		{
			name:    "invalid type",
			value:   123,
			want:    JSONB{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j JSONB
			err := j.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONB.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !mapsEqual(j, tt.want) {
				t.Errorf("JSONB.Scan() = %v, want %v", j, tt.want)
			}
		})
	}
}

func TestJSONB_Value(t *testing.T) {
	tests := []struct {
		name    string
		jsonb   JSONB
		want    driver.Value
		wantErr bool
	}{
		{
			name:    "nil JSONB",
			jsonb:   nil,
			want:    nil,
			wantErr: false,
		},
		{
			name:    "empty JSONB",
			jsonb:   JSONB{},
			want:    []byte(`{}`),
			wantErr: false,
		},
		{
			name:    "valid JSONB",
			jsonb:   JSONB{"key": "value", "number": 123},
			want:    []byte(`{"key":"value","number":123}`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.jsonb.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONB.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.want == nil {
					if got != nil {
						t.Errorf("JSONB.Value() = %v, want %v", got, tt.want)
					}
				} else {
					gotBytes, ok := got.([]byte)
					if !ok {
						t.Errorf("JSONB.Value() returned %T, want []byte", got)
						return
					}
					wantBytes := tt.want.([]byte)
					if string(gotBytes) != string(wantBytes) {
						t.Errorf("JSONB.Value() = %v, want %v", string(gotBytes), string(wantBytes))
					}
				}
			}
		})
	}
}

func TestWorkflowInstanceToModel(t *testing.T) {
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

	model := workflowInstanceToModel(workflow)

	if model.ID != "test-workflow" {
		t.Errorf("workflowInstanceToModel() ID = %v, want %v", model.ID, "test-workflow")
	}
	if model.WorkflowID != "test" {
		t.Errorf("workflowInstanceToModel() WorkflowID = %v, want %v", model.WorkflowID, "test")
	}
	if model.Status != "completed" {
		t.Errorf("workflowInstanceToModel() Status = %v, want %v", model.Status, "completed")
	}
	if model.CurrentStepID == nil || *model.CurrentStepID != "step1" {
		t.Errorf("workflowInstanceToModel() CurrentStepID = %v, want %v", model.CurrentStepID, "step1")
	}
	if model.StartedAt != now {
		t.Errorf("workflowInstanceToModel() StartedAt = %v, want %v", model.StartedAt, now)
	}
	if model.CompletedAt == nil || *model.CompletedAt != now {
		t.Errorf("workflowInstanceToModel() CompletedAt = %v, want %v", model.CompletedAt, now)
	}
	if model.Error == nil || *model.Error != "test error" {
		t.Errorf("workflowInstanceToModel() Error = %v, want %v", model.Error, "test error")
	}
	if model.RetryCount != 2 {
		t.Errorf("workflowInstanceToModel() RetryCount = %v, want %v", model.RetryCount, 2)
	}
	if model.LastRetryAt == nil || *model.LastRetryAt != now {
		t.Errorf("workflowInstanceToModel() LastRetryAt = %v, want %v", model.LastRetryAt, now)
	}
	if model.TraceID != "trace123" {
		t.Errorf("workflowInstanceToModel() TraceID = %v, want %v", model.TraceID, "trace123")
	}
	if model.CorrelationID != "corr123" {
		t.Errorf("workflowInstanceToModel() CorrelationID = %v, want %v", model.CorrelationID, "corr123")
	}
	if model.BusinessID != "biz123" {
		t.Errorf("workflowInstanceToModel() BusinessID = %v, want %v", model.BusinessID, "biz123")
	}

	// Test JSONB fields
	if model.Input == nil {
		t.Errorf("workflowInstanceToModel() Input is nil")
	} else {
		input := map[string]interface{}(*model.Input)
		if input["key"] != "value" {
			t.Errorf("workflowInstanceToModel() Input = %v, want %v", input, map[string]interface{}{"key": "value"})
		}
	}
}

func TestModelToWorkflowInstance(t *testing.T) {
	now := time.Now()
	inputJSON := JSONB{"key": "value"}
	outputJSON := JSONB{"result": "success"}
	contextJSON := JSONB{"ctx": "data"}
	metadataJSON := JSONB{"meta": "data"}

	model := &ORCHWorkflowInstance{
		ID:            "test-workflow",
		WorkflowID:    "test",
		Status:        "completed",
		Input:         &inputJSON,
		Output:        &outputJSON,
		Context:       &contextJSON,
		CurrentStepID: stringPtr("step1"),
		StartedAt:     now,
		CompletedAt:   &now,
		Error:         stringPtr("test error"),
		RetryCount:    2,
		LastRetryAt:   &now,
		Metadata:      &metadataJSON,
		TraceID:       "trace123",
		CorrelationID: "corr123",
		BusinessID:    "biz123",
		Steps: []ORCHStepInstance{
			{ID: "step1", StepID: "s1", WorkflowInstID: "test-workflow", Status: "completed"},
		},
	}

	workflow, err := modelToWorkflowInstance(model)
	if err != nil {
		t.Errorf("modelToWorkflowInstance() error = %v", err)
		return
	}

	if workflow.ID != "test-workflow" {
		t.Errorf("modelToWorkflowInstance() ID = %v, want %v", workflow.ID, "test-workflow")
	}
	if workflow.WorkflowID != "test" {
		t.Errorf("modelToWorkflowInstance() WorkflowID = %v, want %v", workflow.WorkflowID, "test")
	}
	if workflow.Status != WorkflowStatusCompleted {
		t.Errorf("modelToWorkflowInstance() Status = %v, want %v", workflow.Status, WorkflowStatusCompleted)
	}
	if workflow.CurrentStepID != "step1" {
		t.Errorf("modelToWorkflowInstance() CurrentStepID = %v, want %v", workflow.CurrentStepID, "step1")
	}
	if workflow.StartedAt != now {
		t.Errorf("modelToWorkflowInstance() StartedAt = %v, want %v", workflow.StartedAt, now)
	}
	if workflow.CompletedAt == nil || *workflow.CompletedAt != now {
		t.Errorf("modelToWorkflowInstance() CompletedAt = %v, want %v", workflow.CompletedAt, now)
	}
	if workflow.Error == nil || *workflow.Error != "test error" {
		t.Errorf("modelToWorkflowInstance() Error = %v, want %v", workflow.Error, "test error")
	}
	if workflow.RetryCount != 2 {
		t.Errorf("modelToWorkflowInstance() RetryCount = %v, want %v", workflow.RetryCount, 2)
	}
	if workflow.LastRetryAt == nil || *workflow.LastRetryAt != now {
		t.Errorf("modelToWorkflowInstance() LastRetryAt = %v, want %v", workflow.LastRetryAt, now)
	}
	if workflow.TraceID != "trace123" {
		t.Errorf("modelToWorkflowInstance() TraceID = %v, want %v", workflow.TraceID, "trace123")
	}
	if workflow.CorrelationID != "corr123" {
		t.Errorf("modelToWorkflowInstance() CorrelationID = %v, want %v", workflow.CorrelationID, "corr123")
	}
	if workflow.BusinessID != "biz123" {
		t.Errorf("modelToWorkflowInstance() BusinessID = %v, want %v", workflow.BusinessID, "biz123")
	}

	// Test JSONB fields
	if workflow.Input["key"] != "value" {
		t.Errorf("modelToWorkflowInstance() Input = %v, want %v", workflow.Input, map[string]interface{}{"key": "value"})
	}
	if workflow.Output["result"] != "success" {
		t.Errorf("modelToWorkflowInstance() Output = %v, want %v", workflow.Output, map[string]interface{}{"result": "success"})
	}
	if workflow.Context["ctx"] != "data" {
		t.Errorf("modelToWorkflowInstance() Context = %v, want %v", workflow.Context, map[string]interface{}{"ctx": "data"})
	}
	if workflow.Metadata["meta"] != "data" {
		t.Errorf("modelToWorkflowInstance() Metadata = %v, want %v", workflow.Metadata, map[string]interface{}{"meta": "data"})
	}

	// Test steps
	if len(workflow.Steps) != 1 {
		t.Errorf("modelToWorkflowInstance() Steps count = %v, want %v", len(workflow.Steps), 1)
	}
	if workflow.Steps[0].ID != "step1" {
		t.Errorf("modelToWorkflowInstance() Steps[0].ID = %v, want %v", workflow.Steps[0].ID, "step1")
	}
}

func TestStepInstanceToModel(t *testing.T) {
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

	model := stepInstanceToModel(step)

	if model.ID != "test-step" {
		t.Errorf("stepInstanceToModel() ID = %v, want %v", model.ID, "test-step")
	}
	if model.StepID != "step1" {
		t.Errorf("stepInstanceToModel() StepID = %v, want %v", model.StepID, "step1")
	}
	if model.WorkflowInstID != "test-workflow" {
		t.Errorf("stepInstanceToModel() WorkflowInstID = %v, want %v", model.WorkflowInstID, "test-workflow")
	}
	if model.Status != "completed" {
		t.Errorf("stepInstanceToModel() Status = %v, want %v", model.Status, "completed")
	}
	if model.StartedAt == nil || *model.StartedAt != now {
		t.Errorf("stepInstanceToModel() StartedAt = %v, want %v", model.StartedAt, now)
	}
	if model.CompletedAt == nil || *model.CompletedAt != now {
		t.Errorf("stepInstanceToModel() CompletedAt = %v, want %v", model.CompletedAt, now)
	}
	if model.Error == nil || *model.Error != "test error" {
		t.Errorf("stepInstanceToModel() Error = %v, want %v", model.Error, "test error")
	}
	if model.RetryCount != 2 {
		t.Errorf("stepInstanceToModel() RetryCount = %v, want %v", model.RetryCount, 2)
	}
	if model.LastRetryAt == nil || *model.LastRetryAt != now {
		t.Errorf("stepInstanceToModel() LastRetryAt = %v, want %v", model.LastRetryAt, now)
	}
	if model.DurationMs != 1000 {
		t.Errorf("stepInstanceToModel() DurationMs = %v, want %v", model.DurationMs, 1000)
	}
	if model.ExecutionOrder != 1 {
		t.Errorf("stepInstanceToModel() ExecutionOrder = %v, want %v", model.ExecutionOrder, 1)
	}
}

func TestModelToStepInstance(t *testing.T) {
	now := time.Now()
	inputJSON := JSONB{"key": "value"}
	outputJSON := JSONB{"result": "success"}

	model := &ORCHStepInstance{
		ID:             "test-step",
		StepID:         "step1",
		WorkflowInstID: "test-workflow",
		Status:         "completed",
		Input:          &inputJSON,
		Output:         &outputJSON,
		StartedAt:      &now,
		CompletedAt:    &now,
		Error:          stringPtr("test error"),
		RetryCount:     2,
		LastRetryAt:    &now,
		DurationMs:     1000,
		ExecutionOrder: 1,
	}

	step, err := modelToStepInstance(model)
	if err != nil {
		t.Errorf("modelToStepInstance() error = %v", err)
		return
	}

	if step.ID != "test-step" {
		t.Errorf("modelToStepInstance() ID = %v, want %v", step.ID, "test-step")
	}
	if step.StepID != "step1" {
		t.Errorf("modelToStepInstance() StepID = %v, want %v", step.StepID, "step1")
	}
	if step.WorkflowInstID != "test-workflow" {
		t.Errorf("modelToStepInstance() WorkflowInstID = %v, want %v", step.WorkflowInstID, "test-workflow")
	}
	if step.Status != StepStatusCompleted {
		t.Errorf("modelToStepInstance() Status = %v, want %v", step.Status, StepStatusCompleted)
	}
	if step.StartedAt == nil || *step.StartedAt != now {
		t.Errorf("modelToStepInstance() StartedAt = %v, want %v", step.StartedAt, now)
	}
	if step.CompletedAt == nil || *step.CompletedAt != now {
		t.Errorf("modelToStepInstance() CompletedAt = %v, want %v", step.CompletedAt, now)
	}
	if step.Error == nil || *step.Error != "test error" {
		t.Errorf("modelToStepInstance() Error = %v, want %v", step.Error, "test error")
	}
	if step.RetryCount != 2 {
		t.Errorf("modelToStepInstance() RetryCount = %v, want %v", step.RetryCount, 2)
	}
	if step.LastRetryAt == nil || *step.LastRetryAt != now {
		t.Errorf("modelToStepInstance() LastRetryAt = %v, want %v", step.LastRetryAt, now)
	}
	if step.DurationMs != 1000 {
		t.Errorf("modelToStepInstance() DurationMs = %v, want %v", step.DurationMs, 1000)
	}
	if step.ExecutionOrder != 1 {
		t.Errorf("modelToStepInstance() ExecutionOrder = %v, want %v", step.ExecutionOrder, 1)
	}

	// Test JSONB fields
	if step.Input["key"] != "value" {
		t.Errorf("modelToStepInstance() Input = %v, want %v", step.Input, map[string]interface{}{"key": "value"})
	}
	if step.Output["result"] != "success" {
		t.Errorf("modelToStepInstance() Output = %v, want %v", step.Output, map[string]interface{}{"result": "success"})
	}
}

func TestWorkflowEventToModel(t *testing.T) {
	now := time.Now()
	event := &WorkflowEvent{
		ID:             "test-event",
		WorkflowInstID: "test-workflow",
		StepInstID:     stringPtr("test-step"),
		EventType:      "test.event",
		EventData:      map[string]interface{}{"key": "value"},
		Timestamp:      now,
	}

	model := workflowEventToModel(event)

	if model.ID != "test-event" {
		t.Errorf("workflowEventToModel() ID = %v, want %v", model.ID, "test-event")
	}
	if model.WorkflowInstID != "test-workflow" {
		t.Errorf("workflowEventToModel() WorkflowInstID = %v, want %v", model.WorkflowInstID, "test-workflow")
	}
	if model.StepInstID == nil || *model.StepInstID != "test-step" {
		t.Errorf("workflowEventToModel() StepInstID = %v, want %v", model.StepInstID, "test-step")
	}
	if model.EventType != "test.event" {
		t.Errorf("workflowEventToModel() EventType = %v, want %v", model.EventType, "test.event")
	}
	if model.Timestamp != now {
		t.Errorf("workflowEventToModel() Timestamp = %v, want %v", model.Timestamp, now)
	}
}

func TestModelToWorkflowEvent(t *testing.T) {
	now := time.Now()
	eventDataJSON := JSONB{"key": "value"}

	model := &ORCHWorkflowEvent{
		ID:             "test-event",
		WorkflowInstID: "test-workflow",
		StepInstID:     stringPtr("test-step"),
		EventType:      "test.event",
		EventData:      &eventDataJSON,
		Timestamp:      now,
	}

	event, err := modelToWorkflowEvent(model)
	if err != nil {
		t.Errorf("modelToWorkflowEvent() error = %v", err)
		return
	}

	if event.ID != "test-event" {
		t.Errorf("modelToWorkflowEvent() ID = %v, want %v", event.ID, "test-event")
	}
	if event.WorkflowInstID != "test-workflow" {
		t.Errorf("modelToWorkflowEvent() WorkflowInstID = %v, want %v", event.WorkflowInstID, "test-workflow")
	}
	if event.StepInstID == nil || *event.StepInstID != "test-step" {
		t.Errorf("modelToWorkflowEvent() StepInstID = %v, want %v", event.StepInstID, "test-step")
	}
	if event.EventType != "test.event" {
		t.Errorf("modelToWorkflowEvent() EventType = %v, want %v", event.EventType, "test.event")
	}
	if event.Timestamp != now {
		t.Errorf("modelToWorkflowEvent() Timestamp = %v, want %v", event.Timestamp, now)
	}
	if event.EventData["key"] != "value" {
		t.Errorf("modelToWorkflowEvent() EventData = %v, want %v", event.EventData, map[string]interface{}{"key": "value"})
	}
}
