package scheduler

import (
	"database/sql"
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
	DataBaseListenerInterval = 10 * time.Minute

	// Change in DB as well
	Successful  uint = 1
	Error       uint = 2
	TimedOut    uint = 3
	Initialized uint = 5
	Paused      uint = 6
)

type Scheduler struct {
	buz     *bus.Bus
	db      *sqlx.DB
	muTests sync.RWMutex

	tests      map[string]pingr.GenericTest
	closeChans map[string]chan struct{}
}

func New(db *sqlx.DB, buz *bus.Bus) *Scheduler {
	s := &Scheduler{
		db:         db,
		buz:        buz,
		tests:      make(map[string]pingr.GenericTest),
		closeChans: make(map[string]chan struct{}),
	}

	go s.discSpaceMaintainer()

	go s.dataBaseListener() // Best way to handle DB error??

	go s.handleTimeouts()

	s.initVars()

	go s.commands()

	return s
}

func (s *Scheduler) closeTest(testId string) error {
	if c, ok := s.closeChans[testId]; ok {
		close(c)
		delete(s.closeChans, testId)
	}

	err := s.buz.Close(fmt.Sprintf("push:%s", testId))
	if err != nil {
		log.Error("could not close bus channel: " + err.Error())
	}

	s.muTests.Lock()
	defer s.muTests.Unlock()
	if _, ok := s.tests[testId]; ok {
		delete(s.tests, testId)
		log.Info(fmt.Sprintf("TestID: %s, Test paused/deleted", testId))
	}
	return nil

}

