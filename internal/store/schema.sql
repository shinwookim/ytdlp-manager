CREATE TABLE IF NOT EXISTS jobs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    url         TEXT NOT NULL,
    mode        TEXT NOT NULL CHECK (mode IN ('audio', 'video')),
    title       TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'queued' CHECK (
        status IN (
            'queued',
            'downloading',
            'completed',
            'failed',
            'cancelled'
        )
    ),
    progress    REAL NOT NULL DEFAULT 0,
    speed       TEXT NOT NULL DEFAULT '',
    eta         TEXT NOT NULL DEFAULT '',
    output      TEXT NOT NULL DEFAULT '',
    error       TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_jobs_status_created
ON jobs(status, created_at);
