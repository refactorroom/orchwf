package orchwf

import (
	"context"
	"testing"
	"time"
)

func TestWorkflowBuilder_NewWorkflowBuilder(t *testing.T) {
	builder := NewWorkflowBuilder("test-workflow", "Test Workflow")

	if builder.workflow.ID != "test-workflow" {
		t.Errorf("NewWorkflowBuilder() ID = %v, want %v", builder.workflow.ID, "test-workflow")
	}
	if builder.workflow.Name != "Test Workflow" {
		t.Errorf("NewWorkflowBuilder() Name = %v, want %v", builder.workflow.Name, "Test Workflow")
	}
	if builder.workflow.Version != "1.0.0" {
		t.Errorf("NewWorkflowBuilder() Version = %v, want %v", builder.workflow.Version, "1.0.0")
	}
}

func TestWorkflowBuilder_WithDescription(t *testing.T) {
	builder := NewWorkflowBuilder("test-workflow", "Test Workflow")
	builder.WithDescription("Test description")

	if builder.workflow.Description != "Test description" {
		t.Errorf("WithDescription() = %v, want %v", builder.workflow.Description, "Test description")
	}
}

func TestWorkflowBuilder_WithVersion(t *testing.T) {
	builder := NewWorkflowBuilder("test-workflow", "Test Workflow")
	builder.WithVersion("2.0.0")

	if builder.workflow.Version != "2.0.0" {
		t.Errorf("WithVersion() = %v, want %v", builder.workflow.Version, "2.0.0")
	}
}

func TestWorkflowBuilder_WithMetadata(t *testing.T) {
	builder := NewWorkflowBuilder("test-workflow", "Test Workflow")
	builder.WithMetadata("key1", "value1")
	builder.WithMetadata("key2", 123)

	if builder.workflow.Metadata["key1"] != "value1" {
		t.Errorf("WithMetadata() key1 = %v, want %v", builder.workflow.Metadata["key1"], "value1")
	}
	if builder.workflow.Metadata["key2"] != 123 {
		t.Errorf("WithMetadata() key2 = %v, want %v", builder.workflow.Metadata["key2"], 123)
	}
}

func TestWorkflowBuilder_AddStep(t *testing.T) {
	builder := NewWorkflowBuilder("test-workflow", "Test Workflow")

	step := &StepDefinition{
		ID:   "step1",
		Name: "Step 1",
		Executor: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{"result": "success"}, nil
		},
	}

	builder.AddStep(step)

	if len(builder.workflow.Steps) != 1 {
		t.Errorf("AddStep() count = %v, want %v", len(builder.workflow.Steps), 1)
	}
	if builder.workflow.Steps[0].ID != "step1" {
		t.Errorf("AddStep() step ID = %v, want %v", builder.workflow.Steps[0].ID, "step1")
	}
}

func TestWorkflowBuilder_Build(t *testing.T) {
	tests := []struct {
		name    string
		builder func() *WorkflowBuilder
		wantErr bool
	}{
		{
			name: "valid workflow",
			builder: func() *WorkflowBuilder {
				step, _ := NewStepBuilder("step1", "Step 1", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{"result": "success"}, nil
				}).Build()

				return NewWorkflowBuilder("test-workflow", "Test Workflow").
					AddStep(step)
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			builder: func() *WorkflowBuilder {
				step, _ := NewStepBuilder("step1", "Step 1", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{"result": "success"}, nil
				}).Build()

				builder := NewWorkflowBuilder("", "Test Workflow")
				builder.workflow.ID = ""
				return builder.AddStep(step)
			},
			wantErr: true,
		},
		{
			name: "empty name",
			builder: func() *WorkflowBuilder {
				step, _ := NewStepBuilder("step1", "Step 1", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{"result": "success"}, nil
				}).Build()

				builder := NewWorkflowBuilder("test-workflow", "")
				return builder.AddStep(step)
			},
			wantErr: true,
		},
		{
			name: "no steps",
			builder: func() *WorkflowBuilder {
				return NewWorkflowBuilder("test-workflow", "Test Workflow")
			},
			wantErr: true,
		},
		{
			name: "invalid dependency",
			builder: func() *WorkflowBuilder {
				step, _ := NewStepBuilder("step1", "Step 1", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{"result": "success"}, nil
				}).WithDependencies("non-existing").Build()

				return NewWorkflowBuilder("test-workflow", "Test Workflow").
					AddStep(step)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow, err := tt.builder().Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && workflow == nil {
				t.Errorf("Build() returned nil workflow")
			}
		})
	}
}

