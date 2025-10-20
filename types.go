package orchwf

import (
	"context"
	"encoding/json"
	"time"
)

// WorkflowStatus represents the current status of a workflow
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
	WorkflowStatusRetrying  WorkflowStatus = "retrying"
)

// StepStatus represents the current status of a workflow step
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
	StepStatusRetrying  StepStatus = "retrying"
)

// ExecutionMode defines how steps should be executed
type ExecutionMode string

const (
	ExecutionModeSequential ExecutionMode = "sequential" // Steps run one after another
	ExecutionModeParallel   ExecutionMode = "parallel"   // Steps run concurrently
)

// StepExecutor is a function that executes a single step
// It receives the context, input data, and returns output data or error
type StepExecutor func(ctx context.Context, input map[string]interface{}) (output map[string]interface{}, err error)

// StepCompensator is a function that compensates/rolls back a step on failure
type StepCompensator func(ctx context.Context, input map[string]interface{}) error

// WorkflowDefinition defines the structure of a workflow
type WorkflowDefinition struct {
	ID          string
	Name        string
	Description string
	Version     string
	Steps       []*StepDefinition
	Metadata    map[string]interface{}
}

// StepDefinition defines a single step in the workflow
type StepDefinition struct {
	ID           string
	Name         string
	Description  string
	Executor     StepExecutor
	Compensator  StepCompensator
	Dependencies []string // IDs of steps that must complete before this step
	RetryPolicy  *RetryPolicy
	Timeout      time.Duration
	Required     bool // If false, failure won't stop the workflow
	Async        bool // If true, step runs asynchronously
}

// RetryPolicy defines retry behavior for a step
type RetryPolicy struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RetryableErrors []string // Specific error patterns that should trigger retry
}

// WorkflowInstance represents a running instance of a workflow
type WorkflowInstance struct {
	ID             string
	WorkflowID     string
	Status         WorkflowStatus
	Input          map[string]interface{}
	Output         map[string]interface{}
	Context        map[string]interface{}
	CurrentStepID  string
	Steps          []*StepInstance
	StartedAt      time.Time
	CompletedAt    *time.Time
	Error          *string
	RetryCount     int
	LastRetryAt    *time.Time
	Metadata       map[string]interface{}
	TraceID        string
	CorrelationID  string
	BusinessID     string
}

// StepInstance represents a running instance of a step
type StepInstance struct {
	ID              string
	StepID          string
	WorkflowInstID  string
	Status          StepStatus
	Input           map[string]interface{}
	Output          map[string]interface{}
	StartedAt       *time.Time
	CompletedAt     *time.Time
	Error           *string
	RetryCount      int
	LastRetryAt     *time.Time
	DurationMs      int64
	ExecutionOrder  int
}

// WorkflowEvent represents an event in the workflow lifecycle
type WorkflowEvent struct {
	ID             string
	WorkflowInstID string
	StepInstID     *string
	EventType      string
	EventData      map[string]interface{}
	Timestamp      time.Time
}

// StepResult represents the result of a step execution
type StepResult struct {
	Success  bool
	Output   map[string]interface{}
	Error    error
	Duration time.Duration
}

// WorkflowResult represents the final result of a workflow execution
type WorkflowResult struct {
	Success      bool
	WorkflowInst *WorkflowInstance
	Output       map[string]interface{}
	Error        error
	Duration     time.Duration
}

// Helper methods

// SetInput sets the input for a workflow instance
func (w *WorkflowInstance) SetInput(input interface{}) error {
	if input == nil {
		w.Input = make(map[string]interface{})
		return nil
	}

	// If already a map, use directly
	if m, ok := input.(map[string]interface{}); ok {
		w.Input = m
		return nil
	}

	// Otherwise, marshal and unmarshal to convert
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	w.Input = m
	return nil
}

// SetOutput sets the output for a workflow instance
func (w *WorkflowInstance) SetOutput(output interface{}) error {
	if output == nil {
		w.Output = make(map[string]interface{})
		return nil
	}

	// If already a map, use directly
	if m, ok := output.(map[string]interface{}); ok {
		w.Output = m
		return nil
	}

	// Otherwise, marshal and unmarshal to convert
	data, err := json.Marshal(output)
	if err != nil {
		return err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	w.Output = m
	return nil
}

// GetContext gets a value from workflow context
func (w *WorkflowInstance) GetContext(key string) interface{} {
	if w.Context == nil {
		return nil
	}
	return w.Context[key]
}

// SetContext sets a value in workflow context
func (w *WorkflowInstance) SetContext(key string, value interface{}) {
	if w.Context == nil {
		w.Context = make(map[string]interface{})
	}
	w.Context[key] = value
}

// IsCompleted checks if the workflow is in a terminal state
func (w *WorkflowInstance) IsCompleted() bool {
	return w.Status == WorkflowStatusCompleted ||
		w.Status == WorkflowStatusFailed ||
		w.Status == WorkflowStatusCancelled
}

// CanRetry checks if the workflow can be retried
func (w *WorkflowInstance) CanRetry(maxRetries int) bool {
	return w.Status == WorkflowStatusFailed && w.RetryCount < maxRetries
}

// IsCompleted checks if the step is in a terminal state
func (s *StepInstance) IsCompleted() bool {
	return s.Status == StepStatusCompleted ||
		s.Status == StepStatusFailed ||
		s.Status == StepStatusSkipped
}

// CanRetry checks if the step can be retried
func (s *StepInstance) CanRetry(policy *RetryPolicy) bool {
	if policy == nil {
		return false
	}
	return s.Status == StepStatusFailed && s.RetryCount < policy.MaxAttempts
}
