package push

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"pingr/internal/dao"
	"pingr/internal/scheduler"
)

func Init(g *echo.Group) {
	// Listen to push requests
	g.GET("/:test-id/:vanity-name", func(context echo.Context) error {
		testId:= context.Param("test-id")
		db := context.Get("DB").(*sqlx.DB)

		_, err := dao.GetTest(testId, db)
		if err != nil {
			return context.String(400, "invalid testId")
		}

		err = scheduler.NotifyPushTest(testId, nil)
		if err != nil {
			return context.String(500, err.Error())
		}

		return context.String(200, "Push request received")
	})

	g.POST("/:test-id/:vanity-name", func(context echo.Context) error {
		testId:= context.Param("test-id")
		db := context.Get("DB").(*sqlx.DB)

		_, err := dao.GetTest(testId, db)
		if err != nil {
			return context.String(400, "invalid testId")
		}

		reqBody, err := ioutil.ReadAll(context.Request().Body)
		if err != nil {
			return context.String(400, "could not read post body")
		}

		// Notify worker of push retrieval
		err = scheduler.NotifyPushTest(testId, reqBody)
		if err != nil {
			return context.String(500, err.Error())
		}

		return context.String(200, "Push request received")
	})
}