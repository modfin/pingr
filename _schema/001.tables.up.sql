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
    status_id INTEGER NOT NULL,
    message TEXT,
    response_time INTEGER,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (job_id)
        REFERENCES jobs (job_id)
    FOREIGN KEY (status_id)
        REFERENCES status_map (status_id)
);

-- name: create-status-map-table
CREATE TABLE IF NOT EXISTS status_map (
    status_id  INTEGER PRIMARY KEY NOT NULL,
    status_name TEXT NOT NULL,
    UNIQUE (status_id, status_name)
);

-- name: create-contacts-table
CREATE TABLE IF NOT EXISTS contacts (
    contact_id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    contact_type TEXT NOT NULL, -- smtp OR endpoint
    contact_url TEXT NOT NULL -- example@gmail.com OR callr.modfin.se
)

-- name: create-job-contact-mapper
CREATE TABLE IF NOT EXISTS job_contact (
    contact_id INTEGER NOT NULL,
    job_id INTEGER NOT NULL
)

-- name: init-status-mapper
INSERT INTO status_map(status_id, status_name)
VALUES
    (1, "Successful"),
    (2, "Error"),
    (3, "TimedOut");
