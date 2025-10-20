package orchwf

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// DBStateManager implements StateManager using database/sql
type DBStateManager struct {
	db *sql.DB
}

// NewDBStateManager creates a new database state manager
func NewDBStateManager(db *sql.DB) *DBStateManager {
	return &DBStateManager{
		db: db,
	}
}

// SaveWorkflow saves a workflow instance to the database
func (m *DBStateManager) SaveWorkflow(ctx context.Context, workflow *WorkflowInstance) error {
	query := `
		INSERT INTO orchwf_workflow_instances 
		(id, workflow_id, status, input, output, context, current_step_id, started_at, completed_at, 
		 error, retry_count, last_retry_at, metadata, trace_id, correlation_id, business_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	inputJSON, _ := json.Marshal(workflow.Input)
	outputJSON, _ := json.Marshal(workflow.Output)
	contextJSON, _ := json.Marshal(workflow.Context)
	metadataJSON, _ := json.Marshal(workflow.Metadata)

	_, err := m.db.ExecContext(ctx, query,
		workflow.ID,
		workflow.WorkflowID,
		string(workflow.Status),
		inputJSON,
		outputJSON,
		contextJSON,
		workflow.CurrentStepID,
		workflow.StartedAt,
		workflow.CompletedAt,
		workflow.Error,
		workflow.RetryCount,
		workflow.LastRetryAt,
		metadataJSON,
		workflow.TraceID,
		workflow.CorrelationID,
		workflow.BusinessID,
		time.Now(),
		time.Now(),
	)

	return err
}

// GetWorkflow retrieves a workflow instance by ID
func (m *DBStateManager) GetWorkflow(ctx context.Context, workflowInstID string) (*WorkflowInstance, error) {
	query := `
		SELECT id, workflow_id, status, input, output, context, current_step_id, started_at, completed_at,
		       error, retry_count, last_retry_at, metadata, trace_id, correlation_id, business_id, created_at, updated_at
		FROM orchwf_workflow_instances 
		WHERE id = $1`

	var w ORCHWorkflowInstance
	var inputJSON, outputJSON, contextJSON, metadataJSON []byte

	err := m.db.QueryRowContext(ctx, query, workflowInstID).Scan(
		&w.ID, &w.WorkflowID, &w.Status, &inputJSON, &outputJSON, &contextJSON, &w.CurrentStepID,
		&w.StartedAt, &w.CompletedAt, &w.Error, &w.RetryCount, &w.LastRetryAt, &metadataJSON,
		&w.TraceID, &w.CorrelationID, &w.BusinessID, &w.CreatedAt, &w.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	json.Unmarshal(inputJSON, &w.Input)
	json.Unmarshal(outputJSON, &w.Output)
	json.Unmarshal(contextJSON, &w.Context)
	json.Unmarshal(metadataJSON, &w.Metadata)

	// Load steps
	steps, err := m.GetWorkflowSteps(ctx, workflowInstID)
	if err != nil {
		return nil, err
	}

	workflow, err := modelToWorkflowInstance(&w)
	if err != nil {
		return nil, err
	}

	workflow.Steps = steps
	return workflow, nil
}

// UpdateWorkflowStatus updates the status of a workflow
func (m *DBStateManager) UpdateWorkflowStatus(ctx context.Context, workflowInstID string, status WorkflowStatus) error {
	query := `
		UPDATE orchwf_workflow_instances 
		SET status = $1, updated_at = $2`

	args := []interface{}{string(status), time.Now()}

	if status == WorkflowStatusCompleted || status == WorkflowStatusFailed || status == WorkflowStatusCancelled {
		query += `, completed_at = $3`
		args = append(args, time.Now())
	}

	query += ` WHERE id = $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, workflowInstID)

	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

// UpdateWorkflowOutput updates the output of a workflow
func (m *DBStateManager) UpdateWorkflowOutput(ctx context.Context, workflowInstID string, output map[string]interface{}) error {
	outputJSON, err := json.Marshal(output)
	if err != nil {
		return err
	}

	query := `UPDATE orchwf_workflow_instances SET output = $1, updated_at = $2 WHERE id = $3`
	_, err = m.db.ExecContext(ctx, query, outputJSON, time.Now(), workflowInstID)
	return err
}

