package dao

import "database/sql"

func GetLogs(db *sql.DB) ([]Log, error) {
	q := `
		SELECT * FROM logs
	`
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var l Log
		err := rows.Scan(&l.LogId, &l.JobId, &l.Status, &l.Message, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func GetLog(id string, db *sql.DB) (*Log, error) {
	q := `
		SELECT * FROM logs WHERE LogId = ?
	`
	stmt, err := db.Prepare(q)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var l Log
	err = stmt.QueryRow(id).Scan(&l.LogId, &l.JobId, &l.Status, &l.Message, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}