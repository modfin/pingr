package dao

import "time"

type Log struct {
	LogId     string
	JobId     string
	Status    int
	Message   string
	CreatedAt time.Time
}

type Job struct {
	JobId		string
	TestType 	string
	Url 		string
	Interval 	time.Duration
	Timeout 	time.Duration
	CreatedAt	time.Time
}