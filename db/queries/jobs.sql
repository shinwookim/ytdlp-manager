
-- name: CreateJob :one
INSERT INTO jobs (url, mode)
VALUES (?, ?)
RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs
WHERE id = ?;

-- name: ListJobs :many
SELECT * FROM jobs
ORDER BY id DESC;

-- name: ClaimNextJob :one
UPDATE jobs
SET status = 'downloading',
    updated_at = CURRENT_TIMESTAMP,
    progress = 0,
    error = ''
WHERE id = (
    SELECT id FROM jobs
    WHERE status = 'queued'
    ORDER BY id ASC
    LIMIT 1
) AND status = 'queued'
RETURNING *;

-- name: UpdateTitle :exec
UPDATE jobs
SET title = ?
WHERE id = ?;

-- name: UpdateProgress :exec
UPDATE jobs
SET progress = ?,
    speed = ?,
    eta = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND status = 'downloading';

-- name: CompleteJob :exec
UPDATE jobs
SET status = 'completed',
    progress = 100,
    speed = '',
    eta = '',
    output = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND status = 'downloading';

-- name: FailJob :exec
UPDATE jobs
SET status = 'failed',
    error = ?,
    speed = '',
    eta = '',
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND status = 'downloading';

-- name: CancelQueuedJob :execrows
UPDATE jobs
SET status = 'cancelled',
    error = 'Cancelled by user',
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND status = 'queued';

-- name: RecoverInterruptedJobs :exec
UPDATE jobs
SET status = 'queued',
    updated_at = CURRENT_TIMESTAMP
WHERE status = 'downloading';

-- name: DeleteJob :execrows
DELETE FROM jobs
WHERE id = ? AND status IN ('completed', 'failed', 'cancelled');

-- name: CancelRunningJob :exec
UPDATE jobs
SET status = 'cancelled',
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND status = 'downloading';

-- name: ClearFinishedJobs :execrows
DELETE FROM jobs
WHERE status IN ('completed', 'failed', 'cancelled');
