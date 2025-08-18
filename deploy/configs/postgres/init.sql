-- Detectviz Platform Database Initialization

-- Create schema for knowledge base
CREATE SCHEMA IF NOT EXISTS knowledge_base;

-- Create incidents table
CREATE TABLE IF NOT EXISTS knowledge_base.incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id VARCHAR(255) UNIQUE NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    severity VARCHAR(50),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_minutes INTEGER,
    root_cause TEXT,
    impact TEXT,
    resolution TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create lessons learned table
CREATE TABLE IF NOT EXISTS knowledge_base.lessons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID REFERENCES knowledge_base.incidents(id),
    category VARCHAR(100),
    lesson TEXT NOT NULL,
    action_items JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create metrics snapshot table
CREATE TABLE IF NOT EXISTS knowledge_base.metrics_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID REFERENCES knowledge_base.incidents(id),
    metric_name VARCHAR(255),
    metric_value JSONB,
    timestamp TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_incidents_service ON knowledge_base.incidents(service_name);
CREATE INDEX idx_incidents_time ON knowledge_base.incidents(start_time, end_time);
CREATE INDEX idx_lessons_incident ON knowledge_base.lessons(incident_id);
CREATE INDEX idx_metrics_incident ON knowledge_base.metrics_snapshots(incident_id);

-- Create similarity search function (basic implementation)
CREATE OR REPLACE FUNCTION knowledge_base.find_similar_incidents(
    p_service_name VARCHAR,
    p_root_cause TEXT,
    p_limit INTEGER DEFAULT 5
)
RETURNS TABLE (
    incident_id VARCHAR,
    service_name VARCHAR,
    root_cause TEXT,
    similarity_score FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        i.incident_id,
        i.service_name,
        i.root_cause,
        similarity(i.root_cause, p_root_cause) as similarity_score
    FROM knowledge_base.incidents i
    WHERE i.service_name = p_service_name
        AND i.root_cause IS NOT NULL
    ORDER BY similarity_score DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Grant permissions
GRANT ALL PRIVILEGES ON SCHEMA knowledge_base TO detectviz;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA knowledge_base TO detectviz;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA knowledge_base TO detectviz;
