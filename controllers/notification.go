package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/models"
	"github.com/websentry/websentry/utils"
)

func toggleNotification(sentryId primitive.ObjectID, lasttime time.Time, old string, new string, similarity float32) error {
	nid, err := models.GetSentryNotification(sentryId)
	name, _ := models.GetSentryName(sentryId)
	if err != nil {
		return err
	}
	n, err := models.GetNotification(nid)
	if err != nil {
		return err
	}
	// TODO: url

	data := map[string]string{
		"name": name,
		"beforeTime": lasttime.Format("2006-01-02 15:04"),
		"currentTime": time.Now().Format("2006-01-02 15:04"),
		"beforeImage": config.GetBackendUrl() + "v1/common/get_history_image?filename="+old,
		"afterImage": config.GetBackendUrl() + "v1/common/get_history_image?filename="+new,
		"similarity": fmt.Sprintf("%.2f%%", similarity * 100),
	}

	title := name + ": change detected"

	if n.Type == "serverchan" {

		b := &bytes.Buffer{}
		t, err := template.ParseFiles("templates/notifications/serverchan.md")
		if err != nil {
			return err
		}
		err = t.Execute(b, data)
		if err != nil {
			return err
		}
		msg := b.String()

		title = template.URLQueryEscaper(title)
		msg = template.URLQueryEscaper(msg)

		//TODO: check response
		_, err = http.Get(fmt.Sprintf("https://sc.ftqq.com/%s.send?text=%s&desp=%s",
			n.Setting["sckey"], title, msg))
		if err != nil {
			return err
		}

		return nil

	} else if n.Type == "email" {


		// apply email templates
		b := &bytes.Buffer{}

		t, err := template.ParseFiles("templates/emails/baseEmail.html", "templates/notifications/email.html")
		if err != nil {
			return err
		}

		if err = t.ExecuteTemplate(b, "base", data); err != nil {
			return err
		}

		bs := b.String()
		utils.SendEmail(n.Setting["email"].(string), title, &bs)

	}

	return nil
}


func NotificationList(c *gin.Context) {
	results, err := models.NotificationList(c.MustGet("userId").(primitive.ObjectID))
	if err!=nil {
		panic(err)
	}
	JsonResponse(c, CodeOK, "", gin.H{
		"notifications": results,
	})
}

func NotificationAddServerChan(c *gin.Context) {
	name := c.Query("name")
	user := c.MustGet("userId").(primitive.ObjectID)
	sckey := c.Query("sckey")

	id, err := models.NotificationAddServerChan(name, user, sckey)

	if err!=nil {
		panic(err)
	}

	JsonResponse(c, CodeOK, "", gin.H{
		"notificationId": id,
	})
}