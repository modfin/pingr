package scheduler

import (
	"errors"
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
	Successful 	uint 	= 1
	Error     	uint	= 2
	TimedOut  	uint	= 3
	Running		uint	= 4
	Initialized	uint	= 5
	Deleted 	uint    = 6
)


var (
	muJobs 				sync.RWMutex

	jobs 				= make(map[uint64] pingr.Job)
	closeChans 			= make(map[uint64] chan int)

	newJobChan          = make(chan pingr.Job, 10) // To be discussed
	deletedJobChan      = make(chan uint64, 10)
	statusChan          = make(chan jobStatus, 10)
	logChan             = make(chan pingr.Log)
	notifyDBHandler     = make(chan int, 10)
	timeForTimeoutCheck = make(chan int)
)

type jobStatus struct {
	JobId		uint64
	Status 		uint
	Error		error
}

func Scheduler(db *sqlx.DB) {
	go logger(db)

	initVars(db)

	go dataBaseListener(db) // Best way to handle DB error??

	go jobMaintainer(db)

	for {
		select {
			case newJob := <- newJobChan:
				muJobs.Lock()
				jobId := newJob.Get().JobId
				jobs[jobId] = newJob
				muJobs.Unlock()

				if _, ok := closeChans[jobId]; ok {
					close(closeChans[jobId])
					delete(closeChans, jobId)
				}

				addJobLog(jobId, Initialized, 0, nil)
				statusChan <- jobStatus{jobId, Initialized, nil}

				closeChans[jobId] = make(chan int)
				go worker(newJob, closeChans[jobId])

			case deletedJobId := <- deletedJobChan:
				if _, ok := closeChans[deletedJobId]; ok {
					close(closeChans[deletedJobId])
					delete(closeChans, deletedJobId)
				}
				addJobLog(deletedJobId, Deleted, 0, nil)
				statusChan <- jobStatus{deletedJobId, Deleted, nil}

				muJobs.Lock()
				if _, ok := jobs[deletedJobId]; ok {
					delete(jobs, deletedJobId)
					log.Info(fmt.Sprintf("JobID: %d, Job deleted", deletedJobId))
				}
				muJobs.Unlock()
			case <-time.After(JobTimeoutCheckInterval):
				timeForTimeoutCheck <- 1
		}
	}
}

func worker(job pingr.Job, close chan int) {
	jobId := job.Get().JobId
	for {
		select {
		case <-close:
			return
		default:
			statusChan <- jobStatus{jobId, Running,nil}
		}

		rt, err := job.RunTest()

		select {
		case <-close:
			return
		default:
			statusCode := Successful
			if err != nil {
				statusCode = Error
			}
			addJobLog(jobId, statusCode, rt, err)
			statusChan <- jobStatus{jobId, statusCode, err}
		}

		select {
		case <-time.After(job.Get().Interval * time.Second):
		case <-close:
			return
		}
	}
}

