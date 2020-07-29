package push

import (
	"errors"
	"fmt"
	"github.com/prometheus/common/expfmt"
	"sort"
	"strings"
	"sync"
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

func Prometheus(testId string, body []byte, metricTests []MetricTest) error {
	tp := expfmt.TextParser{}
	metricFamilies, err := tp.TextToMetricFamilies(strings.NewReader(string(body)))

	if err != nil {
		return err
	}
	for _, metricTest := range metricTests {
		keyMetricFamily, ok := metricFamilies[metricTest.Key]
		if !ok {
			return fmt.Errorf("invalid prometheus key: %s", metricTest.Key)
		}
		oneMatch := false
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
			oneMatch = true

			switch keyMetricFamily.Type.String() {
			case "GAUGE":
				promValue := *keyMetric.Gauge.Value
				if metricTest.LowerBound > promValue || promValue > metricTest.UpperBound {
					return errors.New(fmt.Sprintf("expected key: %s GAUGE to be between %.3f and %.3f got: %.3f", metricTest.Key, metricTest.LowerBound, metricTest.UpperBound, promValue))
				}
			case "COUNTER":
				mu.Lock()
				promValue := *keyMetric.Counter.Value
				hashedMetric := hash(testId, metricTest.Key, metricTest.Labels)
				if _, ok := prevPromValues[hashedMetric]; !ok {
					// First value, nothing to compare against
					prevPromValues[hashedMetric] = promValue
					mu.Unlock()
					continue
				}
				promValueIncrease := promValue - prevPromValues[hashedMetric]
				prevPromValues[hashedMetric] = promValue
				if promValueIncrease < metricTest.LowerBound || promValueIncrease > metricTest.UpperBound {
					mu.Unlock()
					return errors.New(fmt.Sprintf("expected key: %s COUNTER to increase between %.3f and %.3f got: %.3f", metricTest.Key, metricTest.LowerBound, metricTest.UpperBound, promValueIncrease))
				}
				mu.Unlock()
			}
		}
		if !oneMatch {
			return fmt.Errorf("no mathing labels for prometheus key: %s with labels: %v", metricTest.Key, metricTest.Labels)
		}
	}
	return err
}


func hash(testId string, promKey string, labels map[string]string) string {
	hashedString := testId
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