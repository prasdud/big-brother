-- +goose Up
ALTER TABLE projects ADD COLUMN default_channel_id TEXT NOT NULL DEFAULT '';
ALTER TABLE services ADD COLUMN channel_id TEXT NOT NULL DEFAULT '';
ALTER TABLE services ADD COLUMN template_down TEXT NOT NULL DEFAULT '';
ALTER TABLE services ADD COLUMN template_recovered TEXT NOT NULL DEFAULT '';

CREATE TABLE slack_workspace (
    workspace_id  TEXT PRIMARY KEY REFERENCES workspace (id) ON DELETE CASCADE,
    team_id       TEXT NOT NULL,
    team_name     TEXT NOT NULL DEFAULT '',
    bot_token_enc BLOB NOT NULL,
    installed_at  TEXT NOT NULL
);

CREATE TABLE channels (
    id               TEXT PRIMARY KEY,
    project_id       TEXT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    name             TEXT NOT NULL DEFAULT '',
    slack_channel_id TEXT NOT NULL,
    created_at       TEXT NOT NULL,
    UNIQUE (project_id, slack_channel_id)
);

CREATE INDEX idx_channels_project ON channels (project_id);

CREATE TABLE alert_templates (
    id         TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    trigger    TEXT NOT NULL CHECK (trigger IN ('down', 'recovered')),
    body       TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (project_id, trigger)
);

CREATE TABLE delivery_failures (
    id         TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    service_id TEXT NOT NULL DEFAULT '',
    trigger    TEXT NOT NULL DEFAULT '',
    reason     TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX idx_delivery_failures_project ON delivery_failures (project_id, created_at);

-- +goose Down
DROP TABLE delivery_failures;
DROP TABLE alert_templates;
DROP TABLE channels;
DROP TABLE slack_workspace;
ALTER TABLE services DROP COLUMN template_recovered;
ALTER TABLE services DROP COLUMN template_down;
ALTER TABLE services DROP COLUMN channel_id;
ALTER TABLE projects DROP COLUMN default_channel_id;
