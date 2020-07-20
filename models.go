package pingr

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx/types"
	log "github.com/sirupsen/logrus"
	"pingr/internal/bus"
	"pingr/internal/poll"
	"pingr/internal/push"
	"time"
)

type Log struct {
	LogId        uint64        `json:"log_id" db:"log_id"`
	TestId       string        `json:"test_id" db:"test_id"`
	StatusId     uint          `json:"status_id" db:"status_id"`
	Message      string        `json:"message" db:"message"`
	ResponseTime time.Duration `json:"response_time" db:"response_time"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
}

type Contact struct {
	ContactId   string `json:"contact_id" db:"contact_id"`
	ContactName string `json:"contact_name" db:"contact_name"`
	ContactType string `json:"contact_type" db:"contact_type"`
	ContactUrl  string `json:"contact_url" db:"contact_url"`
}

func (c Contact) Validate() bool {
	if c.ContactId == "" {
		return false
	}
	if c.ContactName == "" {
		return false
	}
	switch c.ContactType {
	case "smtp", "http":
	default:
		return false
	}
	if c.ContactUrl == "" {
		return false
	}
	return true
}

type TestContact struct {
	ContactId string `json:"contact_id" db:"contact_id"`
	TestId    string `json:"test_id" db:"test_id"`
	Threshold uint   `json:"threshold" db:"threshold"`
}

func (c TestContact) Validate() bool {
	if c.ContactId == "" {
		return false
	}
	if c.TestId == "" {
		return false
	}
	if c.Threshold == 0 {
		return false
	}
	return true
}

type Test interface {
	RunTest(buz *bus.Bus) (time.Duration, error)
	Validate() bool
}

type BaseTest struct {
	TestId    string        `json:"test_id" db:"test_id"`
	TestName  string        `json:"test_name" db:"test_name"`
	TestType  string        `json:"test_type" db:"test_type"`
	Timeout   time.Duration `json:"timeout"`
	Url       string        `json:"url"`
	Interval  time.Duration `json:"interval"`
	CreatedAt time.Time     `json:"created_at" db:"created_at"`
}

func (j BaseTest) Get() BaseTest {
	return j
}

type GenericTest struct {
	BaseTest
	Blob types.JSONText `json:"blob" db:"blob"`

	memoize Test
}

func (j GenericTest) RunTest(buz *bus.Bus) (time.Duration, error) {
	var err error
	if j.memoize == nil {
		j.memoize, err = j.Impl()
		if err != nil {
			return 0, err
		}
	}
	return j.memoize.RunTest(buz)
}

func (j GenericTest) Validate() bool {
	var err error
	if j.memoize == nil {
		j.memoize, err = j.Impl()
		if err != nil {
			return false
		}
	}
	return j.memoize.Validate()
}

func (j GenericTest) Impl() (parsedTest Test, err error) {
	switch j.TestType {
	case "SSH":
		var t SSHTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t.Blob)
		if err != nil {
			return
		}
		parsedTest = t
	case "TCP":
		var t TCPTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t.Blob)
		if err != nil {
			return
		}
		parsedTest = t
	case "TLS":
		var t TLSTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t.Blob)
		if err != nil {
			return
		}
		parsedTest = t
	case "Ping":
		var t PingTest
		t.BaseTest = j.BaseTest
		parsedTest = t
	case "HTTP":
		var t HTTPTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t.Blob)
		if err != nil {
			return
		}
		parsedTest = t
	case "DNS":
		var t DNSTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t.Blob)
		if err != nil {
			return
		}
		parsedTest = t
	case "Prometheus":
		var t PrometheusTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t.Blob)
		if err != nil {
			return
		}
		parsedTest = t
	case "PrometheusPush":
		var t PrometheusPushTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t.Blob)
		if err != nil {
			return
		}
		parsedTest = t
	case "HTTPPush":
		var t HTTPPushTest
		t.BaseTest = j.BaseTest
		parsedTest = t
	default:
		err = errors.New(j.TestType + " is not a valid test type")
	}
	return
}

func (j BaseTest) Validate() bool {
	if j.TestId == "" {
		return false
	}
	if j.TestName == "" {
		return false
	}
	switch j.TestType {
	case "HTTP", "Prometheus", "TLS", "DNS", "Ping", "SSH", "TCP":
		if j.Url == "" {
			return false
		}
		if j.Interval < 0 {
			return false
		}
	case "HTTPPush", "PrometheusPush":
		if j.Interval != 0 {
			return false
		}
	default:
		return false
	}
	if j.Timeout == 0 {
		return false
	}
	return true
}

type SSHTest struct {
	Blob struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		Port       string `json:"port"`
		UseKeyPair bool   `json:"use_key_pair"`
	} `json:"blob"`
	BaseTest
}

func (t SSHTest) RunTest(*bus.Bus) (time.Duration, error) {
	return poll.SSH(t.Url, t.Blob.Port, t.Timeout, t.Blob.Username, t.Blob.Password, t.Blob.UseKeyPair)
}

func (t SSHTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	if t.Blob.Username == "" {
		return false
	}
	if t.Blob.Port == "" {
		return false
	}
	if !t.Blob.UseKeyPair && t.Blob.Password == "" {
		return false //Maybe some ssh servers won't require password but I guess most do
	}
	return true
}

type TCPTest struct {
	Blob struct {
		Port string `json:"port"`
	} `json:"blob"`
	BaseTest
}

func (t TCPTest) RunTest(*bus.Bus) (time.Duration, error) {
	log.Info(t)
	return poll.TCP(t.Url, t.Blob.Port, t.Timeout)
}

func (t TCPTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	if t.Blob.Port == "" {
		return false
	}
	return true
}

type TLSTest struct {
	Blob struct {
		Port string `json:"port"`
	} `json:"blob"`
	BaseTest
}

func (t TLSTest) RunTest(*bus.Bus) (time.Duration, error) {
	return poll.TLS(t.Url, t.Blob.Port, t.Timeout)
}

func (t TLSTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	if t.Blob.Port == "" {
		return false
	}
	return true
}

type PingTest struct {
	Blob struct{} `json:"blob"`
	BaseTest
}

func (t PingTest) RunTest(*bus.Bus) (time.Duration, error) {
	return poll.Ping(t.Url, t.Timeout)
}

func (t PingTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	return true
}

type HTTPTest struct {
	Blob struct {
		ReqMethod  string            `json:"req_method""`
		ReqHeaders map[string]string `json:"req_headers"`
		ReqBody    string            `json:"req_body"`

		ResStatus  int               `json:"res_status"`
		ResHeaders map[string]string `json:"res_headers"`
		ResBody    string            `json:"res_body"`
	} `json:"blob"`
	BaseTest
}

