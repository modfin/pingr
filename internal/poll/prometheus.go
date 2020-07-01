package poll

import (
	"fmt"
	"github.com/prometheus/common/expfmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type MetricTest struct {
	Key			string 	`json:"key"`
	Method    	string 	`json:"method"`
	Threshold 	float32 `json:"threshold"`
	IncTest 	bool	`json:"inc_test"`
}

func (t MetricTest) TestThreshold(currentValue float32, prevValue float32) bool {
	comparingValue := currentValue
	if t.IncTest {
		comparingValue -= prevValue
	}
	switch t.Method {
	case "le": // less or equal
		return comparingValue <= t.Threshold
	case "l": // less
		return comparingValue < t.Threshold
	case "ge": // greater or equal
		return comparingValue >= t.Threshold
	case "g": // greater
		return comparingValue > t.Threshold
	case "e": // equal
		return comparingValue == t.Threshold
	case "ne": // not equal
		return comparingValue != t.Threshold
	default:
		return false
	}
}
func (t MetricTest) Validate() bool {
	if t.Key == "" {
		return false
	}
	switch t.Method {
	case "le","l","ge","g","e","ne":
	default:
		return false
	}
	return true
}

func Prometheus(url string, timeout time.Duration, pTests []MetricTest) (time.Duration, error) {
	start := time.Now()
	
	client:= http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil
	}
	rt := time.Since(start)
	
	tp := expfmt.TextParser{}
	m, err := tp.TextToMetricFamilies(strings.NewReader(string(body)))
	if err != nil {
		return 0, err
	}
	for _, pTest := range pTests {
		fmt.Print(pTest)
	}
	/*for _, metric := range m["scylla_storage_proxy_coordinator_cas_read_contention_count"].Metric {
		logrus.Info(metric)
	}*/
	logrus.Info(m["scylla_storage_proxy_coordinator_cas_write_contention_bucket"])

	return rt, nil
}
