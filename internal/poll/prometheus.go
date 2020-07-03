package poll

import (
	"errors"
	"fmt"
	"github.com/prometheus/common/expfmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type MetricTest struct {
	Key			string 				`json:"key"`
	LowerBound	float64 			`json:"lower_bound"`
	UpperBound	float64 			`json:"upper_bound"`
	Labels		map[string]string 	`json:"labels"`
}

func (t MetricTest) Validate() bool {
	if t.Key == "" {
		return false
	}
	if t.LowerBound > t.UpperBound {
		return false
	}
	return true
}

var (
	prevPromValues = make(map[string]float64)
	mu sync.RWMutex
)


func Prometheus(jobId uint64, url string, timeout time.Duration, metricTests []MetricTest) (time.Duration, error) {
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
	metricFamilies, err := tp.TextToMetricFamilies(strings.NewReader(string(body)))
	if err != nil {
		return 0, err
	}
	for _, metricTest := range metricTests {
		keyMetricFamily, ok := metricFamilies[metricTest.Key]
		if !ok {
			return time.Since(start), errors.New("invalid prometheus key")
		}
		for _, keyMetric := range keyMetricFamily.Metric {
			labelCount := 0
			for _, labelPair := range keyMetric.Label {
				value, ok := metricTest.Labels[*labelPair.Name]
				if ok && value == *labelPair.Value {
					labelCount++
				}
			}
			if labelCount != len(metricTest.Labels) {
				continue
			}

			switch keyMetricFamily.Type.String() {
			case "GAUGE":
				promValue := *keyMetric.Gauge.Value
				if metricTest.LowerBound > promValue || promValue > metricTest.UpperBound {
					return rt, errors.New(fmt.Sprintf("expected key: %s GAUGE to be between %f and %f got: %f", metricTest.Key, metricTest.LowerBound, metricTest.UpperBound, promValue))
				}
			case "COUNTER":
				mu.Lock()
				promValue := *keyMetric.Counter.Value
				hashedMetric := hash(jobId, metricTest.Key, metricTest.Labels)
				if _, ok := prevPromValues[hashedMetric]; !ok {
					prevPromValues[hashedMetric] = promValue
					mu.Unlock()
					continue
				}
				promValueIncrease := promValue - prevPromValues[hashedMetric]
				prevPromValues[hashedMetric] = promValue
				if promValueIncrease < metricTest.LowerBound || promValueIncrease > metricTest.UpperBound {
					mu.Unlock()
					return rt, errors.New(fmt.Sprintf("expected key: %s COUNTER to increase between %f and %f got: %f", metricTest.Key, metricTest.LowerBound, metricTest.UpperBound, promValueIncrease))
				}
				mu.Unlock()
			}
		}
	}
	return rt, nil
}


func hash(jobId uint64, promKey string, labels map[string]string) string {
	hashedString := string(jobId)
	hashedString+=promKey

	var labelsSlice []string
	for k, v := range labels {
		labelsSlice = append(labelsSlice, k, v)
	}
	sort.Slice(labelsSlice, func(i, j int) bool {
		return labelsSlice[i] < labelsSlice[j]
	})
	hashedString+=strings.Join(labelsSlice, "")
	return hashedString
}