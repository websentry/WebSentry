package controllers

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/websentry/websentry/models"
	"github.com/websentry/websentry/utils"
)

// [url] the url of the page that needs screenshot
func SentryRequestFullScreenshot(c *gin.Context) {
	u, err := url.ParseRequestURI(c.Query("url"))
	if err != nil || !(strings.EqualFold(u.Scheme, "http") || strings.EqualFold(u.Scheme, "https")) {
		JSONResponse(c, CodeWrongParam, "Wrong protocol", nil)
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

	id := addFullScreenshotTask(task, c.MustGet("userId").(primitive.ObjectID))

	JSONResponse(c, CodeOK, "", gin.H{
		"taskId": id,
	})
}

func SentryWaitFullScreenshot(c *gin.Context) {
	waitFullScreenshot(c)
}

func SentryList(c *gin.Context) {

	results, err := models.GetUserSentries(c.MustGet("userId").(primitive.ObjectID))
	if err != nil {
		panic(err)
	}

	sentries := make([]struct {
		ID            primitive.ObjectID `json:"id"`
		Name          string             `json:"name"`
		URL           string             `json:"url"`
		LastCheckTime time.Time          `json:"lastCheckTime"`
	}, len(results))

	for i := range results {
		sentries[i].Name = results[i].Name
		sentries[i].ID = results[i].ID
		sentries[i].URL = results[i].Task["url"].(string)
		sentries[i].LastCheckTime = results[i].LastCheckTime
	}

	JSONResponse(c, CodeOK, "", gin.H{
		"sentries": sentries,
	})
}

func SentryInfo(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Query("id"))
	if err != nil {
		JSONResponse(c, CodeWrongParam, "Wrong sentry id", nil)
		return
	}

	s, err := models.GetSentry(id)
	if models.IsErrNoDocument(err) || s.User != c.MustGet("userId").(primitive.ObjectID) {
		JSONResponse(c, CodeNotExist, "", nil)
		return
	}
	if err != nil {
		panic(err)
	}

	notification, err := models.GetNotification(s.Notification)
	if err != nil {
		panic(err)
	}

	imageHistory, err := models.GetImageHistory(id)
	if err != nil {
		panic(err)
	}

	sentryJSON := struct {
		ID            primitive.ObjectID     `json:"id"`
		Name          string                 `json:"name"`
		Notification  *models.Notification   `json:"notification"`
		CreateTime    time.Time              `json:"createTime"`
		LastCheckTime time.Time              `json:"lastCheckTime"`
		Interval      int                    `json:"interval"`
		CheckCount    int                    `json:"checkCount"`
		NotifyCount   int                    `json:"notifyCount"`
		Image         *models.SentryImage    `json:"image"`
		ImageHistory  *models.ImageHistory   `json:"imageHistory"`
		Task          map[string]interface{} `json:"task"`
	}{
		s.ID, s.Name, notification, s.CreateTime, s.LastCheckTime,
		s.Interval, s.CheckCount, s.NotifyCount, &s.Image,
		imageHistory, s.Task,
	}

	JSONResponse(c, CodeOK, "", sentryJSON)

}

