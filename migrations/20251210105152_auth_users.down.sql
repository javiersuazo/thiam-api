-- Remove foreign key constraints from notification tables
ALTER TABLE notification_delivery_logs DROP CONSTRAINT IF EXISTS fk_notification_delivery_logs_user_id;
ALTER TABLE push_tokens DROP CONSTRAINT IF EXISTS fk_push_tokens_user_id;
ALTER TABLE notification_preferences DROP CONSTRAINT IF EXISTS fk_notification_preferences_user_id;
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS fk_notifications_user_id;

-- Drop refresh tokens
DROP TABLE IF EXISTS refresh_tokens;

-- Drop users
DROP TABLE IF EXISTS users;
