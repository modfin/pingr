package logging

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"time"
)

func EchoMiddleware(skipper middleware.Skipper) echo.MiddlewareFunc {
	if skipper == nil {
		skipper = middleware.DefaultSkipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if skipper(c) {
				return next(c)
			}

			entry := logrus.NewEntry(logrus.StandardLogger())

			req := c.Request()
			res := c.Response()

			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)
				entry = entry.WithError(err)
			}
			stop := time.Now()

			entry = entry.WithContext(req.Context())
			entry = entry.WithField("type", "echo")
			entry = entry.WithField("duration", stop.Sub(start))
			entry = entry.WithField("uri", req.RequestURI)
			entry = entry.WithField("method", req.Method)
			entry = entry.WithField("host", req.Host)
			entry = entry.WithField("Status", res.Status)

			if err != nil {
				entry.Error()
				return
			}

			if res.Status > 299 {
				entry.Warn()
				return
			}
			entry.Info()
			return
		}
	}
}

func RequestIdMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			var requestId string
			ctx := c.Request().Context()
			infId := ctx.Value(echo.HeaderXRequestID)
			if infId != nil {
				requestId, _ = infId.(string)
			}
			if infId == nil {
				requestId = c.Request().Header.Get(echo.HeaderXRequestID)
			}
			if requestId == "" {
				requestId = uuid.New().String()
			}
			ctx = context.WithValue(ctx, echo.HeaderXRequestID, requestId)
			c.Request().Header.Set(echo.HeaderXRequestID, requestId)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Set("request_id", requestId)

			return next(c)
		}
	}
}

func GetDBMiddleware(db *sql.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			c.Set("DB", db)
			return next(c)
		}
	}
}