func TestStepBuilder_NewStepBuilder(t *testing.T) {
	executor := func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "success"}, nil
	}

	builder := NewStepBuilder("test-step", "Test Step", executor)

	if builder.step.ID != "test-step" {
		t.Errorf("NewStepBuilder() ID = %v, want %v", builder.step.ID, "test-step")
	}
	if builder.step.Name != "Test Step" {
		t.Errorf("NewStepBuilder() Name = %v, want %v", builder.step.Name, "Test Step")
	}
	if builder.step.Executor == nil {
		t.Errorf("NewStepBuilder() Executor is nil")
	}
	if !builder.step.Required {
		t.Errorf("NewStepBuilder() Required = %v, want %v", builder.step.Required, true)
	}
	if builder.step.Async {
		t.Errorf("NewStepBuilder() Async = %v, want %v", builder.step.Async, false)
	}
}

func TestStepBuilder_WithDescription(t *testing.T) {
	executor := func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "success"}, nil
	}

	builder := NewStepBuilder("test-step", "Test Step", executor)
	builder.WithDescription("Test description")

	if builder.step.Description != "Test description" {
		t.Errorf("WithDescription() = %v, want %v", builder.step.Description, "Test description")
	}
}

func TestStepBuilder_WithDependencies(t *testing.T) {
	executor := func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "success"}, nil
	}

	builder := NewStepBuilder("test-step", "Test Step", executor)
	builder.WithDependencies("step1", "step2")

	if len(builder.step.Dependencies) != 2 {
		t.Errorf("WithDependencies() count = %v, want %v", len(builder.step.Dependencies), 2)
	}
	if builder.step.Dependencies[0] != "step1" {
		t.Errorf("WithDependencies() [0] = %v, want %v", builder.step.Dependencies[0], "step1")
	}
	if builder.step.Dependencies[1] != "step2" {
		t.Errorf("WithDependencies() [1] = %v, want %v", builder.step.Dependencies[1], "step2")
	}
}

func TestStepBuilder_WithCompensator(t *testing.T) {
	executor := func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "success"}, nil
	}

	compensator := func(ctx context.Context, input map[string]interface{}) error {
		return nil
	}

	builder := NewStepBuilder("test-step", "Test Step", executor)
	builder.WithCompensator(compensator)

	if builder.step.Compensator == nil {
		t.Errorf("WithCompensator() Compensator is nil")
	}
}

func TestStepBuilder_WithRetryPolicy(t *testing.T) {
	executor := func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "success"}, nil
	}

	policy := &RetryPolicy{
		MaxAttempts:     3,
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
	}

	builder := NewStepBuilder("test-step", "Test Step", executor)
	builder.WithRetryPolicy(policy)

	if builder.step.RetryPolicy != policy {
		t.Errorf("WithRetryPolicy() = %v, want %v", builder.step.RetryPolicy, policy)
	}
}

func TestStepBuilder_WithTimeout(t *testing.T) {
	executor := func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "success"}, nil
	}

	builder := NewStepBuilder("test-step", "Test Step", executor)
	builder.WithTimeout(5 * time.Second)

	if builder.step.Timeout != 5*time.Second {
		t.Errorf("WithTimeout() = %v, want %v", builder.step.Timeout, 5*time.Second)
	}
}

func TestStepBuilder_WithRequired(t *testing.T) {
	executor := func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "success"}, nil
	}

	builder := NewStepBuilder("test-step", "Test Step", executor)
	builder.WithRequired(false)

	if builder.step.Required {
		t.Errorf("WithRequired() = %v, want %v", builder.step.Required, false)
	}
}

func TestStepBuilder_WithAsync(t *testing.T) {
	executor := func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "success"}, nil
	}

	builder := NewStepBuilder("test-step", "Test Step", executor)
	builder.WithAsync(true)

	if !builder.step.Async {
		t.Errorf("WithAsync() = %v, want %v", builder.step.Async, true)
	}
}

