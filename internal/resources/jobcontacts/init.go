package jobcontacts

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"pingr"
	"pingr/internal/dao"
	"strconv"
)

func Init(g *echo.Group) {
	g.GET("", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)
		jContacts, err := dao.GetAllJobContacts(db)
		if err != nil {
			return context.String(500, "Failed to get job contacts:" +  err.Error())
		}

		return context.JSON(200, jContacts)
	})

	g.GET("/:jobId", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)
		jobIdStr:= context.Param("jobId")

		jobId, err := strconv.ParseUint(jobIdStr, 10, 64)
		if err != nil {
			return context.String(500, "Could not parse jobId as int")
		}

		jContacts, err := dao.GetJobContacts(jobId, db)
		if err != nil {
			return context.String(500, "Failed to get job contacts:" +  err.Error())
		}

		return context.JSON(200, jContacts)
	})

	g.POST("/add",func(context echo.Context) error {
		var contacts []pingr.JobContact
		if err := context.Bind(&contacts); err != nil {
			return context.String(500, "Could not parse body as job contact type: " + err.Error())
		}
		if len(contacts) == 0 {
			return context.String(500, "Could not parse body as job contact type")
		}

		db := context.Get("DB").(*sqlx.DB)
		for _, contact := range contacts {
			if !contact.Validate() {
				return context.String(500, "invalid input: JobContact")
			}

			err := dao.PostJobContact(contact, db)
			if err != nil {
				return context.String(500, "could not add job contact to db: " + err.Error())
			}
		}

		return context.String(200, "job contacts added to db")
	})

	g.DELETE("/delete/:jobId", func(context echo.Context) error {
		jobIdStr:= context.Param("jobId")

		jobId, err := strconv.ParseUint(jobIdStr, 10, 64)
		if err != nil {
			return context.String(500, "Could not parse jobId as uint")
		}
		db := context.Get("DB").(*sqlx.DB)
		_, err = dao.GetJobContacts(jobId, db)
		if err != nil {
			return context.String(500, "Not a valid/active jobId: " + err.Error())
		}

		err = dao.DeleteJobContacts(jobId, db)
		if err != nil {
			return context.String(500, "Could not delete job contacts: " + err.Error())
		}

		return context.String(500, "job contacts deleted")
	})

	g.DELETE("/delete/:jobId/:contactId", func(context echo.Context) error {
		jobIdStr:= context.Param("jobId")
		contactIdStr:= context.Param("contactId")

		jobId, err := strconv.ParseUint(jobIdStr, 10, 64)
		if err != nil {
			return context.String(500, "Could not parse jobId as uint")
		}
		contactId, err := strconv.ParseUint(contactIdStr, 10, 64)
		if err != nil {
			return context.String(500, "Could not parse contactId as uint")
		}
		db := context.Get("DB").(*sqlx.DB)
		_, err = dao.GetJobContacts(jobId, db)
		if err != nil {
			return context.String(500, "Not a valid/active jobId: " + err.Error())
		}

		err = dao.DeleteJobContact(jobId, contactId, db)
		if err != nil {
			return context.String(500, "Could not delete job contact: " + err.Error())
		}

		return context.String(500, "job contact deleted")
	})

	g.GET("/test", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)
		logrus.Info("hej")
		asd, err := dao.GetJobContactsType(1, db)
		if err != nil {
			return context.String(500, err.Error())
		}
		return context.JSON(200, asd)
	})
}