-- In-app notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    data JSONB,
    action_url VARCHAR(500),
    image_url VARCHAR(500),
    read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_user_unread ON notifications(user_id, read) WHERE read = FALSE;
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);

-- User notification preferences
CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id UUID PRIMARY KEY,
    email_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sms_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    push_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    in_app_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    quiet_start TIME,
    quiet_end TIME,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Push notification tokens
CREATE TABLE IF NOT EXISTS push_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token VARCHAR(500) NOT NULL,
    platform VARCHAR(20) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, device_id)
);

CREATE INDEX idx_push_tokens_user_id ON push_tokens(user_id);
CREATE INDEX idx_push_tokens_active ON push_tokens(user_id, active) WHERE active = TRUE;

-- Notification delivery logs for audit trail
CREATE TABLE IF NOT EXISTS notification_delivery_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID,
    user_id UUID NOT NULL,
    channel VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    provider VARCHAR(50),
    provider_message_id VARCHAR(255),
    error_message TEXT,
    attempts INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_delivery_logs_notification_id ON notification_delivery_logs(notification_id);
CREATE INDEX idx_delivery_logs_user_id ON notification_delivery_logs(user_id);
CREATE INDEX idx_delivery_logs_status ON notification_delivery_logs(status);
