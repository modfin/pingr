package pingr

import (
	"errors"
	"pingr/internal/poll"
	"pingr/internal/push"
	"sync"
	"time"
)

type Log struct {
	LogId     		uint64			`db:"log_id"`
	TestId     		string			`db:"test_id"`
	StatusId  		uint			`db:"status_id"`
	Message   		string
	ResponseTime 	time.Duration	`db:"response_time"`
	CreatedAt 		time.Time		`db:"created_at"`
}

type Contact struct {
	ContactId 	string `json:"contact_id" db:"contact_id"`
	ContactName string `json:"contact_name" db:"contact_name"`
	ContactType string `json:"contact_type" db:"contact_type"`
	ContactUrl 	string `json:"contact_url" db:"contact_url"`
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
	ContactId 	string 	`json:"contact_id" db:"contact_id"`
	TestId 		string 	`json:"test_id" db:"test_id"`
	Threshold	uint 	`json:"threshold" db:"threshold"`
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
	RunTest() 	(time.Duration, error)
	Validate()	bool
	Get()		BaseTest
}

type BaseTest struct {
	TestId		string			`json:"test_id" db:"test_id"`
	TestName	string 			`json:"test_name" db:"test_name"`
	Timeout 	time.Duration 	`json:"timeout"`
	Url 		string 			`json:"url"`
	Interval 	time.Duration 	`json:"interval"`
	CreatedAt	time.Time		`json:"created_at" db:"created_at"`
	TestType 	string			`json:"test_type" db:"test_type"`
}

func (j BaseTest) Validate() bool {
	if j.TestId == "" {return false}
	if j.TestName == "" {
		return false
	}
	switch j.TestType {
	case "HTTP", "Prometheus", "TLS", "DNS", "Ping", "SSH", "TCP":
		if j.Url == "" {
			return false
		}
	case "HTTPPush", "PrometheusPush":
	default:
		return false
	}
	if j.Interval < 0 {
		return false
	}
	if j.Timeout == 0 {
		return false
	}
	return true
}

type SSHTest struct {
	Username 	string 	`json:"username"`
	Password	string 	`json:"password"`
	Port		string 	`json:"port"`
	UseKeyPair	bool 	`json:"use_key_pair"`
	BaseTest
}

func (t SSHTest) RunTest() (time.Duration, error) {
	return poll.SSH(t.Url, t.Port, t.Timeout, t.Username, t.Password, t.UseKeyPair)
}

func (t SSHTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	if t.Username == "" {
		return false
	}
	if t.Port == "" {
		return false
	}
	if !t.UseKeyPair && t.Password == "" {
		return false//Maybe some ssh servers won't require password but I guess most do
	}
	return true
}

func (t SSHTest) Get() BaseTest {
	return t.BaseTest
}

type TCPTest struct {
	Port	string `json:"port"`
	BaseTest
}

func (t TCPTest) RunTest() (time.Duration, error) {
	return poll.TCP(t.Url, t.Port, t.Timeout)
}

func (t TCPTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	if t.Port == "" {
		return false
	}
	return true
}

func (t TCPTest) Get() BaseTest {
	return t.BaseTest
}

type TLSTest struct {
	Port 	string `json:"port"`
	BaseTest
}

func (t TLSTest) RunTest() (time.Duration, error) {
	return poll.TLS(t.Url, t.Port, t.Timeout)
}

func (t TLSTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	if t.Port == "" {
		return false
	}
	return true
}

func (t TLSTest) Get() BaseTest {
	return t.BaseTest
}

type PingTest struct {
	BaseTest
}

func (t PingTest) RunTest() (time.Duration, error) {
	return poll.Ping(t.Url, t.Timeout)
}

func (t PingTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	return true
}

func (t PingTest) Get() BaseTest {
	return t.BaseTest
}