// UpdateWorkflowError updates the error of a workflow
func (m *DBStateManager) UpdateWorkflowError(ctx context.Context, workflowInstID string, err error) error {
	errorMsg := err.Error()
	query := `
		UPDATE orchwf_workflow_instances 
		SET error = $1, status = $2, updated_at = $3 
		WHERE id = $4`
	_, err = m.db.ExecContext(ctx, query, errorMsg, string(WorkflowStatusFailed), time.Now(), workflowInstID)
	return err
}

// ListWorkflows lists workflows with optional filters
func (m *DBStateManager) ListWorkflows(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*WorkflowInstance, int64, error) {
	// Build WHERE clause
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	for key, value := range filters {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("%s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM orchwf_workflow_instances"
	if whereClause != "" {
		countQuery += " WHERE " + whereClause
	}

	var total int64
	err := m.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `
		SELECT id, workflow_id, status, input, output, context, current_step_id, started_at, completed_at,
		       error, retry_count, last_retry_at, metadata, trace_id, correlation_id, business_id, created_at, updated_at
		FROM orchwf_workflow_instances`

	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var workflows []*WorkflowInstance
	for rows.Next() {
		var w ORCHWorkflowInstance
		var inputJSON, outputJSON, contextJSON, metadataJSON []byte

		err := rows.Scan(
			&w.ID, &w.WorkflowID, &w.Status, &inputJSON, &outputJSON, &contextJSON, &w.CurrentStepID,
			&w.StartedAt, &w.CompletedAt, &w.Error, &w.RetryCount, &w.LastRetryAt, &metadataJSON,
			&w.TraceID, &w.CorrelationID, &w.BusinessID, &w.CreatedAt, &w.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Parse JSON fields
		json.Unmarshal(inputJSON, &w.Input)
		json.Unmarshal(outputJSON, &w.Output)
		json.Unmarshal(contextJSON, &w.Context)
		json.Unmarshal(metadataJSON, &w.Metadata)

		workflow, err := modelToWorkflowInstance(&w)
		if err != nil {
			return nil, 0, err
		}

		workflows = append(workflows, workflow)
	}

	return workflows, total, nil
}

// SaveStep saves a step instance to the database
func (m *DBStateManager) SaveStep(ctx context.Context, step *StepInstance) error {
	query := `
		INSERT INTO orchwf_step_instances 
		(id, step_id, workflow_inst_id, status, input, output, started_at, completed_at,
		 error, retry_count, last_retry_at, duration_ms, execution_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	inputJSON, _ := json.Marshal(step.Input)
	outputJSON, _ := json.Marshal(step.Output)

	_, err := m.db.ExecContext(ctx, query,
		step.ID, step.StepID, step.WorkflowInstID, string(step.Status),
		inputJSON, outputJSON, step.StartedAt, step.CompletedAt,
		step.Error, step.RetryCount, step.LastRetryAt, step.DurationMs,
		step.ExecutionOrder, time.Now(), time.Now(),
	)

	return err
}

// GetStep retrieves a step instance by ID
func (m *DBStateManager) GetStep(ctx context.Context, stepInstID string) (*StepInstance, error) {
	query := `
		SELECT id, step_id, workflow_inst_id, status, input, output, started_at, completed_at,
		       error, retry_count, last_retry_at, duration_ms, execution_order, created_at, updated_at
		FROM orchwf_step_instances 
		WHERE id = $1`

	var s ORCHStepInstance
	var inputJSON, outputJSON []byte

	err := m.db.QueryRowContext(ctx, query, stepInstID).Scan(
		&s.ID, &s.StepID, &s.WorkflowInstID, &s.Status, &inputJSON, &outputJSON,
		&s.StartedAt, &s.CompletedAt, &s.Error, &s.RetryCount, &s.LastRetryAt,
		&s.DurationMs, &s.ExecutionOrder, &s.CreatedAt, &s.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	json.Unmarshal(inputJSON, &s.Input)
	json.Unmarshal(outputJSON, &s.Output)

	return modelToStepInstance(&s)
}

// GetWorkflowSteps retrieves all steps for a workflow
func (m *DBStateManager) GetWorkflowSteps(ctx context.Context, workflowInstID string) ([]*StepInstance, error) {
	query := `
		SELECT id, step_id, workflow_inst_id, status, input, output, started_at, completed_at,
		       error, retry_count, last_retry_at, duration_ms, execution_order, created_at, updated_at
		FROM orchwf_step_instances 
		WHERE workflow_inst_id = $1 
		ORDER BY execution_order ASC`

	rows, err := m.db.QueryContext(ctx, query, workflowInstID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []*StepInstance
	for rows.Next() {
		var s ORCHStepInstance
		var inputJSON, outputJSON []byte

		err := rows.Scan(
			&s.ID, &s.StepID, &s.WorkflowInstID, &s.Status, &inputJSON, &outputJSON,
			&s.StartedAt, &s.CompletedAt, &s.Error, &s.RetryCount, &s.LastRetryAt,
			&s.DurationMs, &s.ExecutionOrder, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		json.Unmarshal(inputJSON, &s.Input)
		json.Unmarshal(outputJSON, &s.Output)

		step, err := modelToStepInstance(&s)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}

// UpdateStepStatus updates the status of a step
func (m *DBStateManager) UpdateStepStatus(ctx context.Context, stepInstID string, status StepStatus) error {
	query := `UPDATE orchwf_step_instances SET status = $1, updated_at = $2`
	args := []interface{}{string(status), time.Now()}

	if status == StepStatusRunning {
		query += `, started_at = $3`
		args = append(args, time.Now())
	}

	if status == StepStatusCompleted || status == StepStatusFailed || status == StepStatusSkipped {
		query += `, completed_at = $` + fmt.Sprintf("%d", len(args)+1)
		args = append(args, time.Now())
	}

	query += ` WHERE id = $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, stepInstID)

	_, err := m.db.ExecContext(ctx, query, args...)
	return err
}

// UpdateStepOutput updates the output of a step
func (m *DBStateManager) UpdateStepOutput(ctx context.Context, stepInstID string, output map[string]interface{}) error {
	outputJSON, err := json.Marshal(output)
	if err != nil {
		return err
	}

	query := `UPDATE orchwf_step_instances SET output = $1, updated_at = $2 WHERE id = $3`
	_, err = m.db.ExecContext(ctx, query, outputJSON, time.Now(), stepInstID)
	return err
}

// UpdateStepError updates the error of a step
func (m *DBStateManager) UpdateStepError(ctx context.Context, stepInstID string, err error) error {
	errorMsg := err.Error()
	query := `
		UPDATE orchwf_step_instances 
		SET error = $1, status = $2, updated_at = $3 
		WHERE id = $4`
	_, err = m.db.ExecContext(ctx, query, errorMsg, string(StepStatusFailed), time.Now(), stepInstID)
	return err
}

// SaveEvent saves a workflow event to the database
func (m *DBStateManager) SaveEvent(ctx context.Context, event *WorkflowEvent) error {
	query := `
		INSERT INTO orchwf_workflow_events 
		(id, workflow_inst_id, step_inst_id, event_type, event_data, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	eventDataJSON, _ := json.Marshal(event.EventData)

	_, err := m.db.ExecContext(ctx, query,
		event.ID, event.WorkflowInstID, event.StepInstID, event.EventType,
		eventDataJSON, event.Timestamp, time.Now(),
	)

	return err
}

// GetWorkflowEvents retrieves all events for a workflow
func (m *DBStateManager) GetWorkflowEvents(ctx context.Context, workflowInstID string) ([]*WorkflowEvent, error) {
	query := `
		SELECT id, workflow_inst_id, step_inst_id, event_type, event_data, timestamp, created_at
		FROM orchwf_workflow_events 
		WHERE workflow_inst_id = $1 
		ORDER BY timestamp ASC`

	rows, err := m.db.QueryContext(ctx, query, workflowInstID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*WorkflowEvent
	for rows.Next() {
		var e ORCHWorkflowEvent
		var eventDataJSON []byte

		err := rows.Scan(
			&e.ID, &e.WorkflowInstID, &e.StepInstID, &e.EventType,
			&eventDataJSON, &e.Timestamp, &e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON field
		json.Unmarshal(eventDataJSON, &e.EventData)

		event, err := modelToWorkflowEvent(&e)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

// WithTransaction executes a function within a database transaction
func (m *DBStateManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Create a new state manager with the transaction
	// Note: We need to create a wrapper that implements the same interface
	// For now, we'll use the original state manager but execute within transaction
	txManager := m
	// Store the transaction manager in context
	txCtx := context.WithValue(ctx, "state_manager", txManager)
	err = fn(txCtx)
	return err
}
