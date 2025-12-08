CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE,
    retry_count INT NOT NULL DEFAULT 0,
    last_error TEXT
);

CREATE INDEX idx_outbox_events_unpublished ON outbox_events (created_at)
    WHERE published_at IS NULL;

CREATE INDEX idx_outbox_events_aggregate ON outbox_events (aggregate_type, aggregate_id);

COMMENT ON TABLE outbox_events IS 'Transactional outbox for reliable event publishing';
COMMENT ON COLUMN outbox_events.aggregate_type IS 'Type of aggregate that raised the event (e.g., user, order)';
COMMENT ON COLUMN outbox_events.aggregate_id IS 'ID of the aggregate instance';
COMMENT ON COLUMN outbox_events.event_type IS 'Event type for routing (e.g., user.created, order.placed)';
COMMENT ON COLUMN outbox_events.payload IS 'JSON payload of the event';
COMMENT ON COLUMN outbox_events.published_at IS 'When the event was published to the message queue';
COMMENT ON COLUMN outbox_events.retry_count IS 'Number of publish attempts';
COMMENT ON COLUMN outbox_events.last_error IS 'Last error message if publish failed';
