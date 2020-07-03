package contacts

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"pingr"
	"pingr/internal/dao"
	"strconv"
)

func Init(g *echo.Group) {
	g.GET("", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)
		contacts, err := dao.GetContacts(db)
		if err != nil {
			return context.String(500, "Failed to get contacts, :" +  err.Error())
		}

		return context.JSON(200, contacts)
	})

	g.GET("/:contactId", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)
		contactIdStr:= context.Param("contactId")

		contactId, err := strconv.ParseUint(contactIdStr, 10, 64)
		if err != nil {
			return context.String(500, "Could not parse contactId as int")
		}

		contact, err := dao.GetContact(contactId, db)
		if err != nil {
			return context.String(500, "Failed to get contact, :" +  err.Error())
		}

		return context.JSON(200, contact)
	})

	g.POST("/add",func(context echo.Context) error {
		var contact pingr.Contact
		if err := context.Bind(&contact); err != nil {
			return context.String(500, "Could not parse body as contact type: " + err.Error())
		}
		if !contact.Validate(false) {
			return context.String(500, "invalid input: Contact")
		}

		db := context.Get("DB").(*sqlx.DB)
		err := dao.PostContact(contact, db)
		if err != nil {
			return context.String(500, "could not add contact to db: " + err.Error())
		}

		return context.String(200, "contact added to db")
	})

	g.PUT("/update", func(context echo.Context) error {
		var contact pingr.Contact
		if err := context.Bind(&contact); err != nil {
			return context.String(500, "Could not parse body as contact type: " + err.Error())
		}
		if !contact.Validate(true) {
			return context.String(500, "invalid input: Contact")
		}

		db := context.Get("DB").(*sqlx.DB)
		_, err := dao.GetContact(contact.ContactId, db)
		if err != nil {
			return context.String(500, "Not a valid/active ContactId, " + err.Error())
		}

		err = dao.PutContact(contact, db)
		if err != nil {
			return context.String(500, "could not update contact: "+err.Error())
		}

		return context.String(200, "contact updated")
	})

	g.DELETE("/delete/:contactId", func(context echo.Context) error {
		contactIdStr:= context.Param("contactId")

		contactId, err := strconv.ParseUint(contactIdStr, 10, 64)
		if err != nil {
			return context.String(500, "Could not parse contactId as uint")
		}
		db := context.Get("DB").(*sqlx.DB)
		_, err = dao.GetContact(contactId, db)
		if err != nil {
			return context.String(500, "Not a valid/active contactId, " + err.Error())
		}

		err = dao.DeleteContact(contactId, db)
		if err != nil {
			return context.String(500, "Could not delete contact, " + err.Error())
		}

		return context.String(500, "contact deleted")
	})
}