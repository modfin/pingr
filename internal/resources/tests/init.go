package tests

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"pingr"
	"pingr/internal/bus"
	"pingr/internal/dao"
	"time"
)

func Init(g *echo.Group, buz *bus.Bus) {
	// Get all Tests
	g.GET("", func(c echo.Context) error {
		db := c.Get("DB").(*sqlx.DB)

		tests, err := dao.GetRawTests(db)
		if err != nil {
			return c.String(500, "Failed to get test: "+err.Error())
		}

		for i := range tests {
			err = tests[i].MaskSensitiveInfo(pingr.GET, nil)
			if err != nil {
				return c.String(500, "could not mask sensitive test data: "+err.Error())
			}
		}

		return c.JSON(200, tests)
	})

	// Get a Test
	g.GET("/:testId", func(c echo.Context) error {
		db := c.Get("DB").(*sqlx.DB)
		testId := c.Param("testId")

		test, err := dao.GetTestStatus(testId, db)
		if err != nil {
			return c.String(500, "Failed to get test, "+err.Error())
		}

		err = test.MaskSensitiveInfo(pingr.GET, nil)
		if err != nil {
			return c.String(500, "Failed to mask test data: "+err.Error())
		}

		return c.JSON(200, test)
	})

	g.GET("/status", func(c echo.Context) error {
		db := c.Get("DB").(*sqlx.DB)

		testStatus, err := dao.GetTestsStatus(db)
		if err != nil {
			return c.String(500, "Failed to get test, "+err.Error())
		}

		return c.JSON(200, testStatus)
	})

	// Get a Test's Logs
	g.GET("/:testId/logs", func(c echo.Context) error {
		db := c.Get("DB").(*sqlx.DB)
		testId := c.Param("testId")

		logs, err := dao.GetTestLogs(testId, db)
		if err != nil {
			return c.String(500, "Failed to get the test's logs, "+err.Error())
		}
		return c.JSON(200, logs)
	})

	// Get a Test's Logs limited
	g.GET("/:testId/logs/:days", func(c echo.Context) error {
		db := c.Get("DB").(*sqlx.DB)
		testId := c.Param("testId")
		days := c.Param("days")

		logs, err := dao.GetTestLogsDaysLimited(testId, days, db)
		if err != nil {
			return c.String(500, "Failed to get the test's logs, "+err.Error())
		}
		return c.JSON(200, logs)
	})

	// Add new Test
	g.POST("", func(c echo.Context) error {
		var testDB pingr.GenericTest
		if err := c.Bind(&testDB); err != nil {
			return c.String(400, "Could not parse body as test type: "+err.Error())
		}

		testDB.CreatedAt = time.Now()
		testDB.TestId = uuid.New().String()

		if !testDB.Validate() {
			return c.String(400, "invalid input: Test")
		}

		err := testDB.MaskSensitiveInfo(pingr.POST, nil)
		if err != nil {
			return c.String(400, "Could not mask sensitive test data: "+err.Error())
		}

		db := c.Get("DB").(*sqlx.DB)
		err = dao.PostTest(testDB, db)
		if err != nil {
			return c.String(500, "Could not add Test to DB, "+err.Error())
		}

		data, err := json.Marshal(testDB)
		if err != nil {
			return c.String(500, fmt.Sprintf("could not marchal test: %s", err.Error()))
		}
		err = buz.Publish("new", data)
		if err != nil {
			return c.String(500, fmt.Sprintf("unable to publish new test: %s", err.Error()))
		}

		return c.JSON(200, testDB)
	})

	g.PUT("/:testId/deactivate", func(c echo.Context) error {
		db := c.Get("DB").(*sqlx.DB)
		testId := c.Param("testId")
		err := dao.DeactivateTest(testId, db)
		if err != nil {
			return c.String(400, "could not deactivate test: "+err.Error())
		}
		err = buz.Publish("deactivate", []byte(testId))
		if err != nil {
			return c.String(400, "could not publish deactivation: "+err.Error())
		}

		return c.String(200, "test paused")
	})

	g.PUT("/:testId/activate", func(c echo.Context) error {
		db := c.Get("DB").(*sqlx.DB)
		testId := c.Param("testId")

		test, err := dao.GetRawTest(testId, db)
		if err != nil {
			return c.String(400, "invalid test id: "+err.Error())
		}

		err = dao.ActivateTest(testId, db)
		if err != nil {
			return c.String(400, "could not activate test: "+err.Error())
		}

		data, err := json.Marshal(test)
		if err != nil {
			return c.String(500, fmt.Sprintf("could not marchal test: %s", err.Error()))
		}
		err = buz.Publish("new", data)
		if err != nil {
			return c.String(400, "could not publish activation: "+err.Error())
		}

		return c.String(200, "test activated")
	})

	// Update Test
	g.PUT("", func(c echo.Context) error {
		var testDB pingr.GenericTest
		if err := c.Bind(&testDB); err != nil {
			return c.String(400, "Could not parse body as test type")
		}

		testDB.CreatedAt = time.Now()

		db := c.Get("DB").(*sqlx.DB)
		testDb, err := dao.GetRawTest(testDB.TestId, db)
		if err != nil {
			return c.String(400, "Not a valid/active testId, "+err.Error())
		}

		err = testDB.MaskSensitiveInfo(pingr.PUT, &testDb)
		if err != nil {
			return c.String(400, "could not mask sensitive test data: "+err.Error())
		}

		if !testDB.Validate() {
			return c.String(400, "invalid input: Test")
		}

		err = dao.PutTest(testDB, db)
		if err != nil {
			return c.String(500, "Could not update Test, "+err.Error())
		}

		if testDB.Active {
			data, err := json.Marshal(testDB)
			if err != nil {
				return c.String(500, fmt.Sprintf("could not marchal test: %s", err.Error()))
			}
			err = buz.Publish("new", data)
			if err != nil {
				return c.String(500, fmt.Sprintf("unable to publish new test: %s", err.Error()))
			}
		}

		return c.JSON(200, testDB)
	})

	// Delete Test
	g.DELETE("/:testId", func(c echo.Context) error {
		testId := c.Param("testId")

		db := c.Get("DB").(*sqlx.DB)
		_, err := dao.GetTest(testId, db)
		if err != nil {
			return c.String(400, "Not a valid/active testId, "+err.Error())
		}

		err = dao.DeleteTest(testId, db)
		if err != nil {
			return c.String(500, "Could not delete Test, "+err.Error())
		}

		err = dao.DeleteTestContacts(testId, db)
		if err != nil {
			return c.String(500, "Could not delete the test's contacts: "+err.Error())
		}

		err = dao.DeleteTestLogs(testId, db)
		if err != nil {
			return c.String(500, "Could not delete the test's logs: "+err.Error())
		}

		err = dao.CloseTestIncident(testId, db)
		if err != nil {
			return c.String(500, "Could not close the test's incident: "+err.Error())
		}

		err = buz.Publish("delete", []byte(testId))
		if err != nil {
			return c.String(500, fmt.Sprintf("unable to publish deletion: %s", err.Error()))
		}

		return c.String(200, "Test deleted")
	})

	g.POST("/test", func(c echo.Context) error {
		var testDB pingr.GenericTest
		if err := c.Bind(&testDB); err != nil {
			return c.String(400, "Could not parse body as test type: "+err.Error())
		}

		testDB.CreatedAt = time.Now()
		testDB.TestId = uuid.New().String()

		pTest, err := testDB.Impl()
		if err != nil {
			return c.String(400, "Could not parse test data: "+err.Error())
		}
		if !pTest.Validate() {
			return c.String(400, "invalid input: Test")
		}

		err = testDB.MaskSensitiveInfo(pingr.POST, nil)
		if err != nil {
			return c.String(400, "Could not mask sensitive test data: "+err.Error())
		}

		pTest, err = testDB.Impl()
		if err != nil {
			return c.String(400, "Could not parse test data: "+err.Error())
		}

		rt, err := pTest.RunTest(buz)
		if err != nil {
			return c.String(200, "test failed: "+err.Error())
		}
		return c.String(200, "test succeeded. response time: "+rt.Round(time.Millisecond).String())
	})

}
