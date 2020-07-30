package logging

import (
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"os"
	"pingr/internal/config"
	"pingr/internal/metrics"
)

func SetDefault() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if config.Get().Dev {
		logrus.SetLevel(logrus.DebugLevel)

		customFormatter := &logrus.TextFormatter{}
		customFormatter.FullTimestamp = true
		customFormatter.TimestampFormat = "2006-01-02 15:04:05"
		logrus.SetFormatter(customFormatter)
	}

	logrus.AddHook(&CtxHook{})
	logrus.AddHook(&MetricHook{})
}

type CtxHook struct {
}

func (h *CtxHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
func (h *CtxHook) Fire(e *logrus.Entry) error {
	ctx := e.Context
	if ctx != nil {
		id := ctx.Value(echo.HeaderXRequestID)
		if id != nil {
			e.Data["request_id"] = id
		}
	}
	return nil
}

type MetricHook struct {
}

func (h *MetricHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
func (h *MetricHook) Fire(e *logrus.Entry) error {
	if e.Level <= logrus.GetLevel() {
		metrics.LogEntriesInc()
	}
	return nil
}
