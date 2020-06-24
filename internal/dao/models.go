package dao

import "time"

type Log struct {
	LogId	string
	JobId	string
	Status	bool
	Message string
	CreatedAt	time.Time
}

type Job struct {
	JobId		string
	TestType 	string
	Url 		string
	Interval 	time.Duration
	CreatedAt	time.Time
}