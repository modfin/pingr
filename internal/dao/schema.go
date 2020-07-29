package dao


const _schema_v0_up = `
CREATE TABLE IF NOT EXISTS _schema( 
    version int PRIMARY KEY,
	created_at TIMESTAMP
);
INSERT INTO _schema(version, created_at) VALUES (0, CURRENT_TIMESTAMP) ON CONFLICT DO NOTHING;
`


const _schema_v1_down = `
-- name: drop-incident-contact-log
DROP TABLE IF EXISTS incident_contact_log ;

-- name: drop-incidents
DROP TABLE IF EXISTS incidents ;

-- name: drop-test-contact-mapper
DROP TABLE IF EXISTS test_contacts ;

-- name: drop-contacts-table
DROP TABLE IF EXISTS contacts ;

-- name: drop-status-map-table
DROP TABLE IF EXISTS status_map ;

-- name: drop-logs-table
DROP TABLE IF EXISTS logs ;

-- name: -tests-table
DROP TABLE IF EXISTS tests ;

`

const _schema_v1_up = `
-- name: create-tests-table
CREATE TABLE IF NOT EXISTS tests (
    test_id TEXT PRIMARY KEY,
    test_name TEXT NOT NULL,
    test_type TEXT CHECK( test_type IN (
                                                'HTTP',
                                                'Prometheus',
                                                'TLS',
                                                'DNS',
                                                'Ping',
                                                'SSH',
                                                'TCP',
                                                'HTTPPush',
                                                'PrometheusPush'
                                            )
                                ),
    url TEXT NOT NULL,
    interval INTEGER NOT NULL,
    timeout  INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
	active INTEGER NOT NULL,
    blob BLOB
)
;

-- name: create-logs-table
CREATE TABLE IF NOT EXISTS logs (
    log_id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    test_id TEXT NOT NULL,
    status_id INTEGER NOT NULL,
    message TEXT,
    response_time INTEGER,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (test_id)
        REFERENCES tests (test_id),
    FOREIGN KEY (status_id)
        REFERENCES status_map (status_id)
)
;

-- name: create-status-map-table
CREATE TABLE IF NOT EXISTS status_map (
    status_id  INTEGER PRIMARY KEY NOT NULL,
    status_name TEXT NOT NULL,
    UNIQUE (status_id, status_name)
)
;

-- name: create-contacts-table
CREATE TABLE IF NOT EXISTS contacts (
    contact_id TEXT NOT NULL,
    contact_name TEXT NOT NULL,
    contact_type TEXT NOT NULL,
    contact_url TEXT NOT NULL
)
;
-- name: create-test-contact-mapper
CREATE TABLE IF NOT EXISTS test_contacts (
    test_id TEXT NOT NULL,
    contact_id TEXT NOT NULL,
    threshold INTEGER NOT NULL,
    UNIQUE (contact_id, test_id),
    FOREIGN KEY (test_id)
        REFERENCES tests (test_id),
    FOREIGN KEY (contact_id)
        REFERENCES contacts (contact_id)
);

-- name: init-status-mapper
INSERT INTO status_map(status_id, status_name)
VALUES
    (1, "Successful"),
    (2, "Error"),
    (3, "TimedOut"),
    (5, "Initialized"),
    (6, "Paused")
;

-- name: create-incidents
CREATE TABLE IF NOT EXISTS incidents (
    incident_id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    test_id TEXT NOT NULL,
    active INTEGER NOT NULL,
	root_cause TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
	closed_at TIMESTAMP,
    FOREIGN KEY (test_id)
        REFERENCES tests (test_id)
);

-- name: create-incident_contact_log
CREATE TABLE IF NOT EXISTS incident_contact_log (
    incident_id INTEGER,
    contact_id TEXT NOT NULL,
	message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    FOREIGN KEY (contact_id)
        REFERENCES contacts (contact_id),
    FOREIGN KEY (incident_id)
        REFERENCES incidents (incident_id)
);

`