package orchwf

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StateManager handles persistence of workflow and step states
type StateManager interface {
	// Workflow operations
	SaveWorkflow(ctx context.Context, workflow *WorkflowInstance) error
	GetWorkflow(ctx context.Context, workflowInstID string) (*WorkflowInstance, error)
	UpdateWorkflowStatus(ctx context.Context, workflowInstID string, status WorkflowStatus) error
	UpdateWorkflowOutput(ctx context.Context, workflowInstID string, output map[string]interface{}) error
	UpdateWorkflowError(ctx context.Context, workflowInstID string, err error) error
	ListWorkflows(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*WorkflowInstance, int64, error)

	// Step operations
	SaveStep(ctx context.Context, step *StepInstance) error
	GetStep(ctx context.Context, stepInstID string) (*StepInstance, error)
	GetWorkflowSteps(ctx context.Context, workflowInstID string) ([]*StepInstance, error)
	UpdateStepStatus(ctx context.Context, stepInstID string, status StepStatus) error
	UpdateStepOutput(ctx context.Context, stepInstID string, output map[string]interface{}) error
	UpdateStepError(ctx context.Context, stepInstID string, err error) error

	// Event operations
	SaveEvent(ctx context.Context, event *WorkflowEvent) error
	GetWorkflowEvents(ctx context.Context, workflowInstID string) ([]*WorkflowEvent, error)

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// InMemoryStateManager implements StateManager using in-memory storage
type InMemoryStateManager struct {
	workflows map[string]*WorkflowInstance
	steps     map[string]*StepInstance
	events    map[string]*WorkflowEvent
	mu        sync.RWMutex
}

// NewInMemoryStateManager creates a new in-memory state manager
func NewInMemoryStateManager() *InMemoryStateManager {
	return &InMemoryStateManager{
		workflows: make(map[string]*WorkflowInstance),
		steps:     make(map[string]*StepInstance),
		events:    make(map[string]*WorkflowEvent),
	}
}

// SaveWorkflow saves a workflow instance to memory
func (m *InMemoryStateManager) SaveWorkflow(ctx context.Context, workflow *WorkflowInstance) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Deep copy to avoid race conditions
	workflowCopy := m.deepCopyWorkflow(workflow)
	m.workflows[workflow.ID] = workflowCopy
	return nil
}

// GetWorkflow retrieves a workflow instance by ID
func (m *InMemoryStateManager) GetWorkflow(ctx context.Context, workflowInstID string) (*WorkflowInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	workflow, ok := m.workflows[workflowInstID]
	if !ok {
		return nil, fmt.Errorf("workflow not found: %s", workflowInstID)
	}

	// Deep copy to avoid race conditions
	return m.deepCopyWorkflow(workflow), nil
}

// UpdateWorkflowStatus updates the status of a workflow
func (m *InMemoryStateManager) UpdateWorkflowStatus(ctx context.Context, workflowInstID string, status WorkflowStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	workflow, ok := m.workflows[workflowInstID]
	if !ok {
		return fmt.Errorf("workflow not found: %s", workflowInstID)
	}

	workflow.Status = status
	if status == WorkflowStatusCompleted || status == WorkflowStatusFailed || status == WorkflowStatusCancelled {
		now := time.Now()
		workflow.CompletedAt = &now
	}

	return nil
}

// UpdateWorkflowOutput updates the output of a workflow
func (m *InMemoryStateManager) UpdateWorkflowOutput(ctx context.Context, workflowInstID string, output map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	workflow, ok := m.workflows[workflowInstID]
	if !ok {
		return fmt.Errorf("workflow not found: %s", workflowInstID)
	}

	workflow.Output = make(map[string]interface{})
	for k, v := range output {
		workflow.Output[k] = v
	}

	return nil
}

// UpdateWorkflowError updates the error of a workflow
func (m *InMemoryStateManager) UpdateWorkflowError(ctx context.Context, workflowInstID string, err error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	workflow, ok := m.workflows[workflowInstID]
	if !ok {
		return fmt.Errorf("workflow not found: %s", workflowInstID)
	}

	errorMsg := err.Error()
	workflow.Error = &errorMsg
	workflow.Status = WorkflowStatusFailed
	now := time.Now()
	workflow.CompletedAt = &now

	return nil
}

