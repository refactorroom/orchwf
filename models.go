package orchwf

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Database models for standard SQL

// ORCHWorkflowInstance represents the database model for workflow instances
type ORCHWorkflowInstance struct {
	ID            string
	WorkflowID    string
	Status        string
	Input         *JSONB
	Output        *JSONB
	Context       *JSONB
	CurrentStepID *string
	StartedAt     time.Time
	CompletedAt   *time.Time
	Error         *string
	RetryCount    int
	LastRetryAt   *time.Time
	Metadata      *JSONB
	TraceID       string
	CorrelationID string
	BusinessID    string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Steps         []ORCHStepInstance
}

// ORCHStepInstance represents the database model for step instances
type ORCHStepInstance struct {
	ID             string
	StepID         string
	WorkflowInstID string
	Status         string
	Input          *JSONB
	Output         *JSONB
	StartedAt      *time.Time
	CompletedAt    *time.Time
	Error          *string
	RetryCount     int
	LastRetryAt    *time.Time
	DurationMs     int64
	ExecutionOrder int
	Priority       int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ORCHWorkflowEvent represents the database model for workflow events
type ORCHWorkflowEvent struct {
	ID             string
	WorkflowInstID string
	StepInstID     *string
	EventType      string
	EventData      *JSONB
	Timestamp      time.Time
	CreatedAt      time.Time
}

// JSONB represents a JSONB field for standard SQL
type JSONB map[string]interface{}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}
}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Conversion functions between domain types and database models

func workflowInstanceToModel(w *WorkflowInstance) *ORCHWorkflowInstance {
	model := &ORCHWorkflowInstance{
		ID:            w.ID,
		WorkflowID:    w.WorkflowID,
		Status:        string(w.Status),
		StartedAt:     w.StartedAt,
		CompletedAt:   w.CompletedAt,
		Error:         w.Error,
		RetryCount:    w.RetryCount,
		LastRetryAt:   w.LastRetryAt,
		TraceID:       w.TraceID,
		CorrelationID: w.CorrelationID,
		BusinessID:    w.BusinessID,
	}

	if w.CurrentStepID != "" {
		model.CurrentStepID = &w.CurrentStepID
	}

	// Convert JSONB fields
	if w.Input != nil {
		jsonb := JSONB(w.Input)
		model.Input = &jsonb
	}
	if w.Output != nil {
		jsonb := JSONB(w.Output)
		model.Output = &jsonb
	}
	if w.Context != nil {
		jsonb := JSONB(w.Context)
		model.Context = &jsonb
	}
	if w.Metadata != nil {
		jsonb := JSONB(w.Metadata)
		model.Metadata = &jsonb
	}

	return model
}

func modelToWorkflowInstance(m *ORCHWorkflowInstance) (*WorkflowInstance, error) {
	w := &WorkflowInstance{
		ID:            m.ID,
		WorkflowID:    m.WorkflowID,
		Status:        WorkflowStatus(m.Status),
		StartedAt:     m.StartedAt,
		CompletedAt:   m.CompletedAt,
		Error:         m.Error,
		RetryCount:    m.RetryCount,
		LastRetryAt:   m.LastRetryAt,
		TraceID:       m.TraceID,
		CorrelationID: m.CorrelationID,
		BusinessID:    m.BusinessID,
	}

	if m.CurrentStepID != nil {
		w.CurrentStepID = *m.CurrentStepID
	}

	// Convert JSONB fields
	if m.Input != nil {
		w.Input = map[string]interface{}(*m.Input)
	} else {
		w.Input = make(map[string]interface{})
	}

	if m.Output != nil {
		w.Output = map[string]interface{}(*m.Output)
	} else {
		w.Output = make(map[string]interface{})
	}

	if m.Context != nil {
		w.Context = map[string]interface{}(*m.Context)
	} else {
		w.Context = make(map[string]interface{})
	}

	if m.Metadata != nil {
		w.Metadata = map[string]interface{}(*m.Metadata)
	} else {
		w.Metadata = make(map[string]interface{})
	}

	// Convert steps
	if len(m.Steps) > 0 {
		w.Steps = make([]*StepInstance, 0, len(m.Steps))
		for _, stepModel := range m.Steps {
			step, err := modelToStepInstance(&stepModel)
			if err != nil {
				return nil, err
			}
			w.Steps = append(w.Steps, step)
		}
	} else {
		w.Steps = make([]*StepInstance, 0)
	}

	return w, nil
}

func stepInstanceToModel(s *StepInstance) *ORCHStepInstance {
	model := &ORCHStepInstance{
		ID:             s.ID,
		StepID:         s.StepID,
		WorkflowInstID: s.WorkflowInstID,
		Status:         string(s.Status),
		StartedAt:      s.StartedAt,
		CompletedAt:    s.CompletedAt,
		Error:          s.Error,
		RetryCount:     s.RetryCount,
		LastRetryAt:    s.LastRetryAt,
		DurationMs:     s.DurationMs,
		ExecutionOrder: s.ExecutionOrder,
	}

	// Convert JSONB fields
	if s.Input != nil {
		jsonb := JSONB(s.Input)
		model.Input = &jsonb
	}
	if s.Output != nil {
		jsonb := JSONB(s.Output)
		model.Output = &jsonb
	}

	return model
}

func modelToStepInstance(m *ORCHStepInstance) (*StepInstance, error) {
	s := &StepInstance{
		ID:             m.ID,
		StepID:         m.StepID,
		WorkflowInstID: m.WorkflowInstID,
		Status:         StepStatus(m.Status),
		StartedAt:      m.StartedAt,
		CompletedAt:    m.CompletedAt,
		Error:          m.Error,
		RetryCount:     m.RetryCount,
		LastRetryAt:    m.LastRetryAt,
		DurationMs:     m.DurationMs,
		ExecutionOrder: m.ExecutionOrder,
	}

	// Convert JSONB fields
	if m.Input != nil {
		s.Input = map[string]interface{}(*m.Input)
	} else {
		s.Input = make(map[string]interface{})
	}

	if m.Output != nil {
		s.Output = map[string]interface{}(*m.Output)
	} else {
		s.Output = make(map[string]interface{})
	}

	return s, nil
}

func workflowEventToModel(e *WorkflowEvent) *ORCHWorkflowEvent {
	model := &ORCHWorkflowEvent{
		ID:             e.ID,
		WorkflowInstID: e.WorkflowInstID,
		StepInstID:     e.StepInstID,
		EventType:      e.EventType,
		Timestamp:      e.Timestamp,
	}

	// Convert JSONB field
	if e.EventData != nil {
		jsonb := JSONB(e.EventData)
		model.EventData = &jsonb
	}

	return model
}

func modelToWorkflowEvent(m *ORCHWorkflowEvent) (*WorkflowEvent, error) {
	e := &WorkflowEvent{
		ID:             m.ID,
		WorkflowInstID: m.WorkflowInstID,
		StepInstID:     m.StepInstID,
		EventType:      m.EventType,
		Timestamp:      m.Timestamp,
	}

	// Convert JSONB field
	if m.EventData != nil {
		e.EventData = map[string]interface{}(*m.EventData)
	} else {
		e.EventData = make(map[string]interface{})
	}

	return e, nil
}