func jobMaintainer(db *sqlx.DB) {
	jobStatusMap := make(map[uint64]uint)
	lastUpdated := make(map[uint64]time.Time)
	consecutiveErrors := make(map[uint64]uint)
	jobContactsNotified := make(map[uint64][]uint64)
	for {
		select {
		case status := <-statusChan:
			id := status.JobId
			switch status.Status {
			case Successful: // sometimes check!!
				jobStatusMap[id] = Successful
				consecutiveErrors[id] = 0
				if _, ok := jobContactsNotified[id]; ok {
					// This is an active incident, notify that test is successful again
					muJobs.RLock()
					// Check that job still exists
					if _, ok := jobs[id]; !ok {
						muJobs.RUnlock()
						break
					}
					var mailsToContact []string
					var postHooksToSend []string
					for _, contactId := range jobContactsNotified[id] {
						// Acquire contacted contacts
						contact, err := dao.GetContact(contactId, db)
						if err != nil {
							log.Error(err)
						}
						switch contact.ContactType {
						case "smtp":
							mailsToContact = append(mailsToContact, contact.ContactUrl)
						case "http":
							postHooksToSend = append(postHooksToSend, contact.ContactUrl)
						}
					}
					var emailErr, postHookErr error
					if len(mailsToContact) != 0 {
						emailErr = notifications.SendEmail(mailsToContact, jobs[id].Get(), status.Error, db)
					}
					if len(postHooksToSend) != 0 {
						postHookErr = notifications.PostHook(postHooksToSend, jobs[id].Get(), status.Error, Successful)
					}
					if emailErr == nil && postHookErr == nil {
						// Keep sending until no errors occur
						delete(jobContactsNotified, id)
					}
					muJobs.RUnlock()
				}
			case Error:
				jobStatusMap[id] = Error
				consecutiveErrors[id]++
				jobContactTypes, err := dao.GetJobContactsType(id, db)
				if err != nil {
					log.Error(err)
				}
				muJobs.RLock()
				// Check that job still exists
				if _, ok := jobs[id]; !ok {
					muJobs.RUnlock()
					break
				}
				jobContactsNotified[id] = notifyContactsError(id, Error, status.Error,
											consecutiveErrors[id], jobContactTypes,
											jobContactsNotified[id], db)
				muJobs.RUnlock()
			case Running:
				jobStatusMap[id] = Running
				lastUpdated[id] = time.Now()
			case Initialized:
				jobStatusMap[id] = Initialized
				consecutiveErrors[id] = 0
			case Deleted:
				delete(jobStatusMap, id)
				delete(lastUpdated, id)
				delete(consecutiveErrors, id)
			default:
				log.Error("invalid status code sent on statusChan")
			}
		case <-timeForTimeoutCheck:
			log.Info("looking for timed out jobs")
			for id, statusCode := range jobStatusMap {
				if statusCode == Running {
					muJobs.RLock()
					if runningJob, ok := jobs[id]; ok {
						jobTimeout := runningJob.Get().Timeout*2
						if time.Now().After(lastUpdated[id].Add(jobTimeout)) {
							jobStatusMap[id] = TimedOut
							consecutiveErrors[id]++
							jobError := errors.New("job timed out - found while manually checking for timeouts")
							addJobLog(id, TimedOut, jobTimeout, jobError)
							jobContactTypes, err := dao.GetJobContactsType(id, db)
							if err != nil {
								log.Error(err)
							}
							// Check that job still exists
							if _, ok := jobs[id]; !ok {
								muJobs.RUnlock()
								continue
							}
							jobContactsNotified[id] = notifyContactsError(id, TimedOut, jobError,
														consecutiveErrors[id], jobContactTypes,
														jobContactsNotified[id], db)
							newJobChan <- jobs[id] // Might not be the right way of doing it
							// could cause an infinite amount of goroutines to get stuck
						}
					}
					muJobs.RUnlock()
				}
			}
		}
	}
}

func notifyContactsError(
		jobId uint64,
		statusCode uint,
		jobError error,
		consecutiveErrors uint,
		jobContacts []dao.JobContactType,
		jobContactsNotified[]uint64,
		db *sqlx.DB,
	)[]uint64 {

	var mailsToContact []string
	var mailContactIds []uint64
	var postHooksToSend []string
	var hookContactIds []uint64

	for _, contactType := range jobContacts {
		// Have we reached the threshold for contact?
		if consecutiveErrors >= contactType.Threshold {
			// Check if contact has already been contacted
			contacted := false
			for _, contactId := range jobContactsNotified {
				if contactId == contactType.ContactId {
					contacted = true
				}
			}
			// If not contacted, determine action to be taken
			if !contacted {
				switch contactType.ContactType {
				case "smtp":
					mailContactIds = append(mailContactIds, contactType.ContactId)
					mailsToContact = append(mailsToContact, contactType.ContactUrl)
				case "http":
					hookContactIds = append(hookContactIds, contactType.ContactId)
					postHooksToSend = append(postHooksToSend, contactType.ContactUrl)
				}
			}
		}
	}
	// Notify contacts
	if len(mailsToContact) != 0 {
		err := notifications.SendEmail(mailsToContact, jobs[jobId].Get(), jobError, db)
		if err != nil {
			log.Error("unable to send email")
		} else {
			jobContactsNotified = append(jobContactsNotified, mailContactIds...)
		}
	}
	if len(postHooksToSend) != 0 {
		err := notifications.PostHook(postHooksToSend, jobs[jobId].Get(), jobError, statusCode)
		if err != nil {
			log.Error("unable to post hook")
		} else {
			jobContactsNotified = append(jobContactsNotified, hookContactIds...)
		}
	}
	return jobContactsNotified
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

		addJobLog(jobId, Initialized, 0, nil)
		statusChan <- jobStatus{jobId, Initialized, nil}

		closeChans[jobId] = make(chan int)
		go worker(job, closeChans[jobId])
	}
}

func addJobLog(jobId uint64, statusCode uint, rt time.Duration, err error) {
	var logMessage string
	if err != nil {
		logMessage = err.Error()
	}
	log.Info(fmt.Sprintf("JobID: %d, StatusCode: %d", jobId, statusCode))
	l := pingr.Log{
		JobId:     		jobId,
		StatusId:  		statusCode,
		Message:   		logMessage,
		ResponseTime: 	rt,
		CreatedAt: 		time.Now(),
	}
	logChan <- l
}

func NotifyNewJob() {
	notifyDBHandler <- 1
}

func NotifyUpdatedJob(job pingr.Job) {
	newJobChan <- job
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