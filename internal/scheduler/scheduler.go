package scheduler

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"pingr/internal/dao"
	"pingr/internal/poll"
	"sync"
	"time"
)

const (

	JobTimeoutDuration = 1 * time.Minute
	DataBaseListenerInterval = 10 * time.Minute

	JobInitialized = 1
	JobRunning = 2
	JobWaiting = 3
	JobClosed = 4
	JobTimedOut = 5
	JobError = 6

	JobInitMessage = "New/Updated job has been initialized."
)

var (
	jobs = make(map[string] dao.Job)
	status = make(map[string]jobStatus)

	notifyDBHandler = make(chan int)
	logChan = make(chan dao.Log, 100)
	newJobChan = make(chan dao.Job, 10)
	deletedJobChan = make(chan string, 10)

    muJobs sync.RWMutex
	muStatus sync.RWMutex
)

type jobStatus struct {
	Status 		int
	CloseChan	chan int
	LastUpdated	time.Time
	Timeout 	time.Duration
}

func Scheduler(db *sql.DB) error {
	log.Info("Starting scheduler")

	initVars(db)


	go dataBaseListener(db) // Best way to handle DB error??

	for {
		select {
			case newJob := <- newJobChan:
				var closeChan chan int
				if _, ok := status[newJob.JobId]; !ok {
					// New job
					closeChan = make(chan int)
					go worker(newJob.JobId, closeChan)
				} else {
					closeChan = status[newJob.JobId].CloseChan
				}
				status[newJob.JobId] = jobStatus{
					Status:		 JobInitialized,
					CloseChan: 	 closeChan,
					LastUpdated: time.Now(),
					Timeout: 	 newJob.Timeout,
				}
				muJobs.Lock()
				jobs[newJob.JobId] = newJob
				muJobs.Unlock()

				log.Info(JobInitMessage + " JobId: " + newJob.JobId)

				l := dao.Log{
					JobId:     newJob.JobId,
					Status:    JobInitialized,
					Message:   JobInitMessage,
					CreatedAt: time.Now(),
				}
				logChan <- l

			case deletedJobId := <- deletedJobChan:
				close(status[deletedJobId].CloseChan)
				muJobs.Lock()
				delete(jobs, deletedJobId)
				muJobs.Unlock()
				delete(status, deletedJobId)
				log.Info(fmt.Sprintf("JobID: %s, Job closed", deletedJobId))
				l := dao.Log{
					JobId:     deletedJobId,
					Status:    JobClosed,
					Message:   "Job closed",
					CreatedAt: time.Now(),
				}
				logChan <- l

		}
	}
}

func initVars(db *sql.DB) {
	jobsDB, err := dao.GetJobs(db)
	if err != nil {
		log.Warn(err)
	}

	log.Info("Starting logger")
	go logger(db)

	muJobs.Lock()
	for _, job := range jobsDB {
		jobs[job.JobId] = job

		closeChan := make(chan int)
		go worker(job.JobId, closeChan)
		status[job.JobId] = jobStatus{
			Status:      JobInitialized,
			CloseChan:	 closeChan,
			LastUpdated: time.Now(),
			Timeout:     job.Timeout,
		}
		log.Info(JobInitMessage + " JobId: " + job.JobId)
		l := dao.Log{
			JobId:     job.JobId,
			Status:    JobInitialized,
			Message:   JobInitMessage,
			CreatedAt: time.Now(),
		}

		logChan <- l
	}
	muJobs.Unlock()
}

func worker(jobId string, close chan int) {
	var job dao.Job
	for {
		muJobs.RLock()
		job = jobs[jobId]
		muJobs.RUnlock()

		
		log.Info(fmt.Sprintf("Starting job. JobID: %s, Type: %s", job.JobId, job.TestType))
		l := dao.Log{
			JobId:     job.JobId,
			Status:    JobRunning,
			Message:   "Job started",
			CreatedAt: time.Now(),
		}
		logChan <- l

		var err error
		if job.TestType == "TLS" {
			err = poll.TLS(job.Url, "HTTPS") // Handle errors more than just logging

		}
		if err != nil {
			// Handle error more than log it
			log.Info(fmt.Sprintf("JobID: %s has failed with error: %s", job.JobId, err))
		} else {
			log.Info(fmt.Sprintf("JobID: %s has succeded, sleeping %d seconds", job.JobId, job.Interval))
		}

		l = dao.Log{
			JobId:     job.JobId,
			Status:    JobWaiting,
			Message:   "Job finished, sleeping",
			CreatedAt: time.Now(),
		}
		logChan <- l

		select {
		case <-time.After(job.Interval * time.Second):
		case <-close:
			return
		}
	}
}

func logger(db *sql.DB) {
	for {
		select {
		case l := <-logChan:
			err := dao.PostLog(l, db)
			if err != nil {
				log.Warn(err)
			}
		}
	}
}

func jobTimeoutHandler() {

}

func Notify() {
	notifyDBHandler <- 1
}

func dataBaseListener(db *sql.DB) {
	for {
		select {
			case <-time.After(DataBaseListenerInterval):
			case <-notifyDBHandler:
		}
		log.Info("Looking for new/updated jobs in DB")
		newJobs, err := dao.GetJobs(db)
		if err != nil {
			continue // Bad
		}

		// Look for new/updated job
		muJobs.RLock()
		for _, j := range newJobs {
			if jobs[j.JobId] != j { // Found a new/changed job
				newJobChan <- j
			}
		}

		// Fairly slow but probably okay since user wont update
		// jobs that often, ALSO: Potential deadlock

		// Look for deleted jobs
		for jobId := range jobs {
			if !stringInSlice(jobId, newJobs) {
				deletedJobChan <- jobId
			}
		}
		muJobs.RUnlock()
	}
}

func stringInSlice(a string, list []dao.Job) bool {
	for _, b := range list {
		if b.JobId == a {
			return true
		}
	}
	return false
}