// ListWorkflows lists workflows with optional filters
func (m *InMemoryStateManager) ListWorkflows(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*WorkflowInstance, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*WorkflowInstance
	for _, workflow := range m.workflows {
		// Apply filters
		matches := true
		for key, value := range filters {
			switch key {
			case "workflow_id":
				if workflow.WorkflowID != value {
					matches = false
				}
			case "status":
				if workflow.Status != value {
					matches = false
				}
			case "trace_id":
				if workflow.TraceID != value {
					matches = false
				}
			case "correlation_id":
				if workflow.CorrelationID != value {
					matches = false
				}
			case "business_id":
				if workflow.BusinessID != value {
					matches = false
				}
			}
			if !matches {
				break
			}
		}

		if matches {
			results = append(results, m.deepCopyWorkflow(workflow))
		}
	}

	total := int64(len(results))

	// Apply pagination
	if offset >= len(results) {
		return []*WorkflowInstance{}, total, nil
	}

	end := offset + limit
	if end > len(results) {
		end = len(results)
	}

	return results[offset:end], total, nil
}

// SaveStep saves a step instance to memory
func (m *InMemoryStateManager) SaveStep(ctx context.Context, step *StepInstance) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Deep copy to avoid race conditions
	stepCopy := m.deepCopyStep(step)
	m.steps[step.ID] = stepCopy
	return nil
}

// GetStep retrieves a step instance by ID
func (m *InMemoryStateManager) GetStep(ctx context.Context, stepInstID string) (*StepInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	step, ok := m.steps[stepInstID]
	if !ok {
		return nil, fmt.Errorf("step not found: %s", stepInstID)
	}

	// Deep copy to avoid race conditions
	return m.deepCopyStep(step), nil
}

// GetWorkflowSteps retrieves all steps for a workflow
func (m *InMemoryStateManager) GetWorkflowSteps(ctx context.Context, workflowInstID string) ([]*StepInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var steps []*StepInstance
	for _, step := range m.steps {
		if step.WorkflowInstID == workflowInstID {
			steps = append(steps, m.deepCopyStep(step))
		}
	}

	return steps, nil
}

// UpdateStepStatus updates the status of a step
func (m *InMemoryStateManager) UpdateStepStatus(ctx context.Context, stepInstID string, status StepStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	step, ok := m.steps[stepInstID]
	if !ok {
		return fmt.Errorf("step not found: %s", stepInstID)
	}

	step.Status = status
	if status == StepStatusRunning {
		now := time.Now()
		step.StartedAt = &now
	}

	if status == StepStatusCompleted || status == StepStatusFailed || status == StepStatusSkipped {
		now := time.Now()
		step.CompletedAt = &now
	}

	return nil
}

// UpdateStepOutput updates the output of a step
func (m *InMemoryStateManager) UpdateStepOutput(ctx context.Context, stepInstID string, output map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	step, ok := m.steps[stepInstID]
	if !ok {
		return fmt.Errorf("step not found: %s", stepInstID)
	}

	step.Output = make(map[string]interface{})
	for k, v := range output {
		step.Output[k] = v
	}

	return nil
}

// UpdateStepError updates the error of a step
func (m *InMemoryStateManager) UpdateStepError(ctx context.Context, stepInstID string, err error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	step, ok := m.steps[stepInstID]
	if !ok {
		return fmt.Errorf("step not found: %s", stepInstID)
	}

	errorMsg := err.Error()
	step.Error = &errorMsg
	step.Status = StepStatusFailed
	now := time.Now()
	step.CompletedAt = &now

	return nil
}

// SaveEvent saves a workflow event to memory
func (m *InMemoryStateManager) SaveEvent(ctx context.Context, event *WorkflowEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Deep copy to avoid race conditions
	eventCopy := m.deepCopyEvent(event)
	m.events[event.ID] = eventCopy
	return nil
}

