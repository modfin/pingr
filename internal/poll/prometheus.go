package poll

import (
	"io/ioutil"
	"net/http"
	"pingr/internal/push"
	"sync"
	"time"
)



var (
	prevPromValues = make(map[string]float64)
	mu sync.RWMutex
)


func Prometheus(testId string, url string, timeout time.Duration, metricTests []push.MetricTest) (time.Duration, error) {
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

	err = push.Prometheus(testId, body, metricTests)
	if err != nil {
		return rt, err
	}

	return rt, nil
}
