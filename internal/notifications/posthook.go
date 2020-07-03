package notifications

import (
	"bytes"
	"encoding/json"
	"net/http"
	"pingr"
	"time"
)

type notificationInformation struct {
	JobId 			uint64 			`json:"job_id"`
	JobName 		string 			`json:"job_name"`
	Url 			string 			`json:"url"`
	TestType 		string 			`json:"test_type"`
	StatusCode 		uint 			`json:"status_code"`
	StatusName 		string 			`json:"status_name"`
	Message			string 			`json:"message"`
	Interval 		time.Duration	`json:"interval"`
}

func PostHook(urls []string,  job pingr.BaseJob, jobErr error, statusCode uint) error {
	postMsg := notificationInformation{
		JobId:      job.JobId,
		JobName:    job.JobName,
		Url:        job.Url,
		TestType:   job.TestType,
		StatusCode: statusCode,
		Interval:   job.Interval,
	}
	if jobErr != nil {
		postMsg.Message = jobErr.Error()
	}
	if statusCode == 2 {
		postMsg.StatusName = "Test failure"
	} else if statusCode == 3 {
		postMsg.StatusName = "Test timed out"
	} else if statusCode == 1 {
		postMsg.StatusName = "Test successful"
	}

	client := http.Client{Timeout: 20 * time.Second}

	marshalPostMsg, err := json.Marshal(postMsg)
	if err != nil {
		return err
	}

	for _, url := range urls {
		_, err := client.Post(url, "application/json", bytes.NewBuffer(marshalPostMsg))
		if err != nil {
			return err
		}
	}
	return nil
}
