package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"os"
	"pingr"
	"pingr/internal/bus"
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
	Successful  uint = 1
	Error       uint = 2
	TimedOut    uint = 3
	Running     uint = 4
	Initialized uint = 5
	Deleted     uint = 6
)

type testStatus struct {
	TestId string
	Status uint
	Error  error
}

type Scheduler struct {
	buz     *bus.Bus
	db      *sqlx.DB
	muTests sync.RWMutex

	tests      map[string]pingr.GenericTest
	closeChans map[string]chan struct{}

	statusChan          chan testStatus
	timeForTimeoutCheck chan int
}

func New(db *sqlx.DB, buz *bus.Bus) *Scheduler {
	s := &Scheduler{
		db:                  db,
		buz:                 buz,
		tests:               make(map[string]pingr.GenericTest),
		closeChans:          make(map[string]chan struct{}),
		statusChan:          make(chan testStatus, 10),
		timeForTimeoutCheck: make(chan int),
	}

	go s.discSpaceMaintainer(db)

	go s.dataBaseListener(db) // Best way to handle DB error??

	go s.testMaintainer(db)

	s.initVars(db)

	go s.commands()

	return s
}

func (s *Scheduler) commands() {
	go func (){
		for {
			data, err := s.buz.Next("delete", time.Minute)
			if  err != nil{
				// Probably a timeout
				// could be channel closed, but it should be fixed next iteration
				s.timeForTimeoutCheck <- 1
				continue
			}
			testId := string(data)
			if c, ok := s.closeChans[testId]; ok {
				close(c)
				delete(s.closeChans, testId)
			}
			addTestLog(testId, Deleted, 0, nil, s.db)
			s.statusChan <- testStatus{testId, Deleted, nil}

			err = s.buz.Close(fmt.Sprintf("push:%s", testId))
			if err != nil{
				log.Error("could not close bus channel: ", err.Error())
			}

			s.muTests.Lock()
			if _, ok := s.tests[testId]; ok {
				delete(s.tests, testId)
				log.Info(fmt.Sprintf("TestID: %s, Test deleted", testId))
			}
			s.muTests.Unlock()
		}

	}()
	go func (){
		for {
			data, err := s.buz.Next("new", time.Minute)
			if  err != nil{
				// Probably a timeout
				// could be channel closed, but it should be fixed next iteration
				continue
			}
			var test pingr.GenericTest
			err = json.Unmarshal(data, &test)
			if err != nil {
				log.Error("could not unmarshal test: ", err)
				continue
			}

			s.muTests.Lock()
			testId := test.Get().TestId
			s.tests[testId] = test
			s.muTests.Unlock()

			if _, ok := s.closeChans[testId]; ok {
				close(s.closeChans[testId])
				delete(s.closeChans, testId)
			}

			addTestLog(testId, Initialized, 0, nil, s.db)
			s.statusChan <- testStatus{testId, Initialized, nil}

			s.closeChans[testId] = make(chan struct{})
			go s.worker(test, s.closeChans[testId])
		}
	}()

}

func (s *Scheduler) worker(test pingr.GenericTest, close chan struct{}) {
	initSleep := rand.Int63n(300)
	if config.Get().Dev {
		initSleep = 0
	}
	time.Sleep(time.Duration(initSleep)*time.Second)

	testId := test.Get().TestId
	for {
		s.statusChan <- testStatus{testId, Running, nil}

		rt, err := test.RunTest(s.buz)

		statusCode := Successful
		if err != nil {
			statusCode = Error
		}
		addTestLog(testId, statusCode, rt, err, s.db)
		s.statusChan <- testStatus{testId, statusCode, err}

		select {
		case <-time.After(test.Get().Interval * time.Second):
		case <-close:
			return
		}
	}
}

func (s *Scheduler) testMaintainer(db *sqlx.DB) {
	testStatusMap := make(map[string]uint)
	lastUpdated := make(map[string]time.Time)
	consecutiveErrors := make(map[string]uint)
	testContactsNotified := make(map[string][]string)
	for {
		select {
		case status := <- s.statusChan:
			id := status.TestId
			switch status.Status {
			case Successful:
				testStatusMap[id] = Successful
				consecutiveErrors[id] = 0
				if _, ok := testContactsNotified[id]; ok {
					// This is an active incident, notify that test is successful again
					s.muTests.RLock()
					// Check that test still exists
					if _, ok := s.tests[id]; !ok {
						s.muTests.RUnlock()
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
						emailErr = notifications.SendEmail(mailsToContact, s.tests[id].Get(), status.Error, db)
					}
					if len(postHooksToSend) != 0 {
						postHookErr = notifications.PostHook(postHooksToSend, s.tests[id].Get(), status.Error, Successful)
					}
					if emailErr == nil && postHookErr == nil {
						// Keep sending until no errors occur
						delete(testContactsNotified, id)
					}
					s.muTests.RUnlock()
				}
			case Error:
				testStatusMap[id] = Error
				consecutiveErrors[id]++
				testContactTypes, err := dao.GetTestContactsType(id, db)
				if err != nil {
					log.Error(err)
				}
				s.muTests.RLock()
				// Check that test still exists
				if _, ok := s.tests[id]; !ok {
					s.muTests.RUnlock()
					break
				}
				testContactsNotified[id] = s.notifyContactsError(id, Error, status.Error,
					consecutiveErrors[id], testContactTypes,
					testContactsNotified[id], db)
				s.muTests.RUnlock()
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
		case <-s.timeForTimeoutCheck:
			log.Info("looking for timed out tests")
			for id, statusCode := range testStatusMap {
				if statusCode == Running {
					s.muTests.RLock()
					if runningTest, ok := s.tests[id]; ok {
						testTimeout := runningTest.Get().Timeout * 2 * time.Second // TODO: *2
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
							if _, ok := s.tests[id]; !ok {
								s.muTests.RUnlock()
								continue
							}
							testContactsNotified[id] = s.notifyContactsError(id, TimedOut, testError,
								consecutiveErrors[id], testContactTypes,
								testContactsNotified[id], db)
							t, err := dao.GetRawTest(id, s.db)
							if err != nil {
								log.Error(err)
								continue
							}
							data, err := json.Marshal(t)
							if err != nil {
								log.Error(err)
								continue
							}
							err = s.buz.Publish("new", data)
							if err != nil {
								log.Error(err)
								continue
							}

						}
					}
					s.muTests.RUnlock()
				}
			}
		}
	}
}

