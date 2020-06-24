package health

import (
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
)

func SetMetrics(e *echo.Echo) {
	prom := prometheus.NewPrometheus("pingr_pingr_echo", nil)
	prom.MetricsPath = "/health/metrics"
	prom.Use(e)
}

func Init(closing <-chan struct{}, g *echo.Group) {

	g.GET("/poll", func(context echo.Context) error {
		return context.String(200, "pong")
	})
}
