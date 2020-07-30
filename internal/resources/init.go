package resources

import (
	"context"
	"crypto/subtle"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"pingr/internal/bus"
	"pingr/internal/config"
	"pingr/internal/logging"
	"pingr/internal/resources/contacts"
	"pingr/internal/resources/health"
	"pingr/internal/resources/incidents"
	"pingr/internal/resources/logs"
	"pingr/internal/resources/push"
	"pingr/internal/resources/testcontacts"
	"pingr/internal/resources/tests"
	"pingr/ui"
)

func Init(closing <-chan struct{}, db *sqlx.DB, buz *bus.Bus) {
	cfg := config.Get()
	basicAuth := middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		validUser := false
		validPass := false
		if subtle.ConstantTimeCompare([]byte(username), []byte(cfg.BasicAuthUser)) == 1 {
			validUser = true
		}
		if subtle.ConstantTimeCompare([]byte(password), []byte(cfg.BasicAuthPass)) == 1 {
			validPass = true
		}
		return validUser && validPass, nil
	})


	e := echo.New()
	e.Use(logging.RequestIdMiddleware())
	e.Use(logging.EchoMiddleware(nil))
	e.Use(logging.GetDBMiddleware(db))
	e.Use(middleware.Recover())

	health.SetMetrics(e)
	health.Init(closing, e.Group("api/health"))

	tests.Init(e.Group("api/tests", basicAuth), buz)
	logs.Init(e.Group("api/logs", basicAuth))
	contacts.Init(e.Group("api/contacts", basicAuth))
	testcontacts.Init(e.Group("api/testcontacts", basicAuth))
	incidents.Init(e.Group("api/incidents", basicAuth))

	push.Init(e.Group("api/push"), buz)

	// UI
	e.GET("/*", func(c echo.Context) error {
		if config.Get().Dev {
			u, err := url.Parse("http://ui:8080")
			if err != nil {
				return err
			}
			proxy := httputil.NewSingleHostReverseProxy(u)
			proxy.ServeHTTP(c.Response().Writer, c.Request())
			return nil
		}
		p, err := url.PathUnescape(c.Param("*"))
		if err != nil {
			return err
		}
		if p == "" {
			p = "index.html"
		}
		data, err := ui.Asset(path.Clean(p))
		if err != nil {
			data, err = ui.Asset(path.Clean("index.html"))
		}
		if err != nil {
			return err
		}
		return c.Blob(200, http.DetectContentType(data), data)
	}, basicAuth)

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
