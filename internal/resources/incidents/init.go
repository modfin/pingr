package incidents

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"pingr/internal/dao"
)

func Init(g *echo.Group) {
	g.GET("", func(c echo.Context) error {
		db := c.Get("DB").(*sqlx.DB)
		incidents, err := dao.GetIncidents(db)
		if err != nil {
			return c.String(500, "could not get incidents: " + err.Error())
		}
		return c.JSON(200, incidents)
	})

	g.GET("/:incidentId", func(c echo.Context) error {
		db := c.Get("DB").(*sqlx.DB)
		incidentId := c.Param("incidentId")

		incidents, err := dao.GetIncident(incidentId, db)
		if err != nil {
			return c.String(500, "could not get incidents: " + err.Error())
		}
		return c.JSON(200, incidents)
	})

}