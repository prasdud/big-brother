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

-- name: ListDueServices :many
SELECT * FROM services
WHERE enabled = 1 AND (next_run_at IS NULL OR next_run_at <= sqlc.arg(now))
ORDER BY COALESCE(next_run_at, '') ASC
LIMIT sqlc.arg(max_rows);

-- name: SetServiceNextRun :exec
UPDATE services SET next_run_at = sqlc.arg(next_run_at) WHERE id = sqlc.arg(id);

-- name: InsertCheck :exec
INSERT INTO checks (
    id, service_id, project_id, checked_at, status, status_code, latency_ms, error
) VALUES (
    sqlc.arg(id), sqlc.arg(service_id), sqlc.arg(project_id), sqlc.arg(checked_at),
    sqlc.arg(status), sqlc.arg(status_code), sqlc.arg(latency_ms), sqlc.arg(error)
);

-- name: ListChecks :many
SELECT * FROM checks
WHERE service_id = sqlc.arg(service_id)
  AND checked_at >= sqlc.arg(from_at)
  AND checked_at <= sqlc.arg(to_at)
ORDER BY checked_at DESC
LIMIT sqlc.arg(max_rows);

-- name: PruneChecksBefore :execrows
DELETE FROM checks
WHERE id IN (
    SELECT c.id FROM checks AS c
    WHERE c.checked_at < sqlc.arg(before)
    ORDER BY c.checked_at
    LIMIT sqlc.arg(max_rows)
);

-- name: GetServiceState :one
SELECT * FROM service_state WHERE service_id = sqlc.arg(service_id);

-- name: ListServiceStatesByProject :many
SELECT * FROM service_state WHERE project_id = sqlc.arg(project_id);

-- name: UpsertServiceState :exec
INSERT INTO service_state (
    service_id, project_id, state, consecutive_failures,
    consecutive_successes, last_change_at, last_check_at
) VALUES (
    sqlc.arg(service_id), sqlc.arg(project_id), sqlc.arg(state),
    sqlc.arg(consecutive_failures), sqlc.arg(consecutive_successes),
    sqlc.arg(last_change_at), sqlc.arg(last_check_at)
)
ON CONFLICT (service_id) DO UPDATE SET
    state = excluded.state,
    consecutive_failures = excluded.consecutive_failures,
    consecutive_successes = excluded.consecutive_successes,
    last_change_at = excluded.last_change_at,
    last_check_at = excluded.last_check_at;

-- name: UpsertRollup :exec
INSERT INTO uptime_rollups (service_id, project_id, hour, up_checks, total_checks)
VALUES (sqlc.arg(service_id), sqlc.arg(project_id), sqlc.arg(hour),
        sqlc.arg(up_checks), sqlc.arg(total_checks))
ON CONFLICT (service_id, hour) DO UPDATE SET
    up_checks = up_checks + excluded.up_checks,
    total_checks = total_checks + excluded.total_checks;

-- name: SumUptime :one
SELECT
    CAST(COALESCE(SUM(up_checks), 0) AS INTEGER)    AS up_checks,
    CAST(COALESCE(SUM(total_checks), 0) AS INTEGER) AS total_checks
FROM uptime_rollups
WHERE service_id = sqlc.arg(service_id) AND hour >= sqlc.arg(from_hour);