func (s *Scheduler) commands() {
	go func() {
		for {
			data, err := s.buz.Next("deactivate", time.Minute)
			if err != nil {
				// Probably a timeout
				// could be channel closed, but it should be fixed next iteration
				continue
			}
			testId := string(data)
			err = s.closeTest(testId)
			if err != nil {
				log.Error(err)
			}
			addTestLog(testId, Paused, 0, nil, s.db)
		}
	}()

	go func() {
		for {
			data, err := s.buz.Next("delete", time.Minute)
			if err != nil {
				// Probably a timeout
				// could be channel closed, but it should be fixed next iteration
				continue
			}
			err = s.closeTest(string(data))
			if err != nil {
				log.Error(err)
			}
		}
	}()

	go func() {
		for {
			data, err := s.buz.Next("new", time.Minute)
			if err != nil {
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

			s.closeChans[testId] = make(chan struct{})
			go s.worker(test, s.closeChans[testId])
		}
	}()

}

func (s *Scheduler) worker(test pingr.GenericTest, close chan struct{}) {
	// Spread out execution of tests
	initSleep := rand.Int63n(int64((test.Timeout + test.Interval)/2))
	if config.Get().Dev {
		initSleep = 0
	}
	time.Sleep(time.Duration(initSleep) * time.Second)
	for {
		rt, err := test.RunTest(s.buz)

		// Avoid adding logs after timeouts etc
		select {
		case <-close:
			return
		default:
			s.reportTestResponse(test.BaseTest, err, rt)
		}

		select {
		case <-time.After(test.Get().Interval * time.Second):
		case <-close:
			return
		}
	}
}

func (s *Scheduler) reportTestResponse(test pingr.BaseTest, testErr error, rt time.Duration) {
	if testErr != nil {
		addTestLog(test.TestId, Error, rt, testErr, s.db)
		s.handleError(test, testErr)
		return
	}

	addTestLog(test.TestId, Successful, rt, testErr, s.db)
	s.handleSuccess(test)
}

func (s *Scheduler) handleSuccess(test pingr.BaseTest) {
	testId := test.TestId
	incident, err := dao.GetActiveIncident(testId, s.db)
	if err == sql.ErrNoRows {
		// No active incident, nothing needs to be done
		return
	}
	if err != nil {
		log.Error(fmt.Sprintf("could not get active incident: %v", err))
		return
	}

	contacts, err := dao.GetIncidentContacts(incident.IncidentId, s.db)
	if err != nil {
		log.Error(fmt.Sprintf("could not get incident contacts: %v", err))
		return
	}
	for _, contact := range contacts {
		switch contact.ContactType {
		case "smtp":
			err = notifications.SendEmail([]string{contact.ContactUrl}, test, nil, s.db)
		case "http":
			err = notifications.PostHook([]string{contact.ContactUrl}, test, nil, Successful)
		}
		if err != nil {
			log.Error(err)
			return
		}
	}

	// All contacts notified of success -> close incident
	err = dao.CloseIncident(incident.IncidentId, s.db)
	if err != nil {
		log.Error(fmt.Sprintf("could not close incident: %v", err))
	}
}

func (s *Scheduler) handleError(test pingr.BaseTest, testErr error) {
	testId := test.TestId
	incident, err := dao.GetActiveIncident(testId, s.db)

	if err != nil && err != sql.ErrNoRows {
		log.Error(fmt.Sprintf("could not get active incident: %v", err))
		return
	}

	var incidentId uint64
	if err == sql.ErrNoRows {
		i := pingr.Incident{
			TestId:    testId,
			Active:    true,
			RootCause: testErr.Error(),
			CreatedAt: time.Now(),
		}
		incidentId, err = dao.PostIncident(i, s.db)
		if err != nil {
			log.Error(fmt.Sprintf("could not post incident: %v", err))
			return
		}
	} else {
		incidentId = incident.IncidentId
	}

	contacts, err := dao.GetTestContactsToContact(testId, s.db)
	if err != nil {
		log.Error(fmt.Sprintf("could not get contacts to contact: %v", err))
		return
	}

	for _, contact := range contacts {
		switch contact.ContactType {
		case "smtp":
			err = notifications.SendEmail([]string{contact.ContactUrl}, test, testErr, s.db)
		case "http":
			err = notifications.PostHook([]string{contact.ContactUrl}, test, testErr, Error)
		}
		if err != nil {
			log.Error(fmt.Sprintf("could not send notification: %v", err))
			continue
		}
		err = dao.PostContactLog(pingr.IncidentContactLog{
			IncidentId: incidentId,
			ContactId:  contact.ContactId,
			Message:    testErr.Error(),
			CreatedAt:  time.Now(),
		}, s.db)
		if err != nil {
			log.Error(fmt.Sprintf("could not post contact log: %v", err))
		}
	}
}

func (s *Scheduler) handleTimeouts() {
	for {
		select {
		case <-time.After(2 * time.Minute):
			s.muTests.RLock()
			for testId, test := range s.tests {
				logs, err := dao.GetTestLogsLimited(testId, 1, s.db)
				if err != nil || len(logs) == 0 {
					log.Errorf("could not get test logs %v", err)
					continue
				}
				// Latest log + interval + timeout < now -> TIMEOUT!
				if time.Now().After(logs[0].CreatedAt.Add((test.Interval + 2*test.Timeout) * time.Second)) {
					err := errors.New("test considered timed out while manually scanning for timeouts")
					addTestLog(testId, TimedOut, (test.Interval+2*test.Timeout)*time.Second, err, s.db)
					s.handleError(test.BaseTest, err)
					testData, err := json.Marshal(test)
					if err != nil {
						log.Error("error marshaling test in timeout: " + err.Error())
						continue
					}
					s.buz.Publish("new", testData)
				}
			}
			s.muTests.RUnlock()

		}

	}
}

func (s *Scheduler) dataBaseListener() {
	for {
		select {
		case <-time.After(DataBaseListenerInterval):
		}
		log.Info("Looking for new/updated tests in DB")
		testsDB, err := dao.GetRawTests(s.db)
		if err != nil {
			continue // Bad
		}

		var newTests []pingr.GenericTest
		var deletedTests []string

		// Look for new/updated test
		s.muTests.RLock()
		for _, j := range testsDB {
			if !reflect.DeepEqual(s.tests[j.Get().TestId], j) && j.Active { // Found a new/updated test
				newTests = append(newTests, j)
			}
		}

		// Look for deleted tests
		for testId := range s.tests {
			if !intInSlice(testId, testsDB) {
				deletedTests = append(deletedTests, testId)
			}
		}
		s.muTests.RUnlock()

		// Save result in array and send on channel after Unlock to avoid deadlock
		for _, newTest := range newTests {
			data, err := json.Marshal(newTest)
			if err != nil {
				log.Error("could not marshal test: ", err.Error())
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

func (s *Scheduler) initVars() {
	testsDB, err := dao.GetRawTests(s.db)
	if err != nil {
		log.Warn(err)
	}

	s.muTests.Lock()
	defer s.muTests.Unlock()

	for _, test := range testsDB {
		if !test.Active {
			continue
		}
		testId := test.Get().TestId
		s.tests[testId] = test

		addTestLog(testId, Initialized, 0, nil, s.db)

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

func (s *Scheduler) discSpaceMaintainer() {
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
			err = dao.DeleteLastNLogs(100000, s.db)
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
