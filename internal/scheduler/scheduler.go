package scheduler

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"pingr"
	"pingr/internal/dao"
	"reflect"
	"sync"
	"time"
)

const (
	JobTimeoutCheckInterval = 1 * time.Minute // Put in config?
	DataBaseListenerInterval = 10 * time.Minute

	// Change in DB as well
	Successful 	uint 	= 1
	Error     	uint	= 2
	TimedOut  	uint	= 3
	Running		uint	= 4
	Initialized	uint	= 5
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

	Status 		uint
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

				jobId := newJob.Get().JobId
				jobs[jobId] = newJob

				if _, ok := status[newJob.Get().JobId]; !ok {
					// New job
					setJobStatus(newJob, Initialized)
					go worker(jobId, status[jobId].CloseChan)
				} else {
					setJobStatus(newJob, Initialized)
				}

				muJobs.Unlock()

			case deletedJobId := <- deletedJobChan:
				muJobs.Lock()

				if _, ok := status[deletedJobId]; ok {
					close(status[deletedJobId].CloseChan)
					delete(status, deletedJobId)
				}
				if _, ok := jobs[deletedJobId]; ok {
					delete(jobs, deletedJobId)
					log.Info(fmt.Sprintf("JobID: %d, Job deleted", deletedJobId))
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
			// Job deleted?
			muJobs.RUnlock()
			return
		}
		muJobs.RUnlock()

		setJobStatus(job, Running)

		rt, err := job.RunTest()

		muJobs.RLock()

		if _, ok = status[jobId]; !ok {
			// Has the job been deleted while it was running?
			muJobs.RUnlock()
			return
		}

		if err != nil {
			// TODO: Handle error better
			setJobStatus(job, Error)
			addJobLog(jobId, Error, rt, err)
		} else {
			setJobStatus(job, Successful)
			addJobLog(jobId, Successful, rt, err)
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
		case l := <-logChan: // maybe loop until we added log?
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
			if jStatus.Status == Running {
				timeOut := jStatus.LastUpdated.Add(time.Second*jStatus.Timeout)
				if time.Now().After(timeOut) {
					// Job has timed out
					// TODO: Handle error better
					muJobs.RLock()
					setJobStatus(jobs[jobId], TimedOut)
					muJobs.RUnlock()
					addJobLog(jobId, TimedOut, time.Now().Sub(jStatus.LastUpdated), nil)
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
		jobId := job.Get().JobId
		jobs[jobId] = job

		setJobStatus(job, Initialized)
		go worker(jobId, status[jobId].CloseChan)
	}
}

func setJobStatus(job pingr.Job, statusCode uint) {
	jobId := job.Get().JobId
	var jStatus *jobStatus
	if _, ok := status[jobId]; !ok { // new job
		jStatus = &jobStatus{ CloseChan: make(chan int), Timeout: job.Get().Timeout}
		status[jobId] = jStatus
	} else {
		jStatus = status[jobId]
	}
	jStatus.Mu.Lock()
	jStatus.Status = statusCode
	jStatus.LastUpdated = time.Now()
	jStatus.Mu.Unlock()
}

func addJobLog(jobId uint64, statusCode uint, rt time.Duration, err error) {
	var message string
	if err != nil {
		message = err.Error()
	}
	log.Info(fmt.Sprintf("JobID: %d, StatusCode: %d", jobId, statusCode))
	l := pingr.Log{
		JobId:     		jobId,
		StatusId:  		statusCode,
		Message:   		message,
		ResponseTime: 	rt,
		CreatedAt: 		time.Now(),
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