func TestStepBuilder_Build(t *testing.T) {
	tests := []struct {
		name    string
		builder func() *StepBuilder
		wantErr bool
	}{
		{
			name: "valid step",
			builder: func() *StepBuilder {
				return NewStepBuilder("test-step", "Test Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{"result": "success"}, nil
				})
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			builder: func() *StepBuilder {
				builder := NewStepBuilder("", "Test Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{"result": "success"}, nil
				})
				builder.step.ID = ""
				return builder
			},
			wantErr: true,
		},
		{
			name: "empty name",
			builder: func() *StepBuilder {
				builder := NewStepBuilder("test-step", "", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{"result": "success"}, nil
				})
				return builder
			},
			wantErr: true,
		},
		{
			name: "nil executor",
			builder: func() *StepBuilder {
				builder := NewStepBuilder("test-step", "Test Step", func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
					return map[string]interface{}{"result": "success"}, nil
				})
				builder.step.Executor = nil
				return builder
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step, err := tt.builder().Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && step == nil {
				t.Errorf("Build() returned nil step")
			}
		})
	}
}

func TestRetryPolicyBuilder_NewRetryPolicyBuilder(t *testing.T) {
	builder := NewRetryPolicyBuilder()

	if builder.policy.MaxAttempts != 3 {
		t.Errorf("NewRetryPolicyBuilder() MaxAttempts = %v, want %v", builder.policy.MaxAttempts, 3)
	}
	if builder.policy.InitialInterval != 1*time.Second {
		t.Errorf("NewRetryPolicyBuilder() InitialInterval = %v, want %v", builder.policy.InitialInterval, 1*time.Second)
	}
	if builder.policy.MaxInterval != 30*time.Second {
		t.Errorf("NewRetryPolicyBuilder() MaxInterval = %v, want %v", builder.policy.MaxInterval, 30*time.Second)
	}
	if builder.policy.Multiplier != 2.0 {
		t.Errorf("NewRetryPolicyBuilder() Multiplier = %v, want %v", builder.policy.Multiplier, 2.0)
	}
}

func TestRetryPolicyBuilder_WithMaxAttempts(t *testing.T) {
	builder := NewRetryPolicyBuilder()
	builder.WithMaxAttempts(5)

	if builder.policy.MaxAttempts != 5 {
		t.Errorf("WithMaxAttempts() = %v, want %v", builder.policy.MaxAttempts, 5)
	}
}

func TestRetryPolicyBuilder_WithInitialInterval(t *testing.T) {
	builder := NewRetryPolicyBuilder()
	builder.WithInitialInterval(2 * time.Second)

	if builder.policy.InitialInterval != 2*time.Second {
		t.Errorf("WithInitialInterval() = %v, want %v", builder.policy.InitialInterval, 2*time.Second)
	}
}

func TestRetryPolicyBuilder_WithMaxInterval(t *testing.T) {
	builder := NewRetryPolicyBuilder()
	builder.WithMaxInterval(60 * time.Second)

	if builder.policy.MaxInterval != 60*time.Second {
		t.Errorf("WithMaxInterval() = %v, want %v", builder.policy.MaxInterval, 60*time.Second)
	}
}

func TestRetryPolicyBuilder_WithMultiplier(t *testing.T) {
	builder := NewRetryPolicyBuilder()
	builder.WithMultiplier(1.5)

	if builder.policy.Multiplier != 1.5 {
		t.Errorf("WithMultiplier() = %v, want %v", builder.policy.Multiplier, 1.5)
	}
}

func TestRetryPolicyBuilder_WithRetryableErrors(t *testing.T) {
	builder := NewRetryPolicyBuilder()
	builder.WithRetryableErrors("timeout", "connection_refused")

	if len(builder.policy.RetryableErrors) != 2 {
		t.Errorf("WithRetryableErrors() count = %v, want %v", len(builder.policy.RetryableErrors), 2)
	}
	if builder.policy.RetryableErrors[0] != "timeout" {
		t.Errorf("WithRetryableErrors() [0] = %v, want %v", builder.policy.RetryableErrors[0], "timeout")
	}
	if builder.policy.RetryableErrors[1] != "connection_refused" {
		t.Errorf("WithRetryableErrors() [1] = %v, want %v", builder.policy.RetryableErrors[1], "connection_refused")
	}
}

func TestRetryPolicyBuilder_Build(t *testing.T) {
	builder := NewRetryPolicyBuilder()
	policy := builder.Build()

	if policy == nil {
		t.Errorf("Build() returned nil policy")
	}
	if policy.MaxAttempts != 3 {
		t.Errorf("Build() MaxAttempts = %v, want %v", policy.MaxAttempts, 3)
	}
}
