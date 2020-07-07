package tests

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"pingr/internal/dao"
	"pingr/internal/scheduler"
	"time"
)

func Init(g *echo.Group) {
	// Get all Tests
	g.GET("", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)

		tests, err := dao.GetTests(db)
		if err != nil {
			return context.String(500, "Failed to get test, :" +  err.Error())
		}

		return context.JSON(200, tests)
	})

	// Get a Test
	g.GET("/:testId", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)
		testId:= context.Param("testId")

		test, err := dao.GetTest(testId, db)
		if err != nil {
			return context.String(500, "Failed to get test, " + err.Error())
		}

		return context.JSON(200, test)
	})

	// Get a Test's Logs
	g.GET("/:testId/logs", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)
		testId:= context.Param("testId")

		logs, err := dao.GetTestLogs(testId, db)
		if err != nil {
			return context.String(500, "Failed to get the test's logs, " + err.Error())
		}
		return context.JSON(200, logs)
	})

	// Add new Test
	g.POST("", func(context echo.Context) error {
		var testDB dao.Test
		if err := context.Bind(&testDB); err != nil {
			return context.String(400, "Could not parse body as test type: " + err.Error())
		}

		testDB.CreatedAt = time.Now()
		testDB.TestId = uuid.New().String()

		pTest, err := testDB.Parse()
		if err != nil {
			return context.String(400,"Could not parse test data: " + err.Error())
		}
		if !pTest.Validate() {
			return context.String(400,"invalid input: Test")
		}

		db := context.Get("DB").(*sqlx.DB)
		err = dao.PostTest(testDB, db)
		if err != nil {
			return context.String(500, "Could not add Test to DB, " +  err.Error())
		}

		scheduler.NotifyNewTest(pTest)

		return context.String(200, "Test added to DB")
	})

	// Update Test
	g.PUT("", func(context echo.Context) error {
		var testDB dao.Test
		if err := context.Bind(&testDB); err != nil {
			return context.String(400, "Could not parse body as test type")
		}

		testDB.CreatedAt = time.Now()

		pTest, err := testDB.Parse()
		if err != nil {
			return err
		}
		if !pTest.Validate() {
			return context.String(400,"invalid input: Test")
		}

		db := context.Get("DB").(*sqlx.DB)
		_, err = dao.GetTest(testDB.TestId, db)
		if err != nil {
			return context.String(400, "Not a valid/active testId, " + err.Error())
		}

		err = dao.PutTest(testDB, db)
		if err != nil {
			return context.String(500, "Could not update Test, " + err.Error())
		}

		scheduler.NotifyNewTest(pTest)

		return context.JSON(200, pTest)
	})

	// Delete Test
	g.DELETE("/:testId", func(context echo.Context) error {
		testId:= context.Param("testId")

		db := context.Get("DB").(*sqlx.DB)
		_, err := dao.GetTest(testId, db)
		if err != nil {
			return context.String(400, "Not a valid/active testId, " + err.Error())
		}

		err = dao.DeleteTest(testId, db)
		if err != nil {
			return context.String(500, "Could not delete Test, " + err.Error())
		}

		err = dao.DeleteTestContacts(testId, db)
		if err != nil {
			return context.String(500, "Could not delete the test's contacts: " + err.Error())
		}

		scheduler.NotifyDeletedTest(testId)

		return context.String(200, "Test deleted")
	})

	g.POST("/test", func(c echo.Context) error {
		var testDB dao.Test
		if err := c.Bind(&testDB); err != nil {
			return c.String(400, "Could not parse body as test type: " + err.Error())
		}

		testDB.CreatedAt = time.Now()
		testDB.TestId = uuid.New().String()

		pTest, err := testDB.Parse()
		if err != nil {
			return c.String(400,"Could not parse test data: " + err.Error())
		}
		if !pTest.Validate() {
			return c.String(400,"invalid input: Test")
		}

		rt, err := pTest.RunTest()
		if err != nil {
			return c.String(200, "test failed: " + err.Error())
		}
		return c.String(200, "test succeeded. response time: " + rt.Round(time.Millisecond).String())
	})

}
