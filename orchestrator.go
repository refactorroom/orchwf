package orchwf

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Orchestrator manages workflow execution with both sync and async support
type Orchestrator struct {
	stateManager StateManager
	workflows    map[string]*WorkflowDefinition
	mu           sync.RWMutex
	asyncWorkers int // Number of goroutines for async execution
}

// NewOrchestrator creates a new workflow orchestrator
func NewOrchestrator(stateManager StateManager) *Orchestrator {
	return &Orchestrator{
		stateManager: stateManager,
		workflows:    make(map[string]*WorkflowDefinition),
		asyncWorkers: 10, // Default number of async workers
	}
}

// NewOrchestratorWithAsyncWorkers creates a new workflow orchestrator with custom async worker count
func NewOrchestratorWithAsyncWorkers(stateManager StateManager, asyncWorkers int) *Orchestrator {
	return &Orchestrator{
		stateManager: stateManager,
		workflows:    make(map[string]*WorkflowDefinition),
		asyncWorkers: asyncWorkers,
	}
}

// RegisterWorkflow registers a workflow definition
func (o *Orchestrator) RegisterWorkflow(workflow *WorkflowDefinition) error {
	if workflow == nil {
		return fmt.Errorf("workflow cannot be nil")
	}
	if workflow.ID == "" {
		return fmt.Errorf("workflow ID cannot be empty")
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	o.workflows[workflow.ID] = workflow
	return nil
}

// GetWorkflow retrieves a registered workflow definition
func (o *Orchestrator) GetWorkflow(workflowID string) (*WorkflowDefinition, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	workflow, ok := o.workflows[workflowID]
	if !ok {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	return workflow, nil
}

// StartWorkflow starts a new workflow instance (synchronous execution)
func (o *Orchestrator) StartWorkflow(ctx context.Context, workflowID string, input map[string]interface{}, metadata map[string]interface{}) (*WorkflowResult, error) {
	// Get workflow definition
	workflow, err := o.GetWorkflow(workflowID)
	if err != nil {
		return nil, err
	}

	// Create workflow instance
	instance := &WorkflowInstance{
		ID:            uuid.New().String(),
		WorkflowID:    workflowID,
		Status:        WorkflowStatusPending,
		Input:         input,
		Output:        make(map[string]interface{}),
		Context:       make(map[string]interface{}),
		StartedAt:     time.Now(),
		Metadata:      metadata,
		TraceID:       getTraceID(ctx, metadata),
		CorrelationID: getCorrelationID(ctx, metadata),
		BusinessID:    getBusinessID(ctx, metadata),
		Steps:         make([]*StepInstance, 0),
	}

	// Save initial state
	if err := o.stateManager.SaveWorkflow(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to save workflow: %w", err)
	}

	// Emit workflow started event
	o.emitEvent(ctx, instance.ID, nil, "workflow.started", map[string]interface{}{
		"workflow_id": workflowID,
	})

	// Execute workflow synchronously
	return o.executeWorkflow(ctx, workflow, instance)
}

// StartWorkflowAsync starts a new workflow instance asynchronously
func (o *Orchestrator) StartWorkflowAsync(ctx context.Context, workflowID string, input map[string]interface{}, metadata map[string]interface{}) (string, error) {
	// Get workflow definition
	workflow, err := o.GetWorkflow(workflowID)
	if err != nil {
		return "", err
	}

	// Create workflow instance
	instance := &WorkflowInstance{
		ID:            uuid.New().String(),
		WorkflowID:    workflowID,
		Status:        WorkflowStatusPending,
		Input:         input,
		Output:        make(map[string]interface{}),
		Context:       make(map[string]interface{}),
		StartedAt:     time.Now(),
		Metadata:      metadata,
		TraceID:       getTraceID(ctx, metadata),
		CorrelationID: getCorrelationID(ctx, metadata),
		BusinessID:    getBusinessID(ctx, metadata),
		Steps:         make([]*StepInstance, 0),
	}

	// Save initial state
	if err := o.stateManager.SaveWorkflow(ctx, instance); err != nil {
		return "", fmt.Errorf("failed to save workflow: %w", err)
	}

	// Emit workflow started event
	o.emitEvent(ctx, instance.ID, nil, "workflow.started", map[string]interface{}{
		"workflow_id": workflowID,
	})

	// Start async execution in a goroutine
	go func() {
		asyncCtx := context.Background()
		o.executeWorkflow(asyncCtx, workflow, instance)
	}()

	return instance.ID, nil
}

// ResumeWorkflow resumes a workflow from a saved state
func (o *Orchestrator) ResumeWorkflow(ctx context.Context, workflowInstID string) (*WorkflowResult, error) {
	// Load workflow instance
	instance, err := o.stateManager.GetWorkflow(ctx, workflowInstID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow: %w", err)
	}

	// Get workflow definition
	workflow, err := o.GetWorkflow(instance.WorkflowID)
	if err != nil {
		return nil, err
	}

	// Check if workflow can be resumed
	if instance.IsCompleted() {
		return &WorkflowResult{
			Success:      instance.Status == WorkflowStatusCompleted,
			WorkflowInst: instance,
			Output:       instance.Output,
			Duration:     time.Since(instance.StartedAt),
		}, nil
	}

	// Resume execution
	return o.executeWorkflow(ctx, workflow, instance)
}

// GetWorkflowStatus retrieves the current status of a workflow
func (o *Orchestrator) GetWorkflowStatus(ctx context.Context, workflowInstID string) (*WorkflowInstance, error) {
	return o.stateManager.GetWorkflow(ctx, workflowInstID)
}

// ListWorkflows lists workflows with optional filters
func (o *Orchestrator) ListWorkflows(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*WorkflowInstance, int64, error) {
	return o.stateManager.ListWorkflows(ctx, filters, limit, offset)
}

// executeWorkflow executes a workflow instance
func (o *Orchestrator) executeWorkflow(ctx context.Context, workflow *WorkflowDefinition, instance *WorkflowInstance) (*WorkflowResult, error) {
	startTime := time.Now()

	// Update status to running
	instance.Status = WorkflowStatusRunning
	if err := o.stateManager.UpdateWorkflowStatus(ctx, instance.ID, WorkflowStatusRunning); err != nil {
		return nil, fmt.Errorf("failed to update workflow status: %w", err)
	}

	// Initialize step instances if not already done
	if len(instance.Steps) == 0 {
		for i, stepDef := range workflow.Steps {
			stepInst := &StepInstance{
				ID:             uuid.New().String(),
				StepID:         stepDef.ID,
				WorkflowInstID: instance.ID,
				Status:         StepStatusPending,
				Input:          make(map[string]interface{}),
				Output:         make(map[string]interface{}),
				ExecutionOrder: i,
			}
			instance.Steps = append(instance.Steps, stepInst)

			if err := o.stateManager.SaveStep(ctx, stepInst); err != nil {
				return nil, fmt.Errorf("failed to save step: %w", err)
			}
		}
	}

	// Build dependency graph
	graph := o.buildDependencyGraph(workflow)

	// Execute steps based on dependencies
	if err := o.executeSteps(ctx, workflow, instance, graph); err != nil {
		// Mark workflow as failed
		instance.Status = WorkflowStatusFailed
		instance.Error = stringPtr(err.Error())
		now := time.Now()
		instance.CompletedAt = &now

		o.stateManager.UpdateWorkflowStatus(ctx, instance.ID, WorkflowStatusFailed)
		o.stateManager.UpdateWorkflowError(ctx, instance.ID, err)

		o.emitEvent(ctx, instance.ID, nil, "workflow.failed", map[string]interface{}{
			"error": err.Error(),
		})

		return &WorkflowResult{
			Success:      false,
			WorkflowInst: instance,
			Error:        err,
			Duration:     time.Since(startTime),
		}, err
	}

	// Mark workflow as completed
	instance.Status = WorkflowStatusCompleted
	now := time.Now()
	instance.CompletedAt = &now

	if err := o.stateManager.UpdateWorkflowStatus(ctx, instance.ID, WorkflowStatusCompleted); err != nil {
		return nil, fmt.Errorf("failed to update workflow status: %w", err)
	}

	o.emitEvent(ctx, instance.ID, nil, "workflow.completed", map[string]interface{}{
		"duration_ms": time.Since(startTime).Milliseconds(),
	})

	return &WorkflowResult{
		Success:      true,
		WorkflowInst: instance,
		Output:       instance.Output,
		Duration:     time.Since(startTime),
	}, nil
}

// executeSteps executes workflow steps based on dependency graph
func (o *Orchestrator) executeSteps(ctx context.Context, workflow *WorkflowDefinition, instance *WorkflowInstance, graph map[string][]string) error {
	executed := make(map[string]bool)
	stepDefMap := make(map[string]*StepDefinition)
	stepInstMap := make(map[string]*StepInstance)

	// Create maps for quick lookup
	for _, stepDef := range workflow.Steps {
		stepDefMap[stepDef.ID] = stepDef
	}
	for _, stepInst := range instance.Steps {
		stepInstMap[stepInst.StepID] = stepInst
	}

	// Execute steps in order based on dependencies
	for {
		// Find steps that can be executed (all dependencies met)
		readySteps := o.findReadySteps(workflow, executed, graph)
		if len(readySteps) == 0 {
			break
		}

		// Sort ready steps by priority (higher priority first)
		sort.Slice(readySteps, func(i, j int) bool {
			return readySteps[i].Priority > readySteps[j].Priority
		})

		// Separate sync and async steps
		var syncSteps, asyncSteps []*StepDefinition
		for _, stepDef := range readySteps {
			if stepDef.Async {
				asyncSteps = append(asyncSteps, stepDef)
			} else {
				syncSteps = append(syncSteps, stepDef)
			}
		}

		// Execute sync steps sequentially
		for _, stepDef := range syncSteps {
			stepInst := stepInstMap[stepDef.ID]
			if stepInst.Status == StepStatusCompleted {
				executed[stepDef.ID] = true
				continue
			}

			if err := o.executeStep(ctx, stepDef, stepInst, instance, stepInstMap); err != nil {
				if stepDef.Required {
					return err
				} else {
					// Non-required step failed, mark as skipped and continue
					stepInst.Status = StepStatusSkipped
					o.stateManager.UpdateStepStatus(ctx, stepInst.ID, StepStatusSkipped)
				}
			}
			executed[stepDef.ID] = true
		}

		// Execute async steps concurrently using goroutines
		if len(asyncSteps) > 0 {
			var wg sync.WaitGroup
			errors := make(chan error, len(asyncSteps))

			for _, stepDef := range asyncSteps {
				stepInst := stepInstMap[stepDef.ID]
				if stepInst.Status == StepStatusCompleted {
					executed[stepDef.ID] = true
					continue
				}

				wg.Add(1)
				go func(sd *StepDefinition, si *StepInstance) {
					defer wg.Done()
					if err := o.executeStep(ctx, sd, si, instance, stepInstMap); err != nil {
						if sd.Required {
							errors <- err
						} else {
							// Non-required step failed, mark as skipped and continue
							si.Status = StepStatusSkipped
							o.stateManager.UpdateStepStatus(ctx, si.ID, StepStatusSkipped)
						}
					}
					executed[sd.ID] = true
				}(stepDef, stepInst)
			}

			wg.Wait()
			close(errors)

			// Check for errors
			for err := range errors {
				if err != nil {
					return err
				}
			}
		}

		// Check if all steps are executed
		if len(executed) == len(workflow.Steps) {
			break
		}
	}

	return nil
}

// executeStep executes a single step with retry logic
func (o *Orchestrator) executeStep(ctx context.Context, stepDef *StepDefinition, stepInst *StepInstance, workflowInst *WorkflowInstance, stepInstMap map[string]*StepInstance) error {
	// Check if step is already completed
	if stepInst.IsCompleted() {
		return nil
	}

	// Prepare input from previous steps
	input := o.prepareStepInput(stepDef, stepInst, workflowInst, stepInstMap)

	// Apply timeout if specified
	stepCtx := ctx
	if stepDef.Timeout > 0 {
		var cancel context.CancelFunc
		stepCtx, cancel = context.WithTimeout(ctx, stepDef.Timeout)
		defer cancel()
	}

	// Execute with retry
	retryPolicy := stepDef.RetryPolicy
	if retryPolicy == nil {
		retryPolicy = &RetryPolicy{
			MaxAttempts:     1,
			InitialInterval: 0,
		}
	}

	var lastErr error
	for attempt := 0; attempt < retryPolicy.MaxAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			interval := o.calculateRetryInterval(retryPolicy, attempt)
			time.Sleep(interval)

			stepInst.Status = StepStatusRetrying
			stepInst.RetryCount = attempt
			now := time.Now()
			stepInst.LastRetryAt = &now
			o.stateManager.UpdateStepStatus(stepCtx, stepInst.ID, StepStatusRetrying)

			o.emitEvent(stepCtx, workflowInst.ID, &stepInst.ID, "step.retry", map[string]interface{}{
				"attempt": attempt + 1,
			})
		}

		// Mark step as running
		if attempt == 0 {
			stepInst.Status = StepStatusRunning
			now := time.Now()
			stepInst.StartedAt = &now
			o.stateManager.UpdateStepStatus(stepCtx, stepInst.ID, StepStatusRunning)

			o.emitEvent(stepCtx, workflowInst.ID, &stepInst.ID, "step.started", map[string]interface{}{
				"step_id": stepDef.ID,
			})
		}

		// Execute step
		startTime := time.Now()
		output, err := stepDef.Executor(stepCtx, input)
		duration := time.Since(startTime)

		stepInst.DurationMs = duration.Milliseconds()

		if err == nil {
			// Step succeeded
			stepInst.Status = StepStatusCompleted
			stepInst.Output = output
			now := time.Now()
			stepInst.CompletedAt = &now

			o.stateManager.UpdateStepStatus(stepCtx, stepInst.ID, StepStatusCompleted)
			o.stateManager.UpdateStepOutput(stepCtx, stepInst.ID, output)

			o.emitEvent(stepCtx, workflowInst.ID, &stepInst.ID, "step.completed", map[string]interface{}{
				"duration_ms": duration.Milliseconds(),
			})

			// Merge output to workflow context
			o.mergeStepOutput(workflowInst, stepDef.ID, output)

			return nil
		}

		lastErr = err
	}

	// All retries exhausted
	stepInst.Status = StepStatusFailed
	stepInst.Error = stringPtr(lastErr.Error())
	now := time.Now()
	stepInst.CompletedAt = &now

	o.stateManager.UpdateStepStatus(ctx, stepInst.ID, StepStatusFailed)
	o.stateManager.UpdateStepError(ctx, stepInst.ID, lastErr)

	o.emitEvent(ctx, workflowInst.ID, &stepInst.ID, "step.failed", map[string]interface{}{
		"error":   lastErr.Error(),
		"retries": stepInst.RetryCount,
	})

	return fmt.Errorf("step %s failed after %d attempts: %w", stepDef.ID, retryPolicy.MaxAttempts, lastErr)
}

