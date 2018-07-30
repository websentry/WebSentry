package controllers

import (
	"time"
	"bytes"
	"html/template"

	"gopkg.in/mgo.v2/bson"
	"github.com/gin-gonic/gin"
	"github.com/websentry/websentry/models"
	"gopkg.in/mgo.v2"
	"net/http"
	"fmt"
	"github.com/websentry/websentry/config"
)

func notificationToggle(db *mgo.Database, sentryId bson.ObjectId, lasttime time.Time, old string, new string) error {
	nid, err := models.GetSentryNotification(db, sentryId)
	name, _ := models.GetSentryName(db, sentryId)
	if err!=nil {
		return err
	}
	n := models.GetNotification(db, nid)

	if n.Type == "serverchan" {

		// TODO: url

		data := map[string]string{
			"name": name,
			"beforeTime": lasttime.Format("2006-01-02 15:04"),
			"currentTime": time.Now().Format("2006-01-02 15:04"),
			"beforeImage": config.GetBaseUrl() + "v1/common/get_history_image?filename="+old,
			"afterImage": config.GetBaseUrl() + "v1/common/get_history_image?filename="+new,
		}

		title := name + " has changed"

		b := bytes.Buffer{}
		t, _ := template.ParseFiles("templates/notifications/serverchan.md")
		t.Execute(&b, data)
		msg := b.String()

		title = template.URLQueryEscaper(title)
		msg = template.URLQueryEscaper(msg)
		http.Get(fmt.Sprintf("https://sc.ftqq.com/%s.send?text=%s&desp=%s", n.Setting["sckey"], title, msg))

		return nil
	}

	return nil
}


func NotificationList(c *gin.Context) {
	results, err := models.NotificationList(c.MustGet("mongo").(*mgo.Database), c.MustGet("userId").(bson.ObjectId))
	if err!=nil {
		panic(err)
	}

	c.JSON(200, gin.H{
		"code":   0,
		"msg":    "OK",
		"notifications": results,
	})
}

func NotificationAddServerChan(c *gin.Context) {
	name := c.Query("name")
	user := c.MustGet("userId").(bson.ObjectId)
	sckey := c.Query("sckey")

	id, err := models.NotificationAddServerChan(c.MustGet("mongo").(*mgo.Database), name, user, sckey)

	if err!=nil {
		panic(err)
	}

	c.JSON(200, gin.H{
		"code":   0,
		"msg":    "OK",
		"notificationId": id,
	})

}