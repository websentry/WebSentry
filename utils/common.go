package utils

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/websentry/websentry/config"
	"gopkg.in/mail.v2"
)

func Init() error {
	// email
	ch = make(chan *mail.Message, chBuffer)
	c = config.GetConfig().VerificationEmail

	d := mail.NewDialer(c.Server, c.Port, c.Email, c.Password)
	d.Timeout = 0
	runDaemon(d)

	// image
	imageBasePath = path.Join(config.GetConfig().FileStoragePath, "sentry", "image", "orig")
	imageThumbBasePath = path.Join(config.GetConfig().FileStoragePath, "sentry", "image", "thumb")

	err := os.MkdirAll(imageBasePath, os.ModePerm)
	if err != nil {
		return errors.WithStack(err)
	}
	err = os.MkdirAll(imageThumbBasePath, os.ModePerm)
	if err != nil {
		return errors.WithStack(err)
	}

	// token
	secreteKey = []byte(config.GetConfig().TokenSecretKey)

	return nil
}
