package scheduler

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"math"
	"os"
	"pingr"
	"pingr/internal/config"
	"pingr/internal/dao"
	"pingr/internal/notifications"
	"reflect"
	"sync"
	"syscall"
	"time"
)

const (
	TestTimeoutCheckInterval = 1 * time.Minute // Put in config?
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
	muTests 			sync.RWMutex

	tests 				= make(map[string] pingr.Test)
	closeChans 			= make(map[string] chan int)

	newTestChan         = make(chan pingr.Test, 10) // To be discussed
	deletedTestChan     = make(chan string, 10)
	statusChan          = make(chan testStatus, 10)
	timeForTimeoutCheck = make(chan int)
)

type testStatus struct {
	TestId		string
	Status 		uint
	Error		error
}

func Scheduler(db *sqlx.DB) {
	go discSpaceMaintainer(db)

	go dataBaseListener(db) // Best way to handle DB error??

	go testMaintainer(db)

	initVars(db)

	for {
		select {
			case newTest := <- newTestChan:
				muTests.Lock()
				testId := newTest.Get().TestId
				tests[testId] = newTest
				muTests.Unlock()

				if _, ok := closeChans[testId]; ok {
					close(closeChans[testId])
					delete(closeChans, testId)
				}

				addTestLog(testId, Initialized, 0, nil, db)
				statusChan <- testStatus{testId, Initialized, nil}

				closeChans[testId] = make(chan int)
				go worker(newTest, closeChans[testId], db)

			case deletedTestId := <- deletedTestChan:
				if c, ok := closeChans[deletedTestId]; ok {
					close(c)
					delete(closeChans, deletedTestId)
				}
				addTestLog(deletedTestId, Deleted, 0, nil, db)
				statusChan <- testStatus{deletedTestId, Deleted, nil}

				pingr.MuPush.Lock()
				if c, ok := pingr.PushChans[deletedTestId]; ok {
					close(c)
					delete(pingr.PushChans, deletedTestId)
				}
				pingr.MuPush.Unlock()

				muTests.Lock()
				if _, ok := tests[deletedTestId]; ok {
					delete(tests, deletedTestId)
					log.Info(fmt.Sprintf("TestID: %d, Test deleted", deletedTestId))
				}
				muTests.Unlock()
			case <-time.After(TestTimeoutCheckInterval):
				timeForTimeoutCheck <- 1
		}
	}
}

func worker(test pingr.Test, close chan int, db *sqlx.DB) {
	testId := test.Get().TestId
	for {
		select {
		case <-close:
			return
		default:
			statusChan <- testStatus{testId, Running,nil}
		}

		rt, err := test.RunTest()

		select {
		case <-close:
			return
		default:
			statusCode := Successful
			if err != nil {
				statusCode = Error
			}
			addTestLog(testId, statusCode, rt, err, db)
			statusChan <- testStatus{testId, statusCode, err}
		}

		select {
		case <-time.After(test.Get().Interval * time.Second):
		case <-close:
			return
		}
	}
}

func testMaintainer(db *sqlx.DB) {
	testStatusMap := make(map[string]uint)
	lastUpdated := make(map[string]time.Time)
	consecutiveErrors := make(map[string]uint)
	testContactsNotified := make(map[string][]string)
	for {
		select {
		case status := <-statusChan:
			id := status.TestId
			switch status.Status {
			case Successful:
				testStatusMap[id] = Successful
				consecutiveErrors[id] = 0
				if _, ok := testContactsNotified[id]; ok {
					// This is an active incident, notify that test is successful again
					muTests.RLock()
					// Check that test still exists
					if _, ok := tests[id]; !ok {
						muTests.RUnlock()
						break
					}
					var mailsToContact []string
					var postHooksToSend []string
					for _, contactId := range testContactsNotified[id] {
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
						emailErr = notifications.SendEmail(mailsToContact, tests[id].Get(), status.Error, db)
					}
					if len(postHooksToSend) != 0 {
						postHookErr = notifications.PostHook(postHooksToSend, tests[id].Get(), status.Error, Successful)
					}
					if emailErr == nil && postHookErr == nil {
						// Keep sending until no errors occur
						delete(testContactsNotified, id)
					}
					muTests.RUnlock()
				}
			case Error:
				testStatusMap[id] = Error
				consecutiveErrors[id]++
				testContactTypes, err := dao.GetTestContactsType(id, db)
				if err != nil {
					log.Error(err)
				}
				muTests.RLock()
				// Check that test still exists
				if _, ok := tests[id]; !ok {
					muTests.RUnlock()
					break
				}
				testContactsNotified[id] = notifyContactsError(id, Error, status.Error,
											consecutiveErrors[id], testContactTypes,
											testContactsNotified[id], db)
				muTests.RUnlock()
			case Running:
				testStatusMap[id] = Running
				lastUpdated[id] = time.Now()
			case Initialized:
				testStatusMap[id] = Initialized
				consecutiveErrors[id] = 0
			case Deleted:
				delete(testStatusMap, id)
				delete(lastUpdated, id)
				delete(consecutiveErrors, id)
			default:
				log.Error("invalid status code sent on statusChan")
			}
		case <-timeForTimeoutCheck:
			log.Info("looking for timed out tests")
			for id, statusCode := range testStatusMap {
				if statusCode == Running {
					muTests.RLock()
					if runningTest, ok := tests[id]; ok {
						testTimeout := runningTest.Get().Timeout*2*time.Second // TODO: *2
						if time.Now().After(lastUpdated[id].Add(testTimeout)) {
							testStatusMap[id] = TimedOut
							consecutiveErrors[id]++
							testError := errors.New("test timed out - found while manually checking for timeouts")
							addTestLog(id, TimedOut, testTimeout, testError, db)
							testContactTypes, err := dao.GetTestContactsType(id, db)
							if err != nil {
								log.Error(err)
							}
							// Check that test still exists
							if _, ok := tests[id]; !ok {
								muTests.RUnlock()
								continue
							}
							testContactsNotified[id] = notifyContactsError(id, TimedOut, testError,
														consecutiveErrors[id], testContactTypes,
														testContactsNotified[id], db)
							newTestChan <- tests[id] // Might not be the right way of doing it
							// could cause an infinite amount of goroutines to get stuck
						}
					}
					muTests.RUnlock()
				}
			}
		}
	}
}

