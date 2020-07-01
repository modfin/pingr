package notifications


import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/jordan-wright/email"
	"github.com/sirupsen/logrus"
	"net/smtp"
	"net/textproto"
	"pingr"
	"pingr/internal/dao"
)

var (
	jobName string = "MFN.se http tester"
	pass string = "2KPy8WJr2uLC46g"
	from string = "pingrman@gmail.com"
	html string = `
			<h2>Pingr - Test error</h2>
			<p>An <strong>error</strong> has occurred in one of your tests.</p>
			<table cellspacing="10"><tbody>
				<tr><td>Test name:  </td>   <td>%s</td></tr>
				<tr><td>Error message:  </td><td>%s</td></tr>
			</tbody></table>
			<h4>Recent logs:</h4>
			<p>%s</p>
	`

)

func SendEmail(job pingr.BaseJob, jobErr error, db *sqlx.DB) {
	logs, _ := dao.GetJobLogsLimited(job.JobId, 10, db)

	t := "agaton.sjoberg@modularfinance.se"
	e := &email.Email {
		To: []string{t},
		From: fmt.Sprintf("Pingr Lad <%s>", from),
		Subject: fmt.Sprintf("Pingr Error: %s", jobName),
		HTML: []byte(fmt.Sprintf(html, jobName, jobErr, logsHTML(logs))),
		Headers: textproto.MIMEHeader{},
	}
	a := smtp.PlainAuth("", from, pass, "smtp.gmail.com")
	err := e.Send("smtp.gmail.com:587", a)
	logrus.Info(err)
}

func logsHTML(logs []pingr.Log) string {
	html := `<table cellspacing="10"><tbody>`
	for _, log := range logs {
		html += fmt.Sprintf(
			"<tr><td>%d</td><td>%s</td><td>%s</td></tr>",
			log.Status, log.Message, log.CreatedAt.Format("2006-01-02T15:04:05"))
	}
	html += "</tbody></table>"
	return html
}