// buildDependencyGraph builds a dependency graph from workflow steps
func (o *Orchestrator) buildDependencyGraph(workflow *WorkflowDefinition) map[string][]string {
	graph := make(map[string][]string)
	for _, step := range workflow.Steps {
		graph[step.ID] = step.Dependencies
	}
	return graph
}

// findReadySteps finds steps that can be executed (all dependencies met)
func (o *Orchestrator) findReadySteps(workflow *WorkflowDefinition, executed map[string]bool, graph map[string][]string) []*StepDefinition {
	ready := make([]*StepDefinition, 0)

	for _, step := range workflow.Steps {
		if executed[step.ID] {
			continue
		}

		// Check if all dependencies are executed
		allDepsExecuted := true
		for _, dep := range graph[step.ID] {
			if !executed[dep] {
				allDepsExecuted = false
				break
			}
		}

		if allDepsExecuted {
			ready = append(ready, step)
		}
	}

	return ready
}

// prepareStepInput prepares input for a step from workflow input and previous step outputs
func (o *Orchestrator) prepareStepInput(stepDef *StepDefinition, stepInst *StepInstance, workflowInst *WorkflowInstance, stepInstMap map[string]*StepInstance) map[string]interface{} {
	input := make(map[string]interface{})

	// Start with workflow input
	for k, v := range workflowInst.Input {
		input[k] = v
	}

	// Add outputs from dependency steps
	for _, depID := range stepDef.Dependencies {
		if depInst, ok := stepInstMap[depID]; ok {
			for k, v := range depInst.Output {
				input[k] = v
			}
			// Also add with step prefix
			input[depID] = depInst.Output
		}
	}

	// Add workflow context
	for k, v := range workflowInst.Context {
		input[k] = v
	}

	return input
}

