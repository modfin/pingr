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
	"pingr/internal/config"
	"pingr/internal/dao"
	"strings"
	"time"
)

var (
	htmlError = `
			<h2>Pingr - Test error</h2>
			<p>An <strong>error</strong> has occurred in one of your tests.</p>
			<table cellspacing="10"><tbody>
				<tr><td>Test name:  </td>   <td>%s</td></tr>
				<tr><td>Error message:  </td><td>%s</td></tr>
			</tbody></table>
			<h4>Recent logs:</h4>
			<p>%s</p>
	`
	htmlSuccess = `
		<h2>Pingr - Test successful again</h2>
			<p>Test <strong>successful</strong>.</p>
			<table cellspacing="10"><tbody>
				<tr><td>Test name:  </td>   <td>%s</td></tr>
			</tbody></table>
			<h4>Recent logs:</h4>
			<p>%s</p>
	`

)

func SendEmail(receivers []string, test pingr.BaseTest, testErr error, db *sqlx.DB) error {
	logrus.Info("sending email")

	logs, err := dao.GetTestLogsLimited(test.TestId, 10, db)
	if err != nil {
		return err
	}

	for i := range receivers {
		// add '+test-name' to receivers
		atIndex := strings.Index(receivers[i], "@")
		receivers[i]=receivers[i][:atIndex]+"+"+slug.Make(test.TestName)+receivers[i][atIndex:]
	}

	e := &email.Email {
		To:      receivers,
		From:    fmt.Sprintf("Pingr Lad <%s>", config.Get().SMTPUsername),
		Headers: textproto.MIMEHeader{},
	}

	if testErr != nil {
		e.Subject = fmt.Sprintf("Pingr - Error: %s", test.TestName)
		e.HTML = []byte(fmt.Sprintf(htmlError, test.TestName, testErr, logsHTML(logs)))
	} else {
		e.Subject = fmt.Sprintf("Pingr - Test Succesful: %s", test.TestName)
		e.HTML = []byte(fmt.Sprintf(htmlSuccess, test.TestName, logsHTML(logs)))
	}

	a := smtp.PlainAuth("", config.Get().SMTPUsername, config.Get().SMTPPassword, config.Get().SMTPHost)
	err = e.Send(fmt.Sprintf("%s:%d", config.Get().SMTPHost, config.Get().SMTPPort), a)
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
			log.Message, log.ResponseTime.Round(time.Millisecond).String())
	}
	html += "</tbody></table>"
	return html
}