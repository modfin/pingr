package testcontacts

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"pingr"
	"pingr/internal/dao"
	)

func Init(g *echo.Group) {
	g.GET("", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)
		jContacts, err := dao.GetAllTestContacts(db)
		if err != nil {
			return context.String(500, "Failed to get test contacts:" +  err.Error())
		}

		return context.JSON(200, jContacts)
	})

	g.GET("/:testId", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)
		testId:= context.Param("testId")

		jContacts, err := dao.GetTestContacts(testId, db)
		if err != nil {
			return context.String(500, "Failed to get test contacts:" +  err.Error())
		}

		return context.JSON(200, jContacts)
	})

	g.GET("/:testId/types", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)
		testId:= context.Param("testId")

		jContactsType, err := dao.GetTestContactsType(testId, db)
		if err != nil {
			return context.String(500, "Failed to get test contacts:" +  err.Error())
		}

		return context.JSON(200, jContactsType)
	})

	g.POST("",func(context echo.Context) error {
		var contacts []pingr.TestContact
		if err := context.Bind(&contacts); err != nil {
			return context.String(400, "Could not parse body as test contact type: " + err.Error())
		}
		if len(contacts) == 0 {
			return context.String(400, "Could not parse body as test contact type")
		}

		db := context.Get("DB").(*sqlx.DB)

		for _, contact := range contacts {
			if !contact.Validate() {
				return context.String(400, "invalid input: TestContact")
			}

			_, err := dao.GetTest(contact.TestId, db)
			if err != nil {
				return context.String(400, "no matching test id or/and contact id")
			}
			_, err = dao.GetContact(contact.ContactId, db)
			if err != nil {
				return context.String(400, "no matching test id or/and contact id")
			}

			err = dao.PostTestContact(contact, db)
			if err != nil {
				return context.String(500, "could not add test contact to db: " + err.Error())
			}
		}

		return context.String(200, "test contacts added to db")
	})

	g.PUT("", func(context echo.Context) error {
		var contacts []pingr.TestContact
		if err := context.Bind(&contacts); err != nil {
			return context.String(400, "Could not parse body as test contact type: " + err.Error())
		}
		if len(contacts) == 0 {
			return context.String(400, "Could not parse body as test contact type")
		}

		db := context.Get("DB").(*sqlx.DB)

		testId := contacts[0].TestId
		_, err := dao.GetTestContacts(testId, db)
		if err != nil {
			return context.String(400, "Not a valid/active testId: " + err.Error())
		}
		err = dao.DeleteTestContacts(testId, db)
		if err != nil {
			return context.String(500, "Could not delete test contacts: " + err.Error())
		}

		for _, contact := range contacts {
			if !contact.Validate() {
				return context.String(400, "invalid input: TestContact")
			}

			if testId != contact.TestId {
				return context.String(400, "all contacts has to have same testId")
			}
			_, err = dao.GetContact(contact.ContactId, db)
			if err != nil {
				return context.String(400, "no matching test id or/and contact id")
			}

			err = dao.PostTestContact(contact, db)
			if err != nil {
				return context.String(500, "could not add test contact to db: " + err.Error())
			}
		}

		return context.String(200, "test contacts added to db")
	})

	g.DELETE("/:testId", func(context echo.Context) error {
		testId:= context.Param("testId")

		db := context.Get("DB").(*sqlx.DB)
		_, err := dao.GetTestContacts(testId, db)
		if err != nil {
			return context.String(400, "Not a valid/active testId: " + err.Error())
		}

		err = dao.DeleteTestContacts(testId, db)
		if err != nil {
			return context.String(500, "Could not delete test contacts: " + err.Error())
		}

		return context.String(200, "test contacts deleted")
	})

	g.DELETE("/:testId/:contactId", func(context echo.Context) error {
		testId:= context.Param("testId")
		contactId:= context.Param("contactId")

		db := context.Get("DB").(*sqlx.DB)
		_, err := dao.GetTestContacts(testId, db)
		if err != nil {
			return context.String(400, "Not a valid/active testId: " + err.Error())
		}

		err = dao.DeleteTestContact(testId, contactId, db)
		if err != nil {
			return context.String(500, "Could not delete test contact: " + err.Error())
		}

		return context.String(200, "test contact deleted")
	})
}