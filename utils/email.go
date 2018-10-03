package utils

import (
	"bytes"
	"fmt"
	"github.com/websentry/websentry/config"
	"gopkg.in/gomail.v2"
	"html/template"
	"strings"
	"time"
	)

const (
	chBuffer = 100
)

var ch chan *gomail.Message
var c config.VerificationEmail

func init() {
	ch = make(chan *gomail.Message, chBuffer)
	c = config.GetVerificationEmailConfig()

	go func() {
		d := gomail.NewDialer(c.Server, c.Port, c.Email, c.Password)

		var sc gomail.SendCloser
		var err error

		open := false

		for {
			select {
			case m, ok := <-ch:
				if !ok {
					return
				}

				if !open {
					if sc, err = d.Dial(); err != nil {
						panic(err)
					}
					open = true
				}

				if err := gomail.Send(sc, m); err != nil {
					// TODO: log
				}

				fmt.Println("[INFO]: Email Sent Successfully To: " + strings.Join(m.GetHeader("To"), ","))

			case <-time.After(time.Minute):
				if open {
					if err = sc.Close(); err != nil {
						panic(err)
					}
					open = false
				}
			}
		}
	}()
}

// SendVerificationEmail sends verification email to new user, it
// takes an email address and the verification code
func SendVerificationEmail(e, vc string) {

	// subject
	var s string
	s =  "Verify Your Account on WebSentry"

	// apply email templates
	b := new(bytes.Buffer)

	t, err := template.ParseFiles("templates/emails/baseEmail.html", "templates/emails/verificationEmail.html")
	if err != nil {
		panic(err)
	}

	if err = t.ExecuteTemplate(b, "base", map[string]string{"verificationCode": vc}); err != nil {
		panic(err)
	}

	bs := b.String()
	SendEmail(e, s, &bs)
}

// sendEmail takes an email address, a subject and a pointer of the body message
func SendEmail(e, s string, b *string) {

	if !config.IsReleaseMode() {
		s = s + " [dev]"
	}

	m := gomail.NewMessage()
	m.SetHeader("From", c.Email)
	m.SetHeader("To", e)
	m.SetHeader("Subject", s)
	m.SetHeader("MIME-version", "1.0")
	m.SetBody("text/html", *b)

	go func() {
		ch <- m
	}()
}
