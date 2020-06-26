package jobs

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"pingr/internal/dao"
	"pingr/internal/scheduler"
	"strconv"
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
	g.GET("/:jobId", func(context echo.Context) error {
		db := context.Get("DB").(*sql.DB)
		jobIdString:= context.Param("jobId")

		jobId, err := strconv.ParseUint(jobIdString, 10, 64)
		if err != nil {
			return context.String(500, "Could not parse JobId as int")
		}

		job, err := dao.GetJob(jobId, db)
		if err != nil {
			return context.String(500, "Failed to get job, " + err.Error())
		}

		return context.JSON(200, job)
	})

	// Get a Job's Logs
	g.GET("/:jobId/logs", func(context echo.Context) error {
		db := context.Get("DB").(*sql.DB)
		jobIdString:= context.Param("jobId")

		jobId, err := strconv.ParseUint(jobIdString, 10, 64)
		if err != nil {
			return context.String(500, "Could not parse JobId as int")
		}

		logs, err := dao.GetJobLogs(jobId, db)
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
		if !job.Validate(false) {
			return context.String(500, "Input invalid")
		}
		job.CreatedAt = time.Now()

		db := context.Get("DB").(*sql.DB)
		job, err := dao.PostJob(job, db)
		if err != nil {
			return context.String(500, "Could not add Job to DB, " +  err.Error())
		}

		scheduler.NotifyNewJob(job)

		return context.String(200, "Job added to DB")
	})

	// Update Job
	g.PUT("/update", func(context echo.Context) error {
		var job dao.Job
		if err := context.Bind(&job); err != nil {
			return context.String(500, "Could not parse body as job type")
		}
		if !job.Validate(true) {
			return context.String(500, "Input invalid")
		}
		job.CreatedAt = time.Now()
		db := context.Get("DB").(*sql.DB)
		_, err := dao.GetJob(job.JobId, db)
		if err != nil {
			return context.String(500, "Not a valid/active jobId, " + err.Error())
		}

		job, err = dao.PutJob(job, db)
		if err != nil {
			return context.String(500, "Could not update Job, " + err.Error())
		}

		scheduler.NotifyNewJob(job)

		return context.JSON(200, job)
	})

	// Delete Job
	g.DELETE("/delete/:jobId", func(context echo.Context) error {
		jobIdString:= context.Param("jobId")

		jobId, err := strconv.ParseUint(jobIdString, 10, 64)
		if err != nil {
			return context.String(500, "Could not parse jobId as int")
		}
		db := context.Get("DB").(*sql.DB)
		_, err = dao.GetJob(jobId, db)
		if err != nil {
			return context.String(500, "Not a valid/active jobId, " + err.Error())
		}

		err = dao.DeleteJob(jobId, db)
		if err != nil {
			return context.String(500, "Could not delete Job, " + err.Error())
		}

		scheduler.NotifyDeletedJob(jobId)

		return context.String(500, "Job deleted")
	})

}
