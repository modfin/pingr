-- name: create-jobs-table
CREATE TABLE IF NOT EXISTS jobs (
    job_id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    test_type TEXT CHECK( test_type IN (
                                                'HTTP',
                                                'Prometheus',
                                                'TLS',
                                                'DNS',
                                                'Ping',
                                                'SSH',
                                                'TCP'
                                            )
                                ),
    url TEXT NOT NULL,
    interval INTEGER NOT NULL,
    timeout  INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    blob BLOB
);

-- name: create-logs-table
CREATE TABLE IF NOT EXISTS logs (
    log_id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    job_id INTEGER NOT NULL,
    status INTEGER NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (job_id)
        REFERENCES jobs (job_id)
);