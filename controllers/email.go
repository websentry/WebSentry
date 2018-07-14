package controllers

import (
	"fmt"
	"github.com/websentry/websentry/config"
	"gopkg.in/gomail.v2"
	// "crypto/tls"
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
		//d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

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
					// TODO: handle sending email failure
				}

				fmt.Println("! [INFO] Email sent")
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
	m.SetHeader("Subject", "Verify Your Account on Websentry")
	m.SetBody("text/html", "<b>"+vc+"</b>")

	go func() {
		ch <- m
	}()
}
