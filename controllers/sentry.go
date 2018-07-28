package controllers

import (
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/websentry/websentry/middlewares"
	"github.com/websentry/websentry/models"
	"github.com/disintegration/imaging"
	"bytes"
	"fmt"
	"strconv"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
	"github.com/websentry/websentry/utils"
)

func init() {
	go sentryTaskScheduler()
}

// [url] the url of the page that needs screenshot
func SentryRequestFullScreenshot(c *gin.Context) {
	u, err := url.ParseRequestURI(c.Query("url"))
	if err != nil || !(strings.EqualFold(u.Scheme, "http") || strings.EqualFold(u.Scheme, "https")) {
		c.JSON(200, gin.H{
			"code": -2,
			"msg":  "Wrong parameter",
		})
		return
	}

	// TODO: handle extreme long page

	task := gin.H{
		"url":      u.String(),
		"timeout":  40000,
		"fullPage": true,
		"viewport": gin.H{
			"width":    900,
			"isMobile": false,
		},
		"output": gin.H{
			"type":        "jpg",
			"progressive": true,
			"quality":     20,
		},
	}

	id := addFullScreenshotTask(task, c.MustGet("userId").(bson.ObjectId))

	c.JSON(200, gin.H{
		"code":   0,
		"msg":    "OK",
		"taskId": id,
	})
}

func SentryWaitFullScreenshot(c *gin.Context) {
	waitFullScreenshot(c)
}

func SentryGetFullScreenshot(c *gin.Context) {
	getFullScreenshot(c)
}

func SentryCreate(c *gin.Context) {
	u, err := url.ParseRequestURI(c.Query("url"))
	db := c.MustGet("mongo").(*mgo.Database)

	if err != nil || !(strings.EqualFold(u.Scheme, "http") || strings.EqualFold(u.Scheme, "https")) {
		c.JSON(200, gin.H{
			"code": -2,
			"msg":  "Wrong parameter",
		})
		return
	}

	if !bson.IsObjectIdHex(c.Query("notification")) {
		c.JSON(200, gin.H{
			"code": -2,
			"msg":  "Wrong parameter",
		})
		return
	}

	notification := bson.ObjectIdHex(c.Query("notification"))

	if !models.NotificationCheckOwner(db, notification, c.MustGet("userId").(bson.ObjectId)) {
		c.JSON(200, gin.H{
			"code": -2,
			"msg":  "Wrong parameter",
		})
		return
	}

	x, _ := strconv.ParseInt(c.Query("x"), 10, 32)
	y, _ := strconv.ParseInt(c.Query("y"), 10, 32)
	width, _ := strconv.ParseInt(c.Query("width"), 10, 32)
	height, _ := strconv.ParseInt(c.Query("height"), 10, 32)

	if !(x >= 0 && y >= 0 && width > 0 && height > 0) {
		c.JSON(200, gin.H{
			"code": -2,
			"msg":  "Wrong parameter",
		})
		return
	}

	if width*height > 500*500 {
		c.JSON(200, gin.H{
			"code": -11,
			"msg":  "Area too large",
		})
		return
	}

	s := &models.Sentry{}
	s.Id = bson.NewObjectId()
	s.Name = c.Query("name")
	s.User = c.MustGet("userId").(bson.ObjectId)
	s.Notification = notification
	s.CreateTime = time.Now()
	s.NextCheckTime = time.Now()
	s.Interval = 4 * 60 // 4 hours
	s.CheckCount = 0
	s.NotifyCount = -1 // will be add 1 at the first check
	s.Image.File = "placeholder"
	s.Task = gin.H{
		"url":      u.String(),
		"timeout":  40000,
		"fullPage": false,
		"clip" : gin.H{
			"x" : int(x),
			"width" : int(width),
			"height" : int(height),
			"y" : int(y),
		},
		"viewport": gin.H{
			"width":    900,
			"isMobile": false,
		},
		"output": gin.H{
			"type": "png",
		},
	}

	err = db.C("Sentries").Insert(s)
	if err!=nil {
		panic(err)
	}

	c.JSON(200, gin.H{
		"code":   0,
		"msg":    "OK",
		"sentryId": s.Id.Hex(),
	})

}

func sentryTaskScheduler() {
	for {
		time.Sleep(2 * time.Minute)

		s := middlewares.GetDBSession()
		db := middlewares.SessionToDB(s)
		for {
			sentry := models.GetUncheckedSentry(db)
			if sentry == nil {
				break
			}
			// add task
			addSentryTask(sentry)
		}
		s.Clone()
	}
}

func compareSentryTaskImage(tid int32, ti *taskInfo) (error) {
	defer func() {
		// clean up
		taskq.infoMux.Lock()
		delete(taskq.info, tid)
		taskq.infoMux.Unlock()
	}()

	b, err := imaging.Decode(bytes.NewReader(ti.image))
	if err!=nil {
		return err
	}

	// first time
	if ti.baseImage.File == "placeholder" {

		imageFilename := utils.ImageSave(b)

		s := middlewares.GetDBSession()
		db := middlewares.SessionToDB(s)
		err := models.UpdateSentryAfterCheck(db, ti.sentryId, true, imageFilename)
		s.Close()

		if err != nil {
			utils.ImageDelete(imageFilename, false)
		}
		return err
	}


	a, err := imaging.Open(utils.ImageGetFullPath(ti.baseImage.File, false))


	if err!=nil {
		// TODO: error handling
		fmt.Println("image error")
		return err
	}

	v, _ := utils.ImageCompare(a,b)
	changed := v < 0.9999
	newImage := ""
	if changed {
		// changed
		// save new image
		newImage = utils.ImageSave(b)
	}

	s := middlewares.GetDBSession()
	db := middlewares.SessionToDB(s)
	err = models.UpdateSentryAfterCheck(db, ti.sentryId, changed, newImage)
	s.Close()


	if changed {
		if err==nil {
			// success

			// notification
			notificationToggle(ti.sentryId, ti.baseImage.File, newImage)

			// delete old file (keep thumb)
			utils.ImageDelete(ti.baseImage.File, true)
		} else {
			// delete new file (delete all)
			utils.ImageDelete(newImage, false)
		}
	}

	return err
}
