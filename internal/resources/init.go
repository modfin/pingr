package resources

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"pingr/internal/config"
	"pingr/internal/logging"
	"pingr/internal/resources/health"
	"pingr/internal/resources/jobs"
	"pingr/internal/resources/logs"
	"pingr/ui"
)

func Init(closing <-chan struct{}, db *sql.DB) {

	e := echo.New()
	e.Use(logging.RequestIdMiddleware())
	e.Use(logging.EchoMiddleware(nil))
	e.Use(logging.GetDBMiddleware(db))
	e.Use(middleware.Recover())

	health.SetMetrics(e)
	health.Init(closing, e.Group("health"))

	jobs.Init(e.Group("jobs"))

	logs.Init(e.Group("logs"))


	// UI
	e.GET("/*", func(c echo.Context) error {
		p, err := url.PathUnescape(c.Param("*"))
		if  err != nil{
			return err
		}
		if p == ""{
			p = "index.html"
		}
		if config.Get().Dev {
			name := filepath.Join("./ui/dist", path.Clean("/"+p)) // "/"+ for security
			return c.File(name)
		}
		name := filepath.Join("build", path.Clean(p))
		data, err := ui.Asset(name)
		if  err != nil{
			return err
		}
		return c.Blob(200, http.DetectContentType(data), data)
	})



	// Setup endpoints for SQLite DB? POST/PUT/DELETE Jobs? Get logs of jobs

	// Init scheduler?

	// Init push?

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
