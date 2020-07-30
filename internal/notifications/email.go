package notifications

import (
	"fmt"
	"github.com/gosimple/slug"
	"github.com/jmoiron/sqlx"
	"github.com/jordan-wright/email"
	"github.com/matcornic/hermes/v2"
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
	h = hermes.Hermes{
		Product: hermes.Product{
			Name: "Pingr",
			Link: config.Get().BaseUrl, // TODO: find url?
			Logo: "https://storage.googleapis.com/gopherizeme.appspot.com/gophers/62d0c3d5f52dbc9c803ea7bfaa2829d75a2f8fa2.png",

			Copyright: "https://github.com/itsy-sh/pingr",
		},
	}
)

func SendEmail(receivers []string, test pingr.BaseTest, testErr error, db *sqlx.DB) error {
	logrus.Info("sending email")

	for i := range receivers {
		// add '+test-name' to receivers
		atIndex := strings.Index(receivers[i], "@")
		receivers[i] = receivers[i][:atIndex] + "+" + slug.Make(test.TestName) + receivers[i][atIndex:]
	}

	e := &email.Email{
		To:      receivers,
		From:    fmt.Sprintf("Pingr <%s>", config.Get().SMTPUsername),
		Headers: textproto.MIMEHeader{},
	}

	body, err := getEmailBody(test, testErr, db)
	if err != nil {
		return err
	}

	if testErr != nil {
		e.Subject = fmt.Sprintf("Error: %s", test.TestName)
	} else {
		e.Subject = fmt.Sprintf("Successful: %s", test.TestName)
	}
	e.HTML = body

	a := smtp.PlainAuth("", config.Get().SMTPUsername, config.Get().SMTPPassword, config.Get().SMTPHost)
	return e.Send(fmt.Sprintf("%s:%d", config.Get().SMTPHost, config.Get().SMTPPort), a)
}

func getEmailBody(test pingr.BaseTest, testErr error, db *sqlx.DB) ([]byte, error) {
	logs, err := dao.GetTestLogsLimited(test.TestId, 10, db)
	if err != nil {
		return nil, err
	}

	var table [][]hermes.Entry
	for _, log := range logs {
		var row []hermes.Entry
		row = append(row, hermes.Entry{Key: "Created at", Value: log.CreatedAt.Local().Format("2006-01-02T15:04:05")})
		row = append(row, hermes.Entry{Key: "Status", Value: log.StatusName})
		row = append(row, hermes.Entry{Key: "Error message", Value: log.Message})
		row = append(row, hermes.Entry{Key: "Response time", Value: log.ResponseTime.Round(time.Millisecond).String()})
		table = append(table, row)
	}

	hermesTable := hermes.Table{
		Data: table,
		Columns: hermes.Columns{
			// Custom style for each rows
			CustomWidth: map[string]string{
				"Created at":    "20%",
				"Status":        "10%",
				"Error message": "60%",
				"Response time": "10%",
			},
		},
	}

	var body hermes.Email
	if testErr != nil {
		body = hermes.Email{
			Body: hermes.Body{
				Title: "Error in one of your tests",
				Intros: []string{
					fmt.Sprintf("The test: %s is throwing an error", test.TestName),
					fmt.Sprintf("Error message: %s", testErr),
				},
				Actions: []hermes.Action{
					{
						Button: hermes.Button{
							Color: "#f45b5b", // Optional action button color
							Text:  "View test",
							Link:  config.Get().BaseUrl + "/tests/" + test.TestId,
						},
					},
				},
				Table:     hermesTable,
				Signature: "Happy troubleshooting",
			},
		}
	} else {
		body = hermes.Email{
			Body: hermes.Body{
				Title: "Test successful again",
				Intros: []string{
					fmt.Sprintf("The test: %s is up and running again", test.TestName),
				},
				Actions: []hermes.Action{
					{
						Button: hermes.Button{
							Color:     "#90ed7d", // Optional action button color
							Text:      "View test",
							TextColor: "#000000",
							Link:      config.Get().BaseUrl + "/tests/" + test.TestId,
						},
					},
				},
				Table:     hermesTable,
				Signature: "Happy troubleshooting",
			},
		}
	}

	bodyString, err := h.GenerateHTML(body)
	if err != nil {
		return nil, err
	}
	return []byte(bodyString), nil
}
