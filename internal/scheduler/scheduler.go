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
	JobTimeoutCheckInterval = 1 * time.Minute
	DataBaseListenerInterval = 10 * time.Minute

	JobInitialized = 1
	JobRunning = 2
	JobWaiting = 3
	JobClosed = 4
	JobTimedOut = 5
	JobError = 6
)

var (
	jobs = make(map[uint64] dao.Job)
	status = make(map[uint64]*jobStatus)

	logChan = make(chan dao.Log, 100)
	newJobChan = make(chan dao.Job, 10)
	deletedJobChan = make(chan uint64, 10)

    muJobs sync.RWMutex
)

type jobStatus struct {
	Status 		int
	CloseChan	chan int
	LastUpdated	time.Time
	Timeout 	time.Duration
	Mu 			sync.RWMutex
}

func Scheduler(db *sql.DB) error {
	log.Info("Starting scheduler")

	go logger(db)

	initVars(db)

	go dataBaseListener(db) // Best way to handle DB error??

	go jobTimeoutHandler()

	for {
		select {
			case newJob := <- newJobChan:
				// New/updated job has to be processed
				muJobs.Lock()
				jobs[newJob.JobId] = newJob

				if _, ok := status[newJob.JobId]; !ok {
					// New job
					closeChan := make(chan int)
					setNewJobInitialized(newJob, closeChan)
					go worker(newJob.JobId, closeChan)
				} else {
					setUpdatedJobInitialized(newJob.JobId)
				}
				muJobs.Unlock()

			case deletedJobId := <- deletedJobChan:
				// Deleted job has to be processed
				muJobs.Lock()
				if _, ok := status[deletedJobId]; ok {
					close(status[deletedJobId].CloseChan)
					delete(status, deletedJobId)
				}
				if _, ok := jobs[deletedJobId]; ok {
					delete(jobs, deletedJobId)
					log.Info(fmt.Sprintf("JobID: %d, Job closed", deletedJobId))
					l := dao.Log{
						JobId:     deletedJobId,
						Status:    JobClosed,
						Message:   "Job closed",
						CreatedAt: time.Now(),
					}
					logChan <- l
				}
				muJobs.Unlock()
		}
	}
}

