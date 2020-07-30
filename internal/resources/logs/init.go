package logs

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"pingr/internal/dao"
	"strconv"
)

func Init(g *echo.Group) {
	// Get all Logs
	g.GET("", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)

		logs, err := dao.GetLogs(db)
		if err != nil {
			return context.String(500, "Failed to get logs, "+err.Error())
		}

		return context.JSON(200, logs)
	})

	// Get a Log
	g.GET("/:logId", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)
		logIdString := context.Param("logId")

		logId, err := strconv.ParseUint(logIdString, 10, 64)
		if err != nil {
			return context.String(400, "Could not parse logId as int")
		}

		log, err := dao.GetLog(logId, db)
		if err != nil {
			return context.String(500, "Failed to get log, "+err.Error())
		}

		return context.JSON(200, log)
	})

	// Delete a log
	g.DELETE("/:logId", func(context echo.Context) error {
		logIdString := context.Param("logId")
		if logIdString == "" {
			return context.String(400, "Please include logId in body")
		}

		logId, err := strconv.ParseUint(logIdString, 10, 64)
		if err != nil {
			return context.String(400, "Could not parse TestId as int")
		}

		db := context.Get("DB").(*sqlx.DB)
		err = dao.DeleteLog(logId, db)
		if err != nil {
			context.String(500, "Could not delete Log, "+err.Error())
		}

		return context.String(200, "Log deleted")
	})
}
