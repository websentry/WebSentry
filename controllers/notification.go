package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/models"
	"github.com/websentry/websentry/utils"
)

func toggleNotification(sentryID int64, lasttime time.Time, old string, new string, similarity float32) error {
	nid, err := models.GetSentryNotification(sentryID)
	if err != nil {
		return errors.WithStack(err)
	}
	name, err := models.GetSentryName(sentryID)
	if err != nil {
		return errors.WithStack(err)
	}
	n, err := models.GetNotification(nid)
	if err != nil {
		return errors.WithStack(err)
	}
	var nSetting map[string]interface{}
	err = json.Unmarshal([]byte(n.Setting), &nSetting)
	if err != nil {
		return errors.WithStack(err)
	}

	// TODO: url

	data := map[string]string{
		"name":        name,
		"beforeTime":  lasttime.Format("2006-01-02 15:04"),
		"currentTime": time.Now().Format("2006-01-02 15:04"),
		"beforeImage": config.GetConfig().BackendURL + "v1/common/get_history_image?filename=" + old,
		"afterImage":  config.GetConfig().BackendURL + "v1/common/get_history_image?filename=" + new,
		"similarity":  fmt.Sprintf("%.2f%%", similarity*100),
	}

	title := name + ": change detected"

	if n.Type == "serverchan" {

		b := &bytes.Buffer{}
		t, err := template.ParseFiles("templates/notifications/serverchan.md")
		if err != nil {
			return errors.WithStack(err)
		}
		err = t.Execute(b, data)
		if err != nil {
			return errors.WithStack(err)
		}
		msg := b.String()

		title = template.URLQueryEscaper(title)
		msg = template.URLQueryEscaper(msg)

		//TODO: check response
		_, err = http.Get(fmt.Sprintf("https://sc.ftqq.com/%s.send?text=%s&desp=%s",
			nSetting["sckey"], title, msg))
		if err != nil {
			return errors.WithStack(err)
		}

		return nil

	} else if n.Type == "email" {

		// apply email templates
		b := &bytes.Buffer{}

		t, err := template.ParseFiles("templates/emails/baseEmail.html", "templates/notifications/email.html")
		if err != nil {
			return errors.WithStack(err)
		}

		if err = t.ExecuteTemplate(b, "base", data); err != nil {
			return errors.WithStack(err)
		}

		bs := b.String()
		utils.SendEmail(nSetting["email"].(string), title, &bs)

	}

	return nil
}

type NotificationListItemJSON struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"createAt"`
}

func NotificationList(c *gin.Context) {
	results, err := models.NotificationList(c.MustGet("userId").(int64))
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}
	notifications := make([]NotificationListItemJSON, len(results))
	for i := range results {
		notifications[i].ID = strconv.FormatInt(results[i].ID, 16)
		notifications[i].Name = results[i].Name
		notifications[i].Type = results[i].Type
		notifications[i].CreatedAt = results[i].CreatedAt
		setting := gin.H{}
		err := json.Unmarshal([]byte(results[i].Setting), &setting)
		if err != nil {
			InternalErrorResponse(c, err)
			return
		}
		switch results[i].Type {
		case "serverchan":
			notifications[i].Detail = setting["sckey"].(string)
		case "email":
			notifications[i].Detail = setting["email"].(string)
		}
	}
	JSONResponse(c, CodeOK, "", gin.H{
		"notifications": notifications,
	})
}

func NotificationAddServerChan(c *gin.Context) {
	name := c.Query("name")
	user := c.MustGet("userId").(int64)
	sckey := c.Query("sckey")

	id, err := models.NotificationAddServerChan(name, user, sckey)

	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	JSONResponse(c, CodeOK, "", gin.H{
		"notificationId": id,
	})
}
