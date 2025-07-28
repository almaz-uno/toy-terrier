-- Drop notifications and user_settings tables
DROP TRIGGER IF EXISTS update_user_settings_updated_at ON user_settings;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS user_settings;
