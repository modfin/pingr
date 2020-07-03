package notifications


import (
	"fmt"
	"github.com/gosimple/slug"
	"github.com/jmoiron/sqlx"
	"github.com/jordan-wright/email"
	"github.com/sirupsen/logrus"
	"net/smtp"
	"net/textproto"
	"pingr"
	"pingr/internal/dao"
	"strings"
)

var (
	pass string = "2KPy8WJr2uLC46g"
	from string = "pingrman@gmail.com"
	htmlError string = `
			<h2>Pingr - Test error</h2>
			<p>An <strong>error</strong> has occurred in one of your tests.</p>
			<table cellspacing="10"><tbody>
				<tr><td>Test name:  </td>   <td>%s</td></tr>
				<tr><td>Error message:  </td><td>%s</td></tr>
			</tbody></table>
			<h4>Recent logs:</h4>
			<p>%s</p>
	`
	htmlSuccess string = `
		<h2>Pingr - Test successful again</h2>
			<p>Test <strong>successful</strong>.</p>
			<table cellspacing="10"><tbody>
				<tr><td>Test name:  </td>   <td>%s</td></tr>
			</tbody></table>
			<h4>Recent logs:</h4>
			<p>%s</p>
	`

)

func SendEmail(receivers []string, job pingr.BaseJob, jobErr error, db *sqlx.DB) error {
	logrus.Info("sending email")
	logs, err := dao.GetJobLogsLimited(job.JobId, 10, db)
	if err != nil {
		return err
	}

	for i := range receivers {
		// add '+job-name' to receivers
		atIndex := strings.Index(receivers[i], "@")
		receivers[i]=receivers[i][:atIndex]+"+"+slug.Make(job.JobName)+receivers[i][atIndex:]
	}

	e := &email.Email {
		To:      receivers,
		From:    fmt.Sprintf("Pingr Lad <%s>", from),
		Headers: textproto.MIMEHeader{},
	}

	if jobErr != nil {
		e.Subject = fmt.Sprintf("Pingr - Error: %s", job.JobName)
		e.HTML = []byte(fmt.Sprintf(htmlError, job.JobName, jobErr, logsHTML(logs)))
	} else {
		e.Subject = fmt.Sprintf("Pingr - Test Succesful: %s", job.JobName)
		e.HTML = []byte(fmt.Sprintf(htmlSuccess, job.JobName, logsHTML(logs)))
	}

	a := smtp.PlainAuth("", from, pass, "smtp.gmail.com")
	err = e.Send("smtp.gmail.com:587", a)
	if err != nil {
		return err
	}

	return nil
}

func logsHTML(logs []dao.FullLog) string {
	html := `<table cellspacing="10"><tbody>
				<tr>
					<td><strong>Created</strong></td>
					<td><strong>Status</strong></td>
					<td><strong>Message</strong></td>
					<td><strong>Response time</strong></td>
				</tr>`
	for _, log := range logs {
		html += fmt.Sprintf(
			"<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
			log.CreatedAt.Format("2006-01-02T15:04:05"), log.StatusName,
			log.Message, log.ResponseTime)
	}
	html += "</tbody></table>"
	return html
}