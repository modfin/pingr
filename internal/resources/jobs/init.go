package jobs

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"pingr/internal/dao"
	"pingr/internal/scheduler"
	"strconv"
	"time"
)

func Init(g *echo.Group) {
	// Get all Jobs
	g.GET("", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)

		jobs, err := dao.GetJobs(db)
		if err != nil {
			return context.String(500, "Failed to get job, :" +  err.Error())
		}

		return context.JSON(200, jobs)
	})

	// Get a Job
	g.GET("/:jobId", func(context echo.Context) error {
		db := context.Get("DB").(*sqlx.DB)
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
		db := context.Get("DB").(*sqlx.DB)
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
		var jobDB dao.Job
		if err := context.Bind(&jobDB); err != nil {
			return context.String(500, "Could not parse body as job type: " + err.Error())
		}

		jobDB.CreatedAt = time.Now()

		pJob, err := jobDB.Parse()
		if err != nil {
			return context.String(500,"Could not parse job data: " + err.Error())
		}
		if !pJob.Validate(false) {
			return context.String(500,"invalid input: Job")
		}

		db := context.Get("DB").(*sqlx.DB)
		err = dao.PostJob(jobDB, db)
		if err != nil {
			return context.String(500, "Could not add Job to DB, " +  err.Error())
		}

		scheduler.NotifyNewJob()

		return context.String(200, "Job added to DB")
	})

	// Update Job
	g.PUT("/update", func(context echo.Context) error {
		var jobDB dao.Job
		if err := context.Bind(&jobDB); err != nil {
			return context.String(500, "Could not parse body as job type")
		}

		jobDB.CreatedAt = time.Now()

		pJob, err := jobDB.Parse()
		if err != nil {
			return err
		}
		if !pJob.Validate(false) {
			return context.String(500,"invalid input: Job")
		}

		db := context.Get("DB").(*sqlx.DB)
		_, err = dao.GetJob(jobDB.JobId, db)
		if err != nil {
			return context.String(500, "Not a valid/active jobId, " + err.Error())
		}

		err = dao.PutJob(jobDB, db)
		if err != nil {
			return context.String(500, "Could not update Job, " + err.Error())
		}

		scheduler.NotifyNewJob()

		return context.JSON(200, pJob)
	})

	// Delete Job
	g.DELETE("/delete/:jobId", func(context echo.Context) error {
		jobIdString:= context.Param("jobId")

		jobId, err := strconv.ParseUint(jobIdString, 10, 64)
		if err != nil {
			return context.String(500, "Could not parse jobId as int")
		}
		db := context.Get("DB").(*sqlx.DB)
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
