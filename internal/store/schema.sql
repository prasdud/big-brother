CREATE TABLE workspace (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL
);

CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    workspace_id  TEXT NOT NULL REFERENCES workspace (id) ON DELETE CASCADE,
    email         TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL DEFAULT '',
    role          TEXT NOT NULL DEFAULT 'viewer'
                  CHECK (role IN ('admin', 'member', 'viewer')),
    google_sub    TEXT,
    disabled      INTEGER NOT NULL DEFAULT 0,
    created_at    TEXT NOT NULL,
    last_login_at TEXT
);

CREATE TABLE projects (
    id           TEXT PRIMARY KEY,
    workspace_id TEXT NOT NULL REFERENCES workspace (id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    slug         TEXT NOT NULL UNIQUE,
    created_at   TEXT NOT NULL
);

CREATE TABLE services (
    id                TEXT PRIMARY KEY,
    project_id        TEXT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    slug              TEXT NOT NULL,
    type              TEXT NOT NULL CHECK (type IN ('http', 'tcp', 'dns')),
    url               TEXT NOT NULL DEFAULT '',
    hostname          TEXT NOT NULL DEFAULT '',
    port              INTEGER NOT NULL DEFAULT 0,
    interval_seconds  INTEGER NOT NULL DEFAULT 60,
    timeout_seconds   INTEGER NOT NULL DEFAULT 10,
    failure_threshold INTEGER NOT NULL DEFAULT 3,
    enabled           INTEGER NOT NULL DEFAULT 1,
    next_run_at       TEXT,
    created_at        TEXT NOT NULL,
    updated_at        TEXT NOT NULL,
    UNIQUE (project_id, slug)
);

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