func (t HTTPTest) RunTest(*bus.Bus) (time.Duration, error) {
	return poll.HTTP(t.Url, t.Blob.ReqMethod, t.Timeout*time.Second, []byte(t.Blob.ReqBody), []byte(t.Blob.ResBody))
}

func (t HTTPTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	switch t.Blob.ReqMethod {
	case "GET", "POST", "PUT", "HEAD", "DELETE":
	default:
		return false
	}
	return true
}

type DNSTest struct {
	Blob struct {
		IpAddr string   `json:"ip_addr"`
		CNAME  string   `json:"cname"`
		TXT    []string `json:"txt"`
	} `json:"blob"`

	BaseTest
}

func (t DNSTest) RunTest(*bus.Bus) (time.Duration, error) {
	return 0, nil //TODO fix poll.DNS(t.Url, t.Timeout*time.Second, t.Blob.IpAddr, t.Blob.CNAME, t.Blob.TXT)
}

func (t DNSTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	if len(t.Blob.TXT) == 0 && t.Blob.CNAME == "" && t.Blob.IpAddr == "" { // Has to test something
		return false
	}
	return true
}

type PrometheusTest struct {
	Blob struct {
		MetricTests []push.MetricTest `json:"metric_tests"`
	} `json:"blob"`
	BaseTest
}

func (t PrometheusTest) RunTest(*bus.Bus) (time.Duration, error) {
	return poll.Prometheus(t.TestId, t.Url, t.Timeout*time.Second, t.Blob.MetricTests)
}

func (t PrometheusTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	if len(t.Blob.MetricTests) == 0 {
		return false
	}
	for _, metricTest := range t.Blob.MetricTests {
		if !metricTest.Validate() {
			return false
		}
	}
	return true
}

type HTTPPushTest struct {
	Blob struct{} `json:"blob"`
	BaseTest
}

func (t HTTPPushTest) RunTest(buz *bus.Bus) (time.Duration, error) {
	start := time.Now()
	_, err := buz.Next(fmt.Sprintf("push:%s", t.TestId), t.Timeout*time.Second)

	return time.Since(start), err
}

func (t HTTPPushTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	return true
}

type PrometheusPushTest struct {
	Blob struct {
		MetricTests []push.MetricTest `json:"metric_tests"`
	} `json:"blob"`
	BaseTest
}

func (t PrometheusPushTest) RunTest(buz *bus.Bus) (time.Duration, error) {
	start := time.Now()
	reqBody, err := buz.Next(fmt.Sprintf("push:%s", t.TestId), t.Timeout*time.Second)
	if err != nil {
		return time.Since(start), err
	}
	err = push.Prometheus(t.TestId, reqBody, t.Blob.MetricTests)

	return time.Since(start), err
}

func (t PrometheusPushTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	if len(t.Blob.MetricTests) == 0 {
		return false
	}
	for _, metricTest := range t.Blob.MetricTests {
		if !metricTest.Validate() {
			return false
		}
	}
	return true
}
