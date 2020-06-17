package resources

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"pingr/internal/config"
	"pingr/internal/logging"
	"pingr/internal/resources/health"
)

func Init(closing <-chan struct{}) {

	e := echo.New()
	e.Use(logging.RequestIdMiddleware())
	e.Use(logging.EchoMiddleware(nil))
	e.Use(middleware.Recover())

	health.SetMetrics(e)
	health.Init(closing, e.Group("health"))

	go e.Start(fmt.Sprintf(":%d", config.Get().Port))

	<-closing
	c, cancel := context.WithTimeout(context.Background(), config.Get().TermDuration)
	defer cancel()
	logrus.Info("Gracefully closing Echo")
	err := e.Shutdown(c)
	if err != nil {
		logrus.Warn("Could not gracefully close Echo, will force it")
		_ = e.Close()
	}
	logrus.Info("Echo has been shutdown")
}
