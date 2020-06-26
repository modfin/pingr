package dao

import "time"

type Log struct {
	LogId     uint64
	JobId     uint64
	Status    int
	Message   string
	CreatedAt time.Time
}

type Job struct {
	JobId		uint64
	TestType 	string
	Url 		string
	Interval 	time.Duration
	Timeout 	time.Duration
	CreatedAt	time.Time
}

func (j *Job) Validate(idReq bool) bool {
	if idReq && j.JobId == 0 {return false}
	if j.TestType == "" {return false}
	if j.Url == "" {return false}
	if j.Interval == 0 {return false}
	if j.Timeout == 0 {return false}
	return true
}