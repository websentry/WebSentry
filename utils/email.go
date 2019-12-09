package utils

import (
	"bytes"
	"html/template"
	"log"
	"strings"
	"time"

	"gopkg.in/mail.v2"

	"github.com/websentry/websentry/config"
)

const (
	chBuffer int           = 1000
	timeOut  time.Duration = time.Minute
)

var ch chan *mail.Message
var c config.VerificationEmail

func init() {
	ch = make(chan *mail.Message, chBuffer)
	c = config.GetVerificationEmailConfig()

	d := mail.NewDialer(c.Server, c.Port, c.Email, c.Password)
	d.Timeout = 0
	runDaemon(d)
}

func runDaemon(d *mail.Dialer) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[INFO]: Email daemon recovered %v", r)
				runDaemon(d)
			}
		}()

		var sc mail.SendCloser
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
						log.Panicln("[Error] Failed to dial SMTP server: ", err)
					}
					open = true
				}

				if err := mail.Send(sc, m); err != nil {
					log.Println("[Warning]: Failed to send email To: " + strings.Join(m.GetHeader("To"), ","))
				}

				log.Println("[INFO]: Email Sent Successfully To: " + strings.Join(m.GetHeader("To"), ","))

			case <-time.After(timeOut):
				if open {
					if err = sc.Close(); err != nil {
						log.Panicln("[ERROR] Failed to close connection", err)
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
	s = "Verify Your Account on WebSentry"

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

// SendEmail takes an email address, a subject and a pointer of the body message
func SendEmail(e, s string, b *string) {

	if !config.IsReleaseMode() {
		s = s + " [dev]"
	}

	m := mail.NewMessage()
	m.SetHeader("From", c.Email)
	m.SetHeader("To", e)
	m.SetHeader("Subject", s)
	m.SetHeader("MIME-version", "1.0")
	m.SetBody("text/html", *b)

	go func() {
		ch <- m
	}()
}