func notifyContactsError(
		testId string,
		statusCode uint,
		testError error,
		consecutiveErrors uint,
		testContacts []dao.TestContactType,
		testContactsNotified[]string,
		db *sqlx.DB,
	)[]string {

	var mailsToContact []string
	var mailContactIds []string
	var postHooksToSend []string
	var hookContactIds []string

	for _, contactType := range testContacts {
		// Have we reached the threshold for contact?
		if consecutiveErrors >= contactType.Threshold {
			// Check if contact has already been contacted
			contacted := false
			for _, contactId := range testContactsNotified {
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
		err := notifications.SendEmail(mailsToContact, tests[testId].Get(), testError, db)
		if err != nil {
			log.Error("unable to send email: ", err)
		} else {
			testContactsNotified = append(testContactsNotified, mailContactIds...)
		}
	}
	if len(postHooksToSend) != 0 {
		err := notifications.PostHook(postHooksToSend, tests[testId].Get(), testError, statusCode)
		if err != nil {
			log.Error("unable to post hook")
		} else {
			testContactsNotified = append(testContactsNotified, hookContactIds...)
		}
	}
	return testContactsNotified
}

func dataBaseListener(db *sqlx.DB) {
	for {
		select {
			case <-time.After(DataBaseListenerInterval):
		}
		log.Info("Looking for new/updated tests in DB")
		testsDB, err := dao.GetTests(db)
		if err != nil {
			continue // Bad
		}

		var newTests [] pingr.Test
		var deletedTests []string

		// Look for new/updated test
		muTests.RLock()
		for _, j := range testsDB {
			if !reflect.DeepEqual(tests[j.Get().TestId], j) { // Found a new/updated test
				newTests = append(newTests, j)
			}
		}

		// Look for deleted tests
		for testId := range tests {
			// slow but probably okay since user wont update/add that often
			if !intInSlice(testId, testsDB) {
				deletedTests = append(deletedTests, testId)
			}
		}
		muTests.RUnlock()

		// Save result in array and send on channel after Unlock to avoid deadlock
		for _, newTest := range newTests {
			newTestChan <- newTest
		}
		for _, testId := range deletedTests {
			deletedTestChan <- testId
		}

	}
}

func initVars(db *sqlx.DB) {
	testsDB, err := dao.GetTests(db)
	if err != nil {
		log.Warn(err)
	}

	muTests.Lock()
	defer muTests.Unlock()

	for _, test := range testsDB {
		testId := test.Get().TestId
		tests[testId] = test

		addTestLog(testId, Initialized, 0, nil, db)
		statusChan <- testStatus{testId, Initialized, nil}

		closeChans[testId] = make(chan int)
		go worker(test, closeChans[testId], db)
	}
}

func addTestLog(testId string, statusCode uint, rt time.Duration, err error, db *sqlx.DB) {
	var logMessage string
	if err != nil {
		logMessage = err.Error()
	}
	log.Info(fmt.Sprintf("TestID: %s, StatusCode: %d", testId, statusCode))
	l := pingr.Log{
		TestId:     		testId,
		StatusId:  		statusCode,
		Message:   		logMessage,
		ResponseTime: 	rt,
		CreatedAt: 		time.Now(),
	}
	err = dao.PostLog(l, db)
	if err != nil {
		log.Warn(err)
	}
}

func discSpaceMaintainer(db *sqlx.DB) {
	for {
		log.Info("checking available disc space")
		var stat syscall.Statfs_t

		wd, err := os.Getwd()
		if err != nil {
			log.Error(err)
		}
		err = syscall.Statfs(wd, &stat)
		if err != nil {
			log.Error(err)
		}
		availableGB := stat.Bavail*uint64(stat.Bsize)/uint64(math.Pow(1024,3))
		if availableGB < config.Get().MinDiscStorage {
			err = dao.DeleteLastNLogs(100000, db)
			if err != nil {
				log.Error(err)
			}
		}
		time.Sleep(time.Hour)
	}

}

func NotifyNewTest(test pingr.Test) {
	newTestChan <- test
}

func NotifyDeletedTest(testId string){
	deletedTestChan <- testId
}

func NotifyPushTest(testId string, body []byte) error {
	pingr.MuPush.RLock()
	defer pingr.MuPush.RUnlock()
	if _, ok := pingr.PushChans[testId]; ok {
		select {
		case pingr.PushChans[testId] <- body:
			return nil
		default:
			return errors.New("test not responding")
		}
	}
	return errors.New("test not initialized yet")
}

func intInSlice(a string, list []pingr.Test) bool {
	for _, b := range list {
		if b.Get().TestId == a {
			return true
		}
	}
	return false
}