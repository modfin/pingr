package resources

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"pingr/internal/bus"
	"pingr/internal/config"
	"pingr/internal/logging"
	"pingr/internal/resources/contacts"
	"pingr/internal/resources/health"
	"pingr/internal/resources/testcontacts"
	"pingr/internal/resources/tests"
	"pingr/internal/resources/logs"
	"pingr/internal/resources/push"
	"pingr/ui"
)

func Init(closing <-chan struct{}, db *sqlx.DB, buz *bus.Bus) {

	e := echo.New()
	e.Use(logging.RequestIdMiddleware())
	e.Use(logging.EchoMiddleware(nil))
	e.Use(logging.GetDBMiddleware(db))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	health.SetMetrics(e)
	health.Init(closing, e.Group("api/health"))

	tests.Init(e.Group("api/tests"), buz)
	logs.Init(e.Group("api/logs"))
	contacts.Init(e.Group("api/contacts"))
	testcontacts.Init(e.Group("api/testcontacts"))

	push.Init(e.Group("api/push"), buz)

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
