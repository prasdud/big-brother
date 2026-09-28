-- name: CountWorkspaces :one
SELECT COUNT(*) FROM workspace;

-- name: CreateWorkspace :exec
INSERT INTO workspace (id, name, slug, created_at)
VALUES (sqlc.arg(id), sqlc.arg(name), sqlc.arg(slug), sqlc.arg(created_at));

-- name: GetWorkspace :one
SELECT id, name, slug, created_at FROM workspace LIMIT 1;

-- name: ListProjects :many
SELECT id, workspace_id, name, slug, created_at FROM projects ORDER BY name;

-- name: GetProjectBySlug :one
SELECT id, workspace_id, name, slug, created_at
FROM projects WHERE slug = sqlc.arg(slug);

-- name: CreateProject :exec
INSERT INTO projects (id, workspace_id, name, slug, created_at)
VALUES (sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(name), sqlc.arg(slug), sqlc.arg(created_at));

-- name: RenameProject :exec
UPDATE projects SET name = sqlc.arg(name) WHERE id = sqlc.arg(id);

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = sqlc.arg(id);

-- name: ListServices :many
SELECT * FROM services WHERE project_id = sqlc.arg(project_id) ORDER BY name;

-- name: GetServiceBySlug :one
SELECT * FROM services
WHERE project_id = sqlc.arg(project_id) AND slug = sqlc.arg(slug);

-- name: CreateService :exec
INSERT INTO services (
    id, project_id, name, slug, type, url, hostname, port,
    interval_seconds, timeout_seconds, failure_threshold, enabled,
    created_at, updated_at
) VALUES (
    sqlc.arg(id), sqlc.arg(project_id), sqlc.arg(name), sqlc.arg(slug),
    sqlc.arg(type), sqlc.arg(url), sqlc.arg(hostname), sqlc.arg(port),
    sqlc.arg(interval_seconds), sqlc.arg(timeout_seconds),
    sqlc.arg(failure_threshold), sqlc.arg(enabled),
    sqlc.arg(created_at), sqlc.arg(updated_at)
);

-- name: UpdateService :exec
UPDATE services SET
    name = sqlc.arg(name),
    url = sqlc.arg(url),
    hostname = sqlc.arg(hostname),
    port = sqlc.arg(port),
    interval_seconds = sqlc.arg(interval_seconds),
    timeout_seconds = sqlc.arg(timeout_seconds),
    failure_threshold = sqlc.arg(failure_threshold),
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: SetServiceEnabled :exec
UPDATE services SET enabled = sqlc.arg(enabled), updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id);

-- name: DeleteService :exec
DELETE FROM services WHERE id = sqlc.arg(id);
