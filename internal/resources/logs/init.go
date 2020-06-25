package logs

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"pingr/internal/dao"
	"pingr/internal/scheduler"
)

func Init(g *echo.Group) {
	// Get all Logs
	g.GET("", func(context echo.Context) error {
		db := context.Get("DB").(*sql.DB)

		logs, err := dao.GetLogs(db)
		if err != nil {
			return context.String(500, "Failed to get logs, " + err.Error())
		}

		return context.JSON(200, logs)
	})

	// Get a Log
	g.GET("/:LogId", func(context echo.Context) error {
		db := context.Get("DB").(*sql.DB)
		LogId := context.Param("LogId")

		log, err := dao.GetLog(LogId, db)
		if err != nil {
			return context.String(500, "Failed to get log, " + err.Error())
		}

		return context.JSON(200, log)
	})

	// Delete a log
	g.DELETE("/delete", func(context echo.Context) error {
		logId := context.FormValue("logId")
		if logId == "" {
			return context.String(500, "Please include logId in body")
		}

		db := context.Get("DB").(*sql.DB)
		err := dao.DeleteLog(logId, db)
		if err != nil {
			context.String(500, "Could not delete Log, " + err.Error())
		}

		scheduler.Notify()

		return context.String(500, "Log deleted")
	})
}
