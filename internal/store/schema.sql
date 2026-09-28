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
    created_at        TEXT NOT NULL,
    updated_at        TEXT NOT NULL,
    UNIQUE (project_id, slug)
);
