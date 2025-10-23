package orchwf

import (
	"fmt"
	"time"
)

// WorkflowBuilder helps build workflow definitions
type WorkflowBuilder struct {
	workflow *WorkflowDefinition
}

// NewWorkflowBuilder creates a new workflow builder
func NewWorkflowBuilder(id, name string) *WorkflowBuilder {
	return &WorkflowBuilder{
		workflow: &WorkflowDefinition{
			ID:       id,
			Name:     name,
			Version:  "1.0.0",
			Steps:    make([]*StepDefinition, 0),
			Metadata: make(map[string]interface{}),
		},
	}
}

// WithDescription sets the workflow description
func (b *WorkflowBuilder) WithDescription(description string) *WorkflowBuilder {
	b.workflow.Description = description
	return b
}

// WithVersion sets the workflow version
func (b *WorkflowBuilder) WithVersion(version string) *WorkflowBuilder {
	b.workflow.Version = version
	return b
}

// WithMetadata adds metadata to the workflow
func (b *WorkflowBuilder) WithMetadata(key string, value interface{}) *WorkflowBuilder {
	b.workflow.Metadata[key] = value
	return b
}

// AddStep adds a step to the workflow
func (b *WorkflowBuilder) AddStep(step *StepDefinition) *WorkflowBuilder {
	b.workflow.Steps = append(b.workflow.Steps, step)
	return b
}

// Build returns the workflow definition
func (b *WorkflowBuilder) Build() (*WorkflowDefinition, error) {
	if b.workflow.ID == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}
	if b.workflow.Name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}
	if len(b.workflow.Steps) == 0 {
		return nil, fmt.Errorf("workflow must have at least one step")
	}

	// Validate dependencies
	stepIDs := make(map[string]bool)
	for _, step := range b.workflow.Steps {
		stepIDs[step.ID] = true
	}

	for _, step := range b.workflow.Steps {
		for _, dep := range step.Dependencies {
			if !stepIDs[dep] {
				return nil, fmt.Errorf("step %s has invalid dependency: %s", step.ID, dep)
			}
		}
	}

	return b.workflow, nil
}

// StepBuilder helps build step definitions
type StepBuilder struct {
	step *StepDefinition
}

// NewStepBuilder creates a new step builder
func NewStepBuilder(id, name string, executor StepExecutor) *StepBuilder {
	return &StepBuilder{
		step: &StepDefinition{
			ID:           id,
			Name:         name,
			Executor:     executor,
			Dependencies: make([]string, 0),
			Required:     true,
			Async:        false,
		},
	}
}

// WithDescription sets the step description
func (b *StepBuilder) WithDescription(description string) *StepBuilder {
	b.step.Description = description
	return b
}

// WithDependencies sets the step dependencies
func (b *StepBuilder) WithDependencies(dependencies ...string) *StepBuilder {
	b.step.Dependencies = dependencies
	return b
}

// WithCompensator sets the step compensator
func (b *StepBuilder) WithCompensator(compensator StepCompensator) *StepBuilder {
	b.step.Compensator = compensator
	return b
}

// WithRetryPolicy sets the retry policy
func (b *StepBuilder) WithRetryPolicy(policy *RetryPolicy) *StepBuilder {
	b.step.RetryPolicy = policy
	return b
}

// WithTimeout sets the step timeout
func (b *StepBuilder) WithTimeout(timeout time.Duration) *StepBuilder {
	b.step.Timeout = timeout
	return b
}

// WithRequired sets whether the step is required
func (b *StepBuilder) WithRequired(required bool) *StepBuilder {
	b.step.Required = required
	return b
}

// WithAsync sets whether the step is async
func (b *StepBuilder) WithAsync(async bool) *StepBuilder {
	b.step.Async = async
	return b
}

// WithPriority sets the step priority (higher number = higher priority)
func (b *StepBuilder) WithPriority(priority int) *StepBuilder {
	b.step.Priority = priority
	return b
}

// Build returns the step definition
func (b *StepBuilder) Build() (*StepDefinition, error) {
	if b.step.ID == "" {
		return nil, fmt.Errorf("step ID is required")
	}
	if b.step.Name == "" {
		return nil, fmt.Errorf("step name is required")
	}
	if b.step.Executor == nil {
		return nil, fmt.Errorf("step executor is required")
	}

	return b.step, nil
}

// RetryPolicyBuilder helps build retry policies
type RetryPolicyBuilder struct {
	policy *RetryPolicy
}

// NewRetryPolicyBuilder creates a new retry policy builder
func NewRetryPolicyBuilder() *RetryPolicyBuilder {
	return &RetryPolicyBuilder{
		policy: &RetryPolicy{
			MaxAttempts:     3,
			InitialInterval: 1 * time.Second,
			MaxInterval:     30 * time.Second,
			Multiplier:      2.0,
			RetryableErrors: make([]string, 0),
		},
	}
}

// WithMaxAttempts sets the maximum number of retry attempts
func (b *RetryPolicyBuilder) WithMaxAttempts(attempts int) *RetryPolicyBuilder {
	b.policy.MaxAttempts = attempts
	return b
}

// WithInitialInterval sets the initial retry interval
func (b *RetryPolicyBuilder) WithInitialInterval(interval time.Duration) *RetryPolicyBuilder {
	b.policy.InitialInterval = interval
	return b
}

// WithMaxInterval sets the maximum retry interval
func (b *RetryPolicyBuilder) WithMaxInterval(interval time.Duration) *RetryPolicyBuilder {
	b.policy.MaxInterval = interval
	return b
}

// WithMultiplier sets the backoff multiplier
func (b *RetryPolicyBuilder) WithMultiplier(multiplier float64) *RetryPolicyBuilder {
	b.policy.Multiplier = multiplier
	return b
}

// WithRetryableErrors sets specific errors that should trigger retry
func (b *RetryPolicyBuilder) WithRetryableErrors(errors ...string) *RetryPolicyBuilder {
	b.policy.RetryableErrors = errors
	return b
}

// Build returns the retry policy
func (b *RetryPolicyBuilder) Build() *RetryPolicy {
	return b.policy
}