// GetWorkflowEvents retrieves all events for a workflow
func (m *InMemoryStateManager) GetWorkflowEvents(ctx context.Context, workflowInstID string) ([]*WorkflowEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var events []*WorkflowEvent
	for _, event := range m.events {
		if event.WorkflowInstID == workflowInstID {
			events = append(events, m.deepCopyEvent(event))
		}
	}

	return events, nil
}

// WithTransaction executes a function within a transaction (no-op for in-memory)
func (m *InMemoryStateManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// For in-memory, we just execute the function
	// The mutex provides the necessary synchronization
	return fn(ctx)
}

// Helper methods for deep copying

func (m *InMemoryStateManager) deepCopyWorkflow(w *WorkflowInstance) *WorkflowInstance {
	copy := &WorkflowInstance{
		ID:            w.ID,
		WorkflowID:    w.WorkflowID,
		Status:        w.Status,
		CurrentStepID: w.CurrentStepID,
		StartedAt:     w.StartedAt,
		RetryCount:    w.RetryCount,
		TraceID:       w.TraceID,
		CorrelationID: w.CorrelationID,
		BusinessID:    w.BusinessID,
	}

	// Copy pointers
	if w.CompletedAt != nil {
		completedAt := *w.CompletedAt
		copy.CompletedAt = &completedAt
	}
	if w.Error != nil {
		error := *w.Error
		copy.Error = &error
	}
	if w.LastRetryAt != nil {
		lastRetryAt := *w.LastRetryAt
		copy.LastRetryAt = &lastRetryAt
	}

	// Copy maps
	copy.Input = make(map[string]interface{})
	for k, v := range w.Input {
		copy.Input[k] = v
	}
	copy.Output = make(map[string]interface{})
	for k, v := range w.Output {
		copy.Output[k] = v
	}
	copy.Context = make(map[string]interface{})
	for k, v := range w.Context {
		copy.Context[k] = v
	}
	copy.Metadata = make(map[string]interface{})
	for k, v := range w.Metadata {
		copy.Metadata[k] = v
	}

	// Copy steps
	copy.Steps = make([]*StepInstance, len(w.Steps))
	for i, step := range w.Steps {
		copy.Steps[i] = m.deepCopyStep(step)
	}

	return copy
}

func (m *InMemoryStateManager) deepCopyStep(s *StepInstance) *StepInstance {
	copy := &StepInstance{
		ID:             s.ID,
		StepID:         s.StepID,
		WorkflowInstID: s.WorkflowInstID,
		Status:         s.Status,
		RetryCount:     s.RetryCount,
		DurationMs:     s.DurationMs,
		ExecutionOrder: s.ExecutionOrder,
	}

	// Copy pointers
	if s.StartedAt != nil {
		startedAt := *s.StartedAt
		copy.StartedAt = &startedAt
	}
	if s.CompletedAt != nil {
		completedAt := *s.CompletedAt
		copy.CompletedAt = &completedAt
	}
	if s.Error != nil {
		error := *s.Error
		copy.Error = &error
	}
	if s.LastRetryAt != nil {
		lastRetryAt := *s.LastRetryAt
		copy.LastRetryAt = &lastRetryAt
	}

	// Copy maps
	copy.Input = make(map[string]interface{})
	for k, v := range s.Input {
		copy.Input[k] = v
	}
	copy.Output = make(map[string]interface{})
	for k, v := range s.Output {
		copy.Output[k] = v
	}

	return copy
}

func (m *InMemoryStateManager) deepCopyEvent(e *WorkflowEvent) *WorkflowEvent {
	copy := &WorkflowEvent{
		ID:             e.ID,
		WorkflowInstID: e.WorkflowInstID,
		EventType:      e.EventType,
		Timestamp:      e.Timestamp,
	}

	// Copy pointer
	if e.StepInstID != nil {
		stepInstID := *e.StepInstID
		copy.StepInstID = &stepInstID
	}

	// Copy map
	copy.EventData = make(map[string]interface{})
	for k, v := range e.EventData {
		copy.EventData[k] = v
	}

	return copy
}
