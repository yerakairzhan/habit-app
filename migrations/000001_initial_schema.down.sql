-- Migration: 000001_initial_schema.down.sql

DROP TRIGGER IF EXISTS trg_user_settings_updated_at ON user_settings;
DROP TRIGGER IF EXISTS trg_habit_logs_updated_at    ON habit_logs;
DROP TRIGGER IF EXISTS trg_habits_updated_at        ON habits;
DROP TRIGGER IF EXISTS trg_users_updated_at         ON users;
DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS habit_logs;
DROP TABLE IF EXISTS habits;
DROP TABLE IF EXISTS users;
