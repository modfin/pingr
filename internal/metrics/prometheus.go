package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var HealthCheckInc = healthChecksPerformed.Inc
var healthChecksPerformed = promauto.NewCounter(prometheus.CounterOpts{
	Name: "pingr_health_checks_performed",
	Help: "The total number of health checks that has been performed",
})

var LogEntriesInc = logEntires.Inc
var logEntires = promauto.NewCounter(prometheus.CounterOpts{
	Name: "pingr_log_entries",
	Help: "The total number of log entries made by the service",
})
