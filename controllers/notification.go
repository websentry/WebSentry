package controllers

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/models"
	"github.com/websentry/websentry/utils"
	"gopkg.in/mgo.v2"
	"html/template"
	"net/http"
	"time"
)

func notificationToggle(db *mgo.Database, sentryId primitive.ObjectID, lasttime time.Time, old string, new string) error {
	nid, err := models.GetSentryNotification(db, sentryId)
	name, _ := models.GetSentryName(db, sentryId)
	if err != nil {
		return err
	}
	n := models.GetNotification(nid)

	// TODO: url

	data := map[string]string{
		"name": name,
		"beforeTime": lasttime.Format("2006-01-02 15:04"),
		"currentTime": time.Now().Format("2006-01-02 15:04"),
		"beforeImage": config.GetBaseUrl() + "v1/common/get_history_image?filename="+old,
		"afterImage": config.GetBaseUrl() + "v1/common/get_history_image?filename="+new,
	}

	title := name + ": change detected"

	if n.Type == "serverchan" {

		b := bytes.Buffer{}
		t, _ := template.ParseFiles("templates/notifications/serverchan.md")
		t.Execute(&b, data)
		msg := b.String()

		title = template.URLQueryEscaper(title)
		msg = template.URLQueryEscaper(msg)
		http.Get(fmt.Sprintf("https://sc.ftqq.com/%s.send?text=%s&desp=%s", n.Setting["sckey"], title, msg))

		return nil
	} else if n.Type == "email" {


		// apply email templates
		b := bytes.Buffer{}

		t, err := template.ParseFiles("templates/emails/baseEmail.html", "templates/notifications/email.html")
		if err != nil {
			panic(err)
		}

		if err = t.ExecuteTemplate(&b, "base", data); err != nil {
			panic(err)
		}

		bs := b.String()
		utils.SendEmail(n.Setting["email"].(string), title, &bs)

	}

	return nil
}


func NotificationList(c *gin.Context) {
	results, err := models.NotificationList(c.MustGet("mongo").(*mgo.Database), c.MustGet("userId").(primitive.ObjectID))
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