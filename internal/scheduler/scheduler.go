package scheduler

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"pingr"
	"pingr/internal/dao"
	"pingr/internal/notifications"
	"reflect"
	"sync"
	"time"
)

const (
	JobTimeoutCheckInterval = 1 * time.Minute // Put in config?
	DataBaseListenerInterval = 10 * time.Minute

	// Change in DB as well
	JobInitialized = 1
	JobRunning = 2
	JobWaiting = 3
	JobClosed = 4
	JobTimedOut = 5
	JobError = 6
)

var (
	jobs = make(map[uint64] pingr.Job)
	status = make(map[uint64]*jobStatus)

	logChan = make(chan pingr.Log, 100)
	newJobChan = make(chan pingr.Job, 10)
	deletedJobChan = make(chan uint64, 10)
	notifyDBHandler = make(chan int, 10)

    muJobs sync.RWMutex
)

type jobStatus struct {
	Mu 			sync.RWMutex

	Status 		int
	CloseChan	chan int
	LastUpdated	time.Time
	Timeout 	time.Duration
}

func Scheduler(db *sqlx.DB) {
	go logger(db)

	initVars(db)

	go dataBaseListener(db) // Best way to handle DB error??

	go jobTimeoutHandler()

	for {
		select {
			case newJob := <- newJobChan:
				muJobs.Lock()
				jobs[newJob.Get().JobId] = newJob

				if _, ok := status[newJob.Get().JobId]; !ok {
					// New job
					closeChan := make(chan int)
					setNewJobInitialized(newJob, closeChan)
					go worker(newJob.Get().JobId, closeChan, db)
				} else {
					setUpdatedJobInitialized(newJob.Get().JobId)
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
					l := pingr.Log{
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

func worker(jobId uint64, close chan int, db *sqlx.DB) {
	for {
		muJobs.RLock()
		job, ok := jobs[jobId]
		if !ok {
			// Job deleted?
			muJobs.RUnlock()
			return
		}
		muJobs.RUnlock()

		setJobRunning(job)

		_, err := job.RunTest()

		muJobs.RLock()

		if _, ok = status[jobId]; !ok {
			// Has the job been deleted while it was running?
			muJobs.RUnlock()
			return
		}

		if err != nil {
			// TODO: Handle error better
			setJobError(jobId, err)
			notifications.SendEmail(job.Get(), err, db)
		} else {
			setJobWaiting(job)
		}
		muJobs.RUnlock()

		select {
		case <-time.After(job.Get().Interval * time.Second):
		case <-close:
			return
		}
	}
}

func logger(db *sqlx.DB) {
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
	for {
		<-time.After(JobTimeoutCheckInterval)
		log.Info("Checking for timed out jobs")
		for jobId, jStatus := range status {
			jStatus.Mu.RLock()
			if jStatus.Status == JobRunning {
				if time.Now().After(jStatus.LastUpdated.Add(time.Second*jStatus.Timeout)) {
					// Job has timed out
					// TODO: Handle error better
					log.Error(fmt.Sprintf("JobID: %d, Timed out", jobId))
				}
			}
			jStatus.Mu.RUnlock()
		}
	}
}

func dataBaseListener(db *sqlx.DB) {
	for {
		select {
			case <-time.After(DataBaseListenerInterval):
			case <-notifyDBHandler:
		}
		log.Info("Looking for new/updated jobs in DB")
		jobsDB, err := dao.GetJobs(db)
		if err != nil {
			continue // Bad
		}

		var newJobs [] pingr.Job
		var deletedJobs []uint64

		// Look for new/updated job
		muJobs.RLock()
		for _, j := range jobsDB {
			if !reflect.DeepEqual(jobs[j.Get().JobId], j) { // Found a new/updated job
				newJobs = append(newJobs, j)
			}
		}

		// Look for deleted jobs
		for jobId := range jobs {
			// slow but probably okay since user wont update/add that often
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

func initVars(db *sqlx.DB) {
	jobsDB, err := dao.GetJobs(db)
	if err != nil {
		log.Warn(err)
	}

	muJobs.Lock()
	defer muJobs.Unlock()

	for _, job := range jobsDB {
		jobs[job.Get().JobId] = job

		closeChan := make(chan int)
		setNewJobInitialized(job, closeChan)

		go worker(job.Get().JobId, closeChan, db)
	}
}

func setNewJobInitialized(job pingr.Job, closeChan chan int) {
	jStatus := jobStatus{
		Status:      JobInitialized,
		CloseChan:	 closeChan,
		LastUpdated: time.Now(),
		Timeout:     job.Get().Timeout,
	}
	status[job.Get().JobId] = &jStatus

	log.Info(fmt.Sprintf("JobID: %d, Initialized", job.Get().JobId))
	l := pingr.Log{
		JobId:     job.Get().JobId,
		Status:    JobInitialized,
		Message:   "New/Updated job has been initialized.",
		CreatedAt: time.Now(),
	}
	logChan <- l
}

func setUpdatedJobInitialized(jobId uint64) {
	status[jobId].Mu.Lock()
	defer status[jobId].Mu.Unlock()

	status[jobId].LastUpdated = time.Now()
	status[jobId].Status = JobInitialized

	log.Info(fmt.Sprintf("JobID: %d, Initialized", jobId))
	l := pingr.Log{
		JobId:     jobId,
		Status:    JobInitialized,
		Message:   "New/Updated job has been initialized.",
		CreatedAt: time.Now(),
	}
	logChan <- l
}

func setJobRunning(job pingr.Job) {
	status[job.Get().JobId].Mu.Lock()
	defer status[job.Get().JobId].Mu.Unlock()

	status[job.Get().JobId].Status = JobRunning
	status[job.Get().JobId].LastUpdated = time.Now()

	log.Info(fmt.Sprintf("JobID: %d started, Type: %s", job.Get().JobId, job.Get().TestType))
	l := pingr.Log{
		JobId:     job.Get().JobId,
		Status:    JobRunning,
		Message:   "Job started",
		CreatedAt: time.Now(),
	}
	logChan <- l
}

func setJobWaiting(job pingr.Job) {
	status[job.Get().JobId].Mu.Lock()
	defer status[job.Get().JobId].Mu.Unlock()

	status[job.Get().JobId].Status = JobWaiting
	status[job.Get().JobId].LastUpdated = time.Now()

	log.Info(fmt.Sprintf("JobID: %d has succeded, sleeping %d seconds", job.Get().JobId, job.Get().Interval))
	l := pingr.Log{
		JobId:     job.Get().JobId,
		Status:    JobWaiting,
		Message:   "Job finished, sleeping",
		CreatedAt: time.Now(),
	}
	logChan <- l
}

func setJobError(jobId uint64, err error) {
	status[jobId].Mu.Lock()
	defer status[jobId].Mu.Unlock()

	status[jobId].Status = JobError
	status[jobId].LastUpdated = time.Now()

	log.Error(fmt.Sprintf("JobID: %d has failed with error: %s", jobId, err))
	l := pingr.Log{
		JobId:     jobId,
		Status:    JobError,
		Message:   "Job finished, sleeping",
		CreatedAt: time.Now(),
	}
	logChan <- l
}

func NotifyNewJob() {
	notifyDBHandler <- 1
}

func NotifyDeletedJob(jobId uint64){
	deletedJobChan <- jobId
}

func intInSlice(a uint64, list []pingr.Job) bool {
	for _, b := range list {
		if b.Get().JobId == a {
			return true
		}
	}
	return false
}