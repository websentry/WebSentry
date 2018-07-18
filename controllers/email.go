package controllers

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

// VerificationEmailInit initializes email daemon
func VerificationEmailInit() {
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

				fmt.Println("[INFO]: Verification Email Sent Successfully To: " + strings.Join(m.GetHeader("To"), ","))

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

// SendVerificationEmail sends email to new user, it
// takes an email address and the verification code
func SendVerificationEmail(e, vc string) {
	m := gomail.NewMessage()
	m.SetHeader("From", c.Email)
	m.SetHeader("To", e)
	m.SetHeader("Subject", "Verify Your Account on WebSentry")
	m.SetHeader("MIME-version", "1.0")
	m.SetBody("text/html", generateVerificationEmailHTML(vc))

	go func() {
		ch <- m
	}()
}

func generateVerificationEmailHTML(vc string) string {
	b := new(bytes.Buffer)

	// gp := os.Getenv("GOPATH")
	// htmlPath := path.Join(gp, "src/github.com/websentry/websentry/templates/verificationEmail.html")

	t, err := template.ParseFiles("templates/verificationEmail.html")

	if err != nil {
		panic(err)
	}

	if err = t.Execute(b, map[string]string{"verificationCode": vc}); err != nil {
		panic(err)
	}

	return b.String()
}
