package mail

import (
	"bytes"
	"fmt"
	"html/template"
	"mime/quotedprintable"
	"net/smtp"
)

// Mail is a generic function to send email
func Mail(
	fromEmail, toEmail, password, host, port, subject, bodyMessage string,
) error {
	addr := fmt.Sprintf("%s:%s", host, port)
	// PlainAuth will only send the credentials if the connection is using TLS
	// or is connected to localhost.
	auth := smtp.PlainAuth("", fromEmail, password, host)

	header := make(map[string]string)

	header["From"] = fromEmail
	header["To"] = toEmail
	header["Subject"] = subject

	header["MIME-Version"] = "1.0"
	header["Content-Type"] = fmt.Sprintf("%s; charset=\"utf-8\"", "text/html")
	header["Content-Transfer-Encoding"] = "quoted-printable"
	header["Content-Disposition"] = "inline"

	headerMessage := ""
	for key, value := range header {
		headerMessage += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	finalMessage := headerMessage + "\r\n" + bodyMessage
	sendErr := smtp.SendMail(
		addr,
		auth,
		fromEmail,
		[]string{toEmail},
		[]byte(finalMessage),
	)
	if sendErr != nil {
		return sendErr
	}

	return nil
}

// NewSummaryMailBody is specific to the PullTestResults
func NewSummaryMailBody(summary *MissingReport) (string, error) {
	const htmlTmpl = `
  <hr style=
  "border: 0;
height: 1px;
background-image: linear-gradient(to right, rgba(0, 0, 0, 0), rgba(0, 0, 0, 0.75), rgba(0, 0, 0, 0));
" />


  <h3>&#129335;District {{.DistrictName}} CleverID {{.DistrictCleverID}} was missing these Clever IDs:</h3>
  <h4>Student Clever IDs</h4>
  <ul>
    <li style="list-style: none">{{range .MissingStudentCleverIDs}}</li>

    <li>{{.}}</li>

    <li style="list-style: none">{{end}}</li>
  </ul>

  <hr style=
  "border: 0;
height: 1px;
background-image: linear-gradient(to right, rgba(0, 0, 0, 0), rgba(0, 0, 0, 0.75), rgba(0, 0, 0, 0));
" />

  <h4>Teacher Clever IDs</h4>
  <ul>
    <li style="list-style: none">{{range .MissingTeacherCleverIDs}}</li>

    <li>{{.}}</li>

    <li style="list-style: none">{{end}}</li>
  </ul>

  <hr style=
  "border: 0;
height: 1px;
background-image: linear-gradient(to right, rgba(0, 0, 0, 0), rgba(0, 0, 0, 0.75), rgba(0, 0, 0, 0));
" />

  <h4>School Clever IDs</h4>
  <ul>
    <li style="list-style: none">{{range .MissingSchoolCleverIDs}}</li>

    <li>{{.}}</li>

    <li style="list-style: none">{{end}}</li>
  </ul>

  <hr style=
  "border: 0;
height: 1px;
background-image: linear-gradient(to right, rgba(0, 0, 0, 0), rgba(0, 0, 0, 0.75), rgba(0, 0, 0, 0));
" />
`

	t, err := template.New("pullRosterSummary").Parse(htmlTmpl)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, summary)
	if err != nil {
		return "", err
	}
	body := tpl.String()

	var bodyMessage bytes.Buffer
	temp := quotedprintable.NewWriter(&bodyMessage)
	defer temp.Close()
	_, writeErr := temp.Write([]byte(body))
	if writeErr != nil {
		return "", writeErr
	}

	return bodyMessage.String(), nil
}

type MissingReport struct {
	DistrictName            string
	DistrictCleverID        string
	MissingStudentCleverIDs []string
	MissingTeacherCleverIDs []string
	MissingSchoolCleverIDs  []string
}