func worker(jobId uint64, close chan int) {
	for {
		muJobs.RLock()
		job, ok := jobs[jobId]
		if !ok {
			muJobs.RUnlock()
			return
		}

		_, ok = status[jobId]
		if !ok {
			muJobs.RUnlock()
			return
		}

		setJobRunning(job)

		var err error
		switch job.TestType {
		case "TLS":
			err = poll.TLS(job.Url, "HTTPS")
		}

		muJobs.RLock()
		_, ok = status[jobId]
		if !ok {
			// Has the job been deleted while it was running?
			muJobs.RUnlock()
			return
		}

		if err != nil {
			// TODO: Handle error better
			setJobError(jobId, err)
		} else {
			setJobWaiting(job)
		}
		muJobs.RUnlock()

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
	time.Sleep(JobTimeoutCheckInterval)
	for {
		log.Info("Checking for timed out jobs")
		for jobId, jStatus := range status {
			if jStatus.Status == JobRunning {
				jStatus.Mu.RLock()
				if time.Now().After(jStatus.LastUpdated.Add(time.Second*jStatus.Timeout)) {
					// Job has timed out
					// TODO: Handle error better
					log.Error(fmt.Sprintf("JobID: %d, Timed out", jobId))
				}
				jStatus.Mu.RUnlock()
			}
		}
		time.Sleep(JobTimeoutCheckInterval)
	}
}

func dataBaseListener(db *sql.DB) {
	for {
		select {
			case <-time.After(DataBaseListenerInterval):
				log.Info("Looking for new/updated jobs in DB")
				jobsDB, err := dao.GetJobs(db)
				if err != nil {
					continue // Bad
				}

				newJobs := make([]dao.Job, 0)
				deletedJobs := make([]uint64, 0)

				// Look for new/updated job
				muJobs.RLock()
				for _, j := range jobsDB {
					if jobs[j.JobId] != j { // Found a new/updated job
						newJobs = append(newJobs, j)
					}
				}

				// Look for deleted jobs
				for jobId := range jobs {
					// slow but probably okay since user wont update/add
					if !intInSlice(jobId, jobsDB) {
						deletedJobs = append(deletedJobs, jobId)
					}
				}
				muJobs.RUnlock()

				// Save result in array and send on channel after Unlock to avoid deadlock
				for _, newJob := range newJobs {
					newJobChan <- newJob
				}
				for _, jobId := range deletedJobs {
					deletedJobChan <- jobId
				}
		}
	}
}

func initVars(db *sql.DB) {
	jobsDB, err := dao.GetJobs(db)
	if err != nil {
		log.Warn(err)
	}

	muJobs.Lock()
	for _, job := range jobsDB {
		jobs[job.JobId] = job

		closeChan := make(chan int)
		setNewJobInitialized(job, closeChan)

		go worker(job.JobId, closeChan)
	}
	muJobs.Unlock()
}

func setNewJobInitialized(job dao.Job, closeChan chan int) {
	jStatus := jobStatus{
		Status:      JobInitialized,
		CloseChan:	 closeChan,
		LastUpdated: time.Now(),
		Timeout:     job.Timeout,
	}
	status[job.JobId] = &jStatus

	log.Info(fmt.Sprintf("JobID: %d, Initialized", job.JobId))
	l := dao.Log{
		JobId:     job.JobId,
		Status:    JobInitialized,
		Message:   "New/Updated job has been initialized.",
		CreatedAt: time.Now(),
	}
	logChan <- l
}

func setUpdatedJobInitialized(jobId uint64) {
	status[jobId].LastUpdated = time.Now()
	status[jobId].Status = JobInitialized

	log.Info(fmt.Sprintf("JobID: %d, Initialized", jobId))
	l := dao.Log{
		JobId:     jobId,
		Status:    JobInitialized,
		Message:   "New/Updated job has been initialized.",
		CreatedAt: time.Now(),
	}
	logChan <- l
}

func setJobRunning(job dao.Job) {
	status[job.JobId].Mu.Lock()
	status[job.JobId].Status = JobRunning
	status[job.JobId].LastUpdated = time.Now()
	status[job.JobId].Mu.Unlock()
	muJobs.RUnlock()

	log.Info(fmt.Sprintf("Starting job. JobID: %d Type: %s", job.JobId, job.TestType))
	l := dao.Log{
		JobId:     job.JobId,
		Status:    JobRunning,
		Message:   "Job started",
		CreatedAt: time.Now(),
	}
	logChan <- l
}

func setJobWaiting(job dao.Job) {
	status[job.JobId].Mu.Lock()
	status[job.JobId].Status = JobWaiting
	status[job.JobId].LastUpdated = time.Now()
	status[job.JobId].Mu.Unlock()
	log.Info(fmt.Sprintf("JobID: %d has succeded, sleeping %d seconds", job.JobId, job.Interval))
	l := dao.Log{
		JobId:     job.JobId,
		Status:    JobWaiting,
		Message:   "Job finished, sleeping",
		CreatedAt: time.Now(),
	}
	logChan <- l
}

func setJobError(jobId uint64, err error) {
	status[jobId].Mu.Lock()
	status[jobId].Status = JobError
	status[jobId].LastUpdated = time.Now()
	status[jobId].Mu.Unlock()
	log.Error(fmt.Sprintf("JobID: %d has failed with error: %s", jobId, err))
	l := dao.Log{
		JobId:     jobId,
		Status:    JobError,
		Message:   "Job finished, sleeping",
		CreatedAt: time.Now(),
	}
	logChan <- l
}

func NotifyNewJob(job dao.Job) {
	newJobChan <- job
}

func NotifyDeletedJob(jobId uint64){
	deletedJobChan <- jobId
}

func intInSlice(a uint64, list []dao.Job) bool {
	for _, b := range list {
		if b.JobId == a {
			return true
		}
	}
	return false
}