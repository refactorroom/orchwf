-- Create ORCHWF (Workflow Orchestration) tables

-- Workflow instances table
CREATE TABLE IF NOT EXISTS orchwf_workflow_instances (
    id VARCHAR(36) PRIMARY KEY,
    workflow_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    input JSONB DEFAULT '{}',
    output JSONB DEFAULT '{}',
    context JSONB DEFAULT '{}',
    current_step_id VARCHAR(255),
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    error TEXT,
    retry_count INT DEFAULT 0,
    last_retry_at TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    trace_id VARCHAR(255),
    correlation_id VARCHAR(255),
    business_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for workflow instances
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_instances_workflow_id ON orchwf_workflow_instances(workflow_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_instances_status ON orchwf_workflow_instances(status);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_instances_trace_id ON orchwf_workflow_instances(trace_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_instances_correlation_id ON orchwf_workflow_instances(correlation_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_instances_business_id ON orchwf_workflow_instances(business_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_instances_created_at ON orchwf_workflow_instances(created_at DESC);

-- Step instances table
CREATE TABLE IF NOT EXISTS orchwf_step_instances (
    id VARCHAR(36) PRIMARY KEY,
    step_id VARCHAR(255) NOT NULL,
    workflow_inst_id VARCHAR(36) NOT NULL,
    status VARCHAR(50) NOT NULL,
    input JSONB DEFAULT '{}',
    output JSONB DEFAULT '{}',
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error TEXT,
    retry_count INT DEFAULT 0,
    last_retry_at TIMESTAMP,
    duration_ms BIGINT DEFAULT 0,
    execution_order INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (workflow_inst_id) REFERENCES orchwf_workflow_instances(id) ON DELETE CASCADE
);

-- Create indexes for step instances
CREATE INDEX IF NOT EXISTS idx_orchwf_step_instances_step_id ON orchwf_step_instances(step_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_step_instances_workflow_inst_id ON orchwf_step_instances(workflow_inst_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_step_instances_status ON orchwf_step_instances(status);
CREATE INDEX IF NOT EXISTS idx_orchwf_step_instances_execution_order ON orchwf_step_instances(workflow_inst_id, execution_order);

-- Workflow events table
CREATE TABLE IF NOT EXISTS orchwf_workflow_events (
    id VARCHAR(36) PRIMARY KEY,
    workflow_inst_id VARCHAR(36) NOT NULL,
    step_inst_id VARCHAR(36),
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB DEFAULT '{}',
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (workflow_inst_id) REFERENCES orchwf_workflow_instances(id) ON DELETE CASCADE,
    FOREIGN KEY (step_inst_id) REFERENCES orchwf_step_instances(id) ON DELETE CASCADE
);

-- Create indexes for workflow events
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_events_workflow_inst_id ON orchwf_workflow_events(workflow_inst_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_events_step_inst_id ON orchwf_workflow_events(step_inst_id);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_events_event_type ON orchwf_workflow_events(event_type);
CREATE INDEX IF NOT EXISTS idx_orchwf_workflow_events_timestamp ON orchwf_workflow_events(timestamp DESC);

-- Create updated_at trigger function for workflow instances
CREATE OR REPLACE FUNCTION update_orchwf_workflow_instances_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for workflow instances
DROP TRIGGER IF EXISTS trigger_update_orchwf_workflow_instances_updated_at ON orchwf_workflow_instances;
CREATE TRIGGER trigger_update_orchwf_workflow_instances_updated_at
    BEFORE UPDATE ON orchwf_workflow_instances
    FOR EACH ROW
    EXECUTE FUNCTION update_orchwf_workflow_instances_updated_at();

-- Create updated_at trigger function for step instances
CREATE OR REPLACE FUNCTION update_orchwf_step_instances_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for step instances
DROP TRIGGER IF EXISTS trigger_update_orchwf_step_instances_updated_at ON orchwf_step_instances;
CREATE TRIGGER trigger_update_orchwf_step_instances_updated_at
    BEFORE UPDATE ON orchwf_step_instances
    FOR EACH ROW
    EXECUTE FUNCTION update_orchwf_step_instances_updated_at();
