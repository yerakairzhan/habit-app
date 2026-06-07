-- Migration: 000001_initial_schema.up.sql
-- Run by golang-migrate on startup.

-- ─── Extensions ──────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- gen_random_uuid()

-- ─── Users ───────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        NOT NULL UNIQUE,
    name          TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

-- ─── Habits ──────────────────────────────────────────────────────────────────
CREATE TABLE habits (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title          TEXT        NOT NULL,
    color          TEXT        NOT NULL DEFAULT '#4F46E5',
    target_per_day INT         NOT NULL DEFAULT 1 CHECK (target_per_day >= 1),
    schedule_type  VARCHAR(32) NOT NULL,
    weekdays       JSONB       NOT NULL DEFAULT '[]',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ          -- soft delete
);

CREATE INDEX idx_habits_user_id    ON habits(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_habits_deleted_at ON habits(deleted_at);

-- ─── Habit Logs ──────────────────────────────────────────────────────────────
-- One row per (habit, calendar day). progress is incremented atomically.
CREATE TABLE habit_logs (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    habit_id   UUID        NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    date       DATE        NOT NULL,
    progress   INT         NOT NULL DEFAULT 0 CHECK (progress >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_habit_logs_habit_date UNIQUE (habit_id, date)
);

CREATE INDEX idx_habit_logs_user_id ON habit_logs(user_id);
CREATE INDEX idx_habit_logs_date    ON habit_logs(date);
-- Covering index for streak/stats queries.
CREATE INDEX idx_habit_logs_habit_date_progress ON habit_logs(habit_id, date, progress);

-- ─── Refresh Tokens ──────────────────────────────────────────────────────────
CREATE TABLE refresh_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id    ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- ─── User Settings ───────────────────────────────────────────────────────────
CREATE TABLE user_settings (
    id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    language              TEXT        NOT NULL DEFAULT 'en',
    notifications_enabled BOOLEAN     NOT NULL DEFAULT TRUE,
    show_dates            BOOLEAN     NOT NULL DEFAULT TRUE,
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── updated_at trigger ──────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_habits_updated_at
    BEFORE UPDATE ON habits
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_habit_logs_updated_at
    BEFORE UPDATE ON habit_logs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_user_settings_updated_at
    BEFORE UPDATE ON user_settings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
