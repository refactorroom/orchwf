package orchwf

import (
	"testing"
)

func TestWorkflowInstance_SetInput(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: make(map[string]interface{}),
			wantErr:  false,
		},
		{
			name:     "map input",
			input:    map[string]interface{}{"key": "value"},
			expected: map[string]interface{}{"key": "value"},
			wantErr:  false,
		},
		{
			name:     "struct input",
			input:    struct{ Name string }{"test"},
			expected: map[string]interface{}{"Name": "test"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WorkflowInstance{}
			err := w.SetInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !mapsEqual(w.Input, tt.expected) {
				t.Errorf("SetInput() = %v, want %v", w.Input, tt.expected)
			}
		})
	}
}

func TestWorkflowInstance_SetOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   interface{}
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "nil output",
			output:   nil,
			expected: make(map[string]interface{}),
			wantErr:  false,
		},
		{
			name:     "map output",
			output:   map[string]interface{}{"result": "success"},
			expected: map[string]interface{}{"result": "success"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WorkflowInstance{}
			err := w.SetOutput(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !mapsEqual(w.Output, tt.expected) {
				t.Errorf("SetOutput() = %v, want %v", w.Output, tt.expected)
			}
		})
	}
}

func TestWorkflowInstance_GetContext(t *testing.T) {
	w := &WorkflowInstance{
		Context: map[string]interface{}{"key": "value"},
	}

	tests := []struct {
		name     string
		key      string
		expected interface{}
	}{
		{"existing key", "key", "value"},
		{"non-existing key", "missing", nil},
		{"nil context", "any", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "nil context" {
				w.Context = nil
			}
			if got := w.GetContext(tt.key); got != tt.expected {
				t.Errorf("GetContext() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWorkflowInstance_SetContext(t *testing.T) {
	w := &WorkflowInstance{}
	w.SetContext("key", "value")

	if w.Context["key"] != "value" {
		t.Errorf("SetContext() failed to set value")
	}

	// Test with nil context
	w.Context = nil
	w.SetContext("key2", "value2")
	if w.Context["key2"] != "value2" {
		t.Errorf("SetContext() failed to initialize context")
	}
}

func TestWorkflowInstance_IsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		status   WorkflowStatus
		expected bool
	}{
		{"pending", WorkflowStatusPending, false},
		{"running", WorkflowStatusRunning, false},
		{"completed", WorkflowStatusCompleted, true},
		{"failed", WorkflowStatusFailed, true},
		{"cancelled", WorkflowStatusCancelled, true},
		{"retrying", WorkflowStatusRetrying, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WorkflowInstance{Status: tt.status}
			if got := w.IsCompleted(); got != tt.expected {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWorkflowInstance_CanRetry(t *testing.T) {
	tests := []struct {
		name       string
		status     WorkflowStatus
		retryCount int
		maxRetries int
		expected   bool
	}{
		{"failed with retries left", WorkflowStatusFailed, 1, 3, true},
		{"failed no retries left", WorkflowStatusFailed, 3, 3, false},
		{"completed", WorkflowStatusCompleted, 0, 3, false},
		{"running", WorkflowStatusRunning, 0, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WorkflowInstance{Status: tt.status, RetryCount: tt.retryCount}
			if got := w.CanRetry(tt.maxRetries); got != tt.expected {
				t.Errorf("CanRetry() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStepInstance_IsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		status   StepStatus
		expected bool
	}{
		{"pending", StepStatusPending, false},
		{"running", StepStatusRunning, false},
		{"completed", StepStatusCompleted, true},
		{"failed", StepStatusFailed, true},
		{"skipped", StepStatusSkipped, true},
		{"retrying", StepStatusRetrying, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StepInstance{Status: tt.status}
			if got := s.IsCompleted(); got != tt.expected {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStepInstance_CanRetry(t *testing.T) {
	tests := []struct {
		name       string
		status     StepStatus
		retryCount int
		policy     *RetryPolicy
		expected   bool
	}{
		{"failed with retries left", StepStatusFailed, 1, &RetryPolicy{MaxAttempts: 3}, true},
		{"failed no retries left", StepStatusFailed, 3, &RetryPolicy{MaxAttempts: 3}, false},
		{"completed", StepStatusCompleted, 0, &RetryPolicy{MaxAttempts: 3}, false},
		{"nil policy", StepStatusFailed, 0, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StepInstance{Status: tt.status, RetryCount: tt.retryCount}
			if got := s.CanRetry(tt.policy); got != tt.expected {
				t.Errorf("CanRetry() = %v, want %v", got, tt.expected)
			}
		})
	}
}
