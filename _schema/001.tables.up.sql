-- name: create-jobs-table
CREATE TABLE IF NOT EXISTS jobs (
    JobId INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    TestType VARCHAR(255),
    Url VARCHAR(255),
    Interval INTERVAL,
    CreatedAt DATE
);

-- name: create-logs-table
CREATE TABLE IF NOT EXISTS logs (
    LogId INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    JobId INTEGER,
    Status INTEGER(1),
    Message VARCHAR(255),
    CreatedAt DATE,
    FOREIGN KEY (JobId)
        REFERENCES jobs (JobId)
);