// mergeStepOutput merges step output into workflow context
func (o *Orchestrator) mergeStepOutput(workflowInst *WorkflowInstance, stepID string, output map[string]interface{}) {
	if workflowInst.Context == nil {
		workflowInst.Context = make(map[string]interface{})
	}

	// Store output under step ID
	workflowInst.Context[stepID] = output

	// Also merge directly into context
	for k, v := range output {
		workflowInst.Context[k] = v
	}

	// Update workflow output
	if workflowInst.Output == nil {
		workflowInst.Output = make(map[string]interface{})
	}
	for k, v := range output {
		workflowInst.Output[k] = v
	}
}

// calculateRetryInterval calculates the retry interval with exponential backoff
func (o *Orchestrator) calculateRetryInterval(policy *RetryPolicy, attempt int) time.Duration {
	if policy.InitialInterval == 0 {
		return 0
	}

	interval := float64(policy.InitialInterval) * pow(policy.Multiplier, float64(attempt-1))
	if policy.MaxInterval > 0 && time.Duration(interval) > policy.MaxInterval {
		return policy.MaxInterval
	}

	return time.Duration(interval)
}

// emitEvent emits a workflow event
func (o *Orchestrator) emitEvent(ctx context.Context, workflowInstID string, stepInstID *string, eventType string, data map[string]interface{}) {
	event := &WorkflowEvent{
		ID:             uuid.New().String(),
		WorkflowInstID: workflowInstID,
		StepInstID:     stepInstID,
		EventType:      eventType,
		EventData:      data,
		Timestamp:      time.Now(),
	}

	// Best effort - don't fail workflow if event saving fails
	o.stateManager.SaveEvent(ctx, event)
}

// Helper functions

func getTraceID(ctx context.Context, metadata map[string]interface{}) string {
	if metadata != nil {
		if traceID, ok := metadata["trace_id"].(string); ok {
			return traceID
		}
	}
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return uuid.New().String()
}

func getCorrelationID(ctx context.Context, metadata map[string]interface{}) string {
	if metadata != nil {
		if correlationID, ok := metadata["correlation_id"].(string); ok {
			return correlationID
		}
	}
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}
	return uuid.New().String()
}

func getBusinessID(ctx context.Context, metadata map[string]interface{}) string {
	if metadata != nil {
		if businessID, ok := metadata["business_id"].(string); ok {
			return businessID
		}
	}
	if businessID := ctx.Value("business_id"); businessID != nil {
		if id, ok := businessID.(string); ok {
			return id
		}
	}
	return ""
}

func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	result := base
	for i := 1; i < int(exp); i++ {
		result *= base
	}
	return result
}