type HTTPTest struct {
	Method 		string `json:"method"`
	Payload		[]byte `json:"payload"`
	ExpResult	[]byte `json:"exp_result"`
	BaseTest
}

func (t HTTPTest) RunTest() (time.Duration, error) {
	return poll.HTTP(t.Url, t.Method, t.Timeout*time.Second, t.Payload, t.ExpResult)
}

func (t HTTPTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	switch t.Method {
	case "GET", "POST", "PUT", "HEAD", "DELETE":
	default:
		return false
	}
	return true
}

func (t HTTPTest) Get() BaseTest {
	return t.BaseTest
}

type DNSTest struct {
	 IpAddr string 		`json:"ip_addr"`
	 CNAME	string 		`json:"cname"`
	 TXT	[]string 	`json:"txt"`
	 BaseTest
}

func (t DNSTest) RunTest() (time.Duration, error) {
	return poll.DNS(t.Url, t.Timeout*time.Second, t.IpAddr, t.CNAME, t.TXT)
}

func (t DNSTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	if len(t.TXT) == 0 && t.CNAME == "" && t.IpAddr == "" { // Has to test something
		return false
	}
	return true
}

func (t DNSTest) Get() BaseTest {
	return t.BaseTest
}

var (
	MuPush				sync.RWMutex
	PushChans 			= make(map[string] chan []byte)
)

type PrometheusTest struct {
	MetricTests []push.MetricTest `json:"metric_tests"`
	BaseTest
}

func (t PrometheusTest) RunTest() (time.Duration, error) {
	return poll.Prometheus(t.TestId, t.Url, t.Timeout * time.Second, t.MetricTests)
}

func (t PrometheusTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	if len(t.MetricTests) == 0 {
		return false
	}
	for _, metricTest := range t.MetricTests {
		if !metricTest.Validate() {
			return false
		}
	}
	return true
}

func (t PrometheusTest) Get() BaseTest {
	return t.BaseTest
}


type HTTPPushTest struct {
	BaseTest
}

func (t HTTPPushTest) RunTest() (time.Duration, error) {
	MuPush.RLock()
	pushChannel, ok := PushChans[t.TestId]
	MuPush.RUnlock()

	if !ok {
		MuPush.Lock()
		PushChans[t.TestId] = make(chan []byte)
		MuPush.Unlock()
		pushChannel = PushChans[t.TestId]
	}

	start := time.Now()
	select {
	case <-pushChannel:
		return time.Since(start), nil
	case <-time.After(t.Timeout*time.Second):
		return time.Since(start), errors.New("timeout reached on push test")
	}
}

func (t HTTPPushTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	return true
}

func (t HTTPPushTest) Get() BaseTest {
	return t.BaseTest
}

type PrometheusPushTest struct {
	MetricTests []push.MetricTest 	`json:"metric_tests"`
	BaseTest
}

func (t PrometheusPushTest) RunTest() (time.Duration, error) {
	MuPush.RLock()
	pushChannel, ok := PushChans[t.TestId]
	MuPush.RUnlock()

	if !ok {
		MuPush.Lock()
		PushChans[t.TestId] = make(chan []byte)
		MuPush.Unlock()
		pushChannel = PushChans[t.TestId]
	}

	start := time.Now()
	select {
	case reqBody, ok := <-pushChannel:
		if ok {
			err := push.Prometheus(t.TestId, reqBody, t.MetricTests)
			return time.Since(start), err
		}
		return 0, nil
	case <-time.After(t.Timeout*time.Second):
		return time.Since(start), errors.New("timeout reached on push test")
	}
}

func (t PrometheusPushTest) Validate() bool {
	if !t.BaseTest.Validate() {
		return false
	}
	if len(t.MetricTests) == 0 {
		return false
	}
	for _, metricTest := range t.MetricTests {
		if !metricTest.Validate() {
			return false
		}
	}
	return true
}

func (t PrometheusPushTest) Get() BaseTest {
	return t.BaseTest
}