// SentryCreate creates a new sentry
func SentryCreate(c *gin.Context) {
	u, err := url.ParseRequestURI(c.Query("url"))

	if err != nil || !(strings.EqualFold(u.Scheme, "http") || strings.EqualFold(u.Scheme, "https")) {
		JSONResponse(c, CodeWrongParam, "Wrong protocol", nil)
		return
	}

	notification, err := primitive.ObjectIDFromHex(c.Query("notification"))

	if err != nil {
		JSONResponse(c, CodeWrongParam, "Wrong notificationId", nil)
		return
	}

	if !models.NotificationCheckOwner(notification, c.MustGet("userId").(primitive.ObjectID)) {
		JSONResponse(c, CodeNotExist, "notification does not exist", nil)
		return
	}

	x, _ := strconv.ParseInt(c.Query("x"), 10, 32)
	y, _ := strconv.ParseInt(c.Query("y"), 10, 32)
	width, _ := strconv.ParseInt(c.Query("width"), 10, 32)
	height, _ := strconv.ParseInt(c.Query("height"), 10, 32)

	if !(x >= 0 && y >= 0 && width > 0 && height > 0) {
		JSONResponse(c, CodeWrongParam, "Wrong area", nil)
		return
	}

	if width*height > 500*500 {
		JSONResponse(c, CodeAreaTooLarge, "", nil)
		return
	}

	var similarityThreshold float64
	if c.Query("similarityThreshold") == "" {
		// default value
		similarityThreshold = 0.9999
	} else {
		similarityThreshold, err = strconv.ParseFloat(c.Query("similarityThreshold"), 64)
		if err != nil || similarityThreshold <= 0 || similarityThreshold > 1 {
			JSONResponse(c, CodeWrongParam, "Invalid similarityThreshold", nil)
			return
		}
	}

	s := &models.Sentry{}
	s.ID = primitive.NewObjectID()
	s.Name = c.Query("name")
	s.User = c.MustGet("userId").(primitive.ObjectID)
	s.Notification = notification
	s.CreateTime = time.Now()
	s.NextCheckTime = time.Now()
	s.Interval = 4 * 60 // 4 hours
	s.CheckCount = 0
	s.NotifyCount = -1 // will be add 1 at the first check
	s.Image.File = "placeholder"
	s.Trigger.SimilarityThreshold = similarityThreshold
	s.Task = gin.H{
		"url":      u.String(),
		"timeout":  40000,
		"fullPage": false,
		"clip": gin.H{
			"x":      int(x),
			"width":  int(width),
			"height": int(height),
			"y":      int(y),
		},
		"viewport": gin.H{
			"width":    900,
			"isMobile": false,
		},
		"output": gin.H{
			"type": "png",
		},
	}

	err = models.AddSentry(s)
	if err != nil {
		panic(err)
	}

	JSONResponse(c, CodeOK, "", gin.H{
		"sentryId": s.ID.Hex(),
	})
}

func GetHistoryImage(c *gin.Context) {
	filename := c.Query("filename")
	// filename is unsafe
	if utils.ImageCheckFilename(filename) {
		filename = utils.ImageGetFullPath(filename, true)
		fileBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			c.String(404, "")
		} else {
			c.Data(200, "image/jpeg", fileBytes)
		}
	} else {
		c.String(404, "")
	}
}

func sentryTaskScheduler() {
	for {
		time.Sleep(2 * time.Minute)

		for {
			sentry, err := models.GetUncheckedSentry()
			if err != nil {
				// TODO: log
				break
			}
			if sentry == nil {
				break
			}
			// add task
			addSentryTask(sentry)
		}
	}
}

func compareSentryTaskImage(tid int32, ti *taskInfo) error {
	defer func() {
		// clean up
		taskq.infoMux.Lock()
		delete(taskq.info, tid)
		taskq.infoMux.Unlock()
	}()

	b, err := imaging.Decode(bytes.NewReader(ti.image))
	if err != nil {
		return errors.WithStack(err)
	}

	// first time
	if ti.baseImage.File == "placeholder" {

		imageFilename, err := utils.ImageSave(b)
		if err != nil {
			return errors.WithStack(err)
		}

		err = models.UpdateSentryAfterCheck(ti.sentryID, true, imageFilename)

		if err != nil {
			utils.ImageDelete(imageFilename, false)
		}
		return errors.WithStack(err)
	}

	a, err := imaging.Open(utils.ImageGetFullPath(ti.baseImage.File, false))

	if err != nil {
		// TODO: error handling
		return errors.WithStack(err)
	}

	similarity, err := utils.ImageCompare(a, b)
	if err != nil {
		return err
	}
	changed := float64(similarity) < ti.trigger.SimilarityThreshold
	newImage := ""
	if changed {
		// changed
		// save new image
		newImage, err = utils.ImageSave(b)
		if err != nil {
			return err
		}
	}

	log.Printf("[compareSentryTaskImage] Info: sentry: %s, similarity: %.2f%%, changed: %v \n", ti.sentryID.Hex(), similarity*100, changed)

	err = models.UpdateSentryAfterCheck(ti.sentryID, changed, newImage)

	if changed {
		if err == nil {
			// success

			// notification
			e := toggleNotification(ti.sentryID, ti.baseImage.Time, ti.baseImage.File, newImage, similarity)
			if e != nil {
				log.Printf("[toggleNotification] Error occurred in sentry: %s, err: %v \n", ti.sentryID.Hex(), e)
			}

			// delete old file (keep thumb)
			utils.ImageDelete(ti.baseImage.File, true)
		} else {
			// delete new file (delete all)
			utils.ImageDelete(newImage, false)
		}
	}

	return errors.WithStack(err)
}
