-- +goose Up
ALTER TABLE services ADD COLUMN next_run_at TEXT;

CREATE INDEX idx_services_due ON services (enabled, next_run_at);

CREATE TABLE checks (
    id          TEXT PRIMARY KEY,
    service_id  TEXT NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    project_id  TEXT NOT NULL,
    checked_at  TEXT NOT NULL,
    status      TEXT NOT NULL CHECK (status IN ('up', 'down')),
    status_code INTEGER NOT NULL DEFAULT 0,
    latency_ms  INTEGER NOT NULL DEFAULT 0,
    error       TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_checks_service_checked ON checks (service_id, checked_at);
CREATE INDEX idx_checks_project_checked ON checks (project_id, checked_at);

CREATE TABLE service_state (
    service_id           TEXT PRIMARY KEY REFERENCES services (id) ON DELETE CASCADE,
    project_id           TEXT NOT NULL,
    state                TEXT NOT NULL CHECK (state IN ('up', 'down', 'pending', 'paused')),
    consecutive_failures INTEGER NOT NULL DEFAULT 0,
    consecutive_successes INTEGER NOT NULL DEFAULT 0,
    last_change_at       TEXT NOT NULL,
    last_check_at        TEXT
);

CREATE TABLE uptime_rollups (
    service_id   TEXT NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    project_id   TEXT NOT NULL,
    hour         TEXT NOT NULL,
    up_checks    INTEGER NOT NULL DEFAULT 0,
    total_checks INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (service_id, hour)
);

-- +goose Down
DROP TABLE uptime_rollups;
DROP TABLE service_state;
DROP TABLE checks;
DROP INDEX IF EXISTS idx_services_due;
ALTER TABLE services DROP COLUMN next_run_at;