func (s *Scheduler) notifyContactsError(
	testId string,
	statusCode uint,
	testError error,
	consecutiveErrors uint,
	testContacts []dao.TestContactType,
	testContactsNotified []string,
	db *sqlx.DB,
) []string {

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
		err := notifications.SendEmail(mailsToContact, s.tests[testId].Get(), testError, db)
		if err != nil {
			log.Error("unable to send email: ", err)
		} else {
			testContactsNotified = append(testContactsNotified, mailContactIds...)
		}
	}
	if len(postHooksToSend) != 0 {
		err := notifications.PostHook(postHooksToSend, s.tests[testId].Get(), testError, statusCode)
		if err != nil {
			log.Error("unable to post hook")
		} else {
			testContactsNotified = append(testContactsNotified, hookContactIds...)
		}
	}
	return testContactsNotified
}

func (s *Scheduler) dataBaseListener(db *sqlx.DB) {
	for {
		select {
		case <-time.After(DataBaseListenerInterval):
		}
		log.Info("Looking for new/updated tests in DB")
		testsDB, err := dao.GetRawTests(db)
		if err != nil {
			continue // Bad
		}

		var newTests []pingr.GenericTest
		var deletedTests []string

		// Look for new/updated test
		s.muTests.RLock()
		for _, j := range testsDB {
			if !reflect.DeepEqual(s.tests[j.Get().TestId], j) { // Found a new/updated test
				newTests = append(newTests, j)
			}
		}

		// Look for deleted tests
		for testId := range s.tests {
			// slow but probably okay since user wont update/add that often
			if !intInSlice(testId, testsDB) {
				deletedTests = append(deletedTests, testId)
			}
		}
		s.muTests.RUnlock()

		// Save result in array and send on channel after Unlock to avoid deadlock
		for _, newTest := range newTests {
			data, err := json.Marshal(newTest)
			if err != nil {
				log.Error("could not marchal test: ", err.Error())
				continue
			}
			err = s.buz.Publish("new", data)
			if err != nil {
				log.Error("could not add test: ", err.Error())
			}
		}
		for _, testId := range deletedTests {
			err = s.buz.Publish("delete", []byte(testId))
			if err != nil {
				log.Error("could not delete test: ", err.Error())
			}
		}

	}
}

func (s *Scheduler) initVars(db *sqlx.DB) {
	testsDB, err := dao.GetRawTests(db)
	if err != nil {
		log.Warn(err)
	}

	s.muTests.Lock()
	defer s.muTests.Unlock()

	for _, test := range testsDB {
		testId := test.Get().TestId
		s.tests[testId] = test

		addTestLog(testId, Initialized, 0, nil, db)
		s.statusChan <- testStatus{testId, Initialized, nil}

		s.closeChans[testId] = make(chan struct{})
		go s.worker(test, s.closeChans[testId])
	}
}

func addTestLog(testId string, statusCode uint, rt time.Duration, err error, db *sqlx.DB) {
	var logMessage string
	if err != nil {
		logMessage = err.Error()
	}
	log.Info(fmt.Sprintf("TestID: %s, StatusCode: %d", testId, statusCode))
	l := pingr.Log{
		TestId:       testId,
		StatusId:     statusCode,
		Message:      logMessage,
		ResponseTime: rt,
		CreatedAt:    time.Now(),
	}
	err = dao.PostLog(l, db)
	if err != nil {
		log.Warn(err)
	}
}

func (s *Scheduler) discSpaceMaintainer(db *sqlx.DB) {
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
		availableGB := stat.Bavail * uint64(stat.Bsize) / uint64(math.Pow(1024, 3))
		if availableGB < config.Get().MinDiscStorage {
			err = dao.DeleteLastNLogs(100000, db)
			if err != nil {
				log.Error(err)
			}
		}
		time.Sleep(time.Hour)
	}

}

func intInSlice(a string, list []pingr.GenericTest) bool {
	for _, b := range list {
		if b.Get().TestId == a {
			return true
		}
	}
	return false
}
