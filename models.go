package pingr

import (
	"pingr/internal/poll"
	"time"
)

type Log struct {
	LogId     		uint64			`db:"log_id"`
	JobId     		uint64			`db:"job_id"`
	StatusId  		uint			`db:"status_id"`
	Message   		string
	ResponseTime 	time.Duration	`db:"response_time"`
	CreatedAt 		time.Time		`db:"created_at"`
}

func (j BaseJob) Validate(idReq bool) bool {
	if idReq && j.JobId == 0 {return false}
	switch j.TestType {
	case "HTTP", "Prometheus", "TLS", "DNS", "Ping", "SSH", "TCP":
	default:
		return false
	}
	if j.Url == "" {return false}
	if j.Interval == 0 {return false}
	if j.Timeout == 0 {
		return false
	}
	return true
}

type Job interface {
	RunTest() 		(time.Duration, error)
	Validate(bool) 	bool
	Get()			BaseJob
}

type BaseJob struct {
	JobId		uint64			`json:"job_id" db:"job_id"`
	Url 		string 			`json:"url"`
	Interval 	time.Duration 	`json:"interval"`
	Timeout 	time.Duration 	`json:"timeout"`
	CreatedAt	time.Time		`json:"created_at" db:"created_at"`
	TestType 	string			`json:"test_type" db:"test_type"`
}

type SSHTest struct {
	Username 	string 	`json:"username"`
	Password	string 	`json:"password"`
	Port		string 	`json:"port"`
	UseKeyPair	bool 	`json:"use_key_pair"`
	BaseJob
}

func (t SSHTest) RunTest() (time.Duration, error) {
	return poll.SSH(t.Url, t.Port, t.Username, t.Password, t.UseKeyPair)
}

func (t SSHTest) Validate(idReq bool) bool {
	if !t.BaseJob.Validate(idReq) {
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

func (t SSHTest) Get() BaseJob {
	return t.BaseJob
}

type TCPTest struct {
	Port	string `json:"port"`
	BaseJob
}

func (t TCPTest) RunTest() (time.Duration, error) {
	return poll.TCP(t.Url, t.Port)
}

func (t TCPTest) Validate(idReq bool) bool {
	if !t.BaseJob.Validate(idReq) {
		return false
	}
	if t.Port == "" {
		return false
	}
	return true
}

func (t TCPTest) Get() BaseJob {
	return t.BaseJob
}

type TLSTest struct {
	Port string `json:"port"`
	BaseJob
}

func (t TLSTest) RunTest() (time.Duration, error) {
	return poll.TLS(t.Url, t.Port)
}

func (t TLSTest) Validate(idReq bool) bool {
	if !t.BaseJob.Validate(idReq) {
		return false
	}
	if t.Port == "" {
		return false
	}
	return true
}

func (t TLSTest) Get() BaseJob {
	return t.BaseJob
}

type PingTest struct {
	BaseJob
}

func (t PingTest) RunTest() (time.Duration, error) {
	return poll.Ping(t.Url)
}

func (t PingTest) Validate(idReq bool) bool {
	if !t.BaseJob.Validate(idReq) {
		return false
	}
	return true
}

func (t PingTest) Get() BaseJob {
	return t.BaseJob
}

type HTTPTest struct {
	Method 		string `json:"method"`
	Payload		[]byte `json:"payload"`
	ExpResult	[]byte `json:"exp_result"`
	BaseJob
}

func (t HTTPTest) RunTest() (time.Duration, error) {
	return poll.HTTP(t.Url, t.Method, t.Timeout*time.Second, t.Payload, t.ExpResult)
}

func (t HTTPTest) Validate(idReq bool) bool {
	if !t.BaseJob.Validate(idReq) {
		return false
	}
	switch t.Method {
	case "GET", "POST", "PUT", "HEAD", "DELETE":
	default:
		return false
	}
	return true
}

func (t HTTPTest) Get() BaseJob {
	return t.BaseJob
}

type DNSTest struct {
	 IpAddr string 		`json:"ip_addr"`
	 CNAME	string 		`json:"cname"`
	 TXT	[]string 	`json:"txt"`
	 BaseJob
}

func (t DNSTest) RunTest() (time.Duration, error) {
	return poll.DNS(t.Url, t.Timeout*time.Second, t.IpAddr, t.CNAME, t.TXT)
}

func (t DNSTest) Validate(idReq bool) bool {
	if !t.BaseJob.Validate(idReq) {
		return false
	}
	if len(t.TXT) == 0 && t.CNAME == "" && t.IpAddr == "" { // Has to test something
		return false
	}
	return true
}

func (t DNSTest) Get() BaseJob {
	return t.BaseJob
}


type PrometheusTest struct {
	MetricTests []poll.MetricTest `json:"metric_tests"`
	BaseJob
}

func (t PrometheusTest) RunTest() (time.Duration, error) {
	return poll.Prometheus(t.Url, t.Timeout * time.Second, t.MetricTests)
}

func (t PrometheusTest) Validate(idReq bool) bool {
	if !t.BaseJob.Validate(idReq) {
		return false
	}
	/*if len(t.MetricTests) == 0 {
		return false
	}*/
	return true
}

func (t PrometheusTest) Get() BaseJob {
	return t.BaseJob
}