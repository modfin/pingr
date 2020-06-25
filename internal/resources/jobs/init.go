package jobs

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"pingr/internal/dao"
	"pingr/internal/scheduler"
	"time"
)

func Init(g *echo.Group) {
	// Get all Jobs
	g.GET("", func(context echo.Context) error {
		db := context.Get("DB").(*sql.DB)

		jobs, err := dao.GetJobs(db)
		if err != nil {
			return context.String(500, "Failed to get job, :" +  err.Error())
		}

		return context.JSON(200, jobs)
	})

	// Get a Job
	g.GET("/:JobId", func(context echo.Context) error {
		db := context.Get("DB").(*sql.DB)
		JobId:= context.Param("JobId")

		job, err := dao.GetJob(JobId, db)
		if err != nil {
			return context.String(500, "Failed to get job, " + err.Error())
		}

		return context.JSON(200, job)
	})

	// Get a Job's Logs
	g.GET("/:JobId/logs", func(context echo.Context) error {
		db := context.Get("DB").(*sql.DB)
		JobId:= context.Param("JobId")
		logs, err := dao.GetJobLogs(JobId, db)
		if err != nil {
			return context.String(500, "Failed to get the job's logs, " + err.Error())
		}
		return context.JSON(200, logs)
	})

	// Add new Job
	g.POST("/add", func(context echo.Context) error {
		var job dao.Job
		if err := context.Bind(&job); err != nil {
			return context.String(500, "Could not parse body as job type")
		}

		t, err := time.Parse(time.RFC3339, context.FormValue("CreatedAt"))
		if err != nil {
			return context.String(500, "Could not parse CreatedAt as Time")
		}
		job.CreatedAt = t

		db := context.Get("DB").(*sql.DB)
		err = dao.PostJob(job, db)
		if err != nil {
			return context.String(500, "Could not add Job to DB, " +  err.Error())
		}

		scheduler.Notify()

		return context.String(200, "Job added to DB")
	})

	// Update Job
	g.PUT("/update", func(context echo.Context) error {
		var job dao.Job
		if err := context.Bind(&job); err != nil {
			return context.String(500, "Could not parse body as job type")
		}

		t, err := time.Parse(time.RFC3339, context.FormValue("CreatedAt"))
		if err != nil {
			return context.String(500, "Could not parse CreatedAt as Time")
		}
		job.CreatedAt = t

		db := context.Get("DB").(*sql.DB)
		err = dao.PutJob(job, db)
		if err != nil {
			return context.String(500, "Could not update Job, " + err.Error())
		}

		scheduler.Notify()

		return context.JSON(200, job)
	})

	// Delete Job
	g.DELETE("/delete/:jobId", func(context echo.Context) error {
		jobId := context.Param("jobId")
		if jobId == "" {
			return context.String(500, "Please include jobId in body")
		}

		db := context.Get("DB").(*sql.DB)
		err := dao.DeleteJob(jobId, db)
		if err != nil {
			context.String(500, "Could not delete Job, " + err.Error())
		}

		scheduler.Notify()

		return context.String(500, "Job deleted")
	})
}
