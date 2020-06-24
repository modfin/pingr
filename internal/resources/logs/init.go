package logs

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"pingr/internal/dao"
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
	g.GET("/:logId", func(context echo.Context) error {
		db := context.Get("DB").(*sql.DB)
		logId := context.Param("logId")

		log, err := dao.GetLog(logId, db)
		if err != nil {
			return context.String(500, "Failed to get log, " + err.Error())
		}

		return context.JSON(200, *log)
	})

}
