package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

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

	id := addFullScreenshotTask(task, c.MustGet("userId").(int64))

	JSONResponse(c, CodeOK, "", gin.H{
		"taskId": id,
	})
}

func SentryWaitFullScreenshot(c *gin.Context) {
	waitFullScreenshot(c)
}

type SentryListItemJSON struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	URL           string     `json:"url"`
	LastCheckTime *time.Time `json:"lastCheckTime"`
}

func SentryList(c *gin.Context) {
	var results []models.Sentry

	err := models.Transaction(func(tx models.TX) (err error) {
		results, err = tx.GetUserSentries(c.MustGet("userId").(int64))
		return
	})
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	sentries := make([]SentryListItemJSON, len(results))
	for i := range results {
		sentries[i].Name = results[i].Name
		sentries[i].ID = strconv.FormatInt(results[i].ID, 16)

		var task map[string]interface{}
		err = errors.WithStack(json.Unmarshal([]byte(results[i].Task), &task))
		if err != nil {
			InternalErrorResponse(c, err)
			return
		}
		sentries[i].URL = task["url"].(string)
		sentries[i].LastCheckTime = results[i].LastCheckTime
	}

	JSONResponse(c, CodeOK, "", gin.H{
		"sentries": sentries,
	})
}

type SentryImageJson struct {
	File      string    `json:"file"`
	CreatedAt time.Time `json:"createdAt"`
}

type NotificationMethodJson struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type SentryJSON struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Notification  NotificationMethodJson `json:"notification"`
	LastCheckTime *time.Time             `json:"lastCheckTime"`
	Interval      int                    `json:"interval"`
	CheckCount    int                    `json:"checkCount"`
	NotifyCount   int                    `json:"notifyCount"`
	ImageHistory  []SentryImageJson      `json:"imageHistory"`
	Task          map[string]interface{} `json:"task"`
	CreatedAt     time.Time              `json:"createdAt"`
}

func SentryInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Query("id"), 16, 64)
	if err != nil {
		JSONResponse(c, CodeWrongParam, "Wrong sentry id", nil)
		return
	}

	var exists bool
	var s *models.Sentry
	var notification *models.NotificationMethod
	var imageHistory []models.SentryImage

	err = models.Transaction(func(tx models.TX) (err error) {
		s, err = tx.GetSentry(id)
		if err != nil {
			if models.IsErrNoDocument(err) {
				exists = false
				return nil
			} else {
				return err
			}
		}
		exists = s.UserID == c.MustGet("userId").(int64)

		notification, err = tx.GetNotification(s.NotificationID)
		if err != nil {
			return
		}
		imageHistory, err = tx.GetImageHistory(id)
		return
	})
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}
	if !exists {
		JSONResponse(c, CodeNotExist, "", nil)
		return
	}

	imageHistoryJSON := make([]SentryImageJson, len(imageHistory))
	for i := range imageHistory {
		imageHistoryJSON[i].CreatedAt = imageHistory[i].CreatedAt
		imageHistoryJSON[i].File = imageHistory[i].File
	}

	var task map[string]interface{}
	err = errors.WithStack(json.Unmarshal([]byte(s.Task), &task))
	if err != nil {
		InternalErrorResponse(c, err)
		return
	}

	notificationJson := NotificationMethodJson{
		strconv.FormatInt(notification.ID, 16),
		notification.Name,
		notification.Type,
	}

	sentryJSON := SentryJSON{
		strconv.FormatInt(s.ID, 16), s.Name, notificationJson, s.LastCheckTime,
		s.Interval, s.CheckCount, s.NotifyCount,
		imageHistoryJSON, task, s.CreatedAt,
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

	notification, err := strconv.ParseInt(c.Query("notification"), 16, 64)
	if err != nil {
		JSONResponse(c, CodeWrongParam, "Wrong notificationId", nil)
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

	similarityThreshold, err := strconv.ParseFloat(c.DefaultQuery("similarityThreshold", "0.9999"), 64)
	if err != nil || similarityThreshold <= 0 || similarityThreshold > 1 {
		JSONResponse(c, CodeWrongParam, "Invalid similarityThreshold", nil)
		return
	}

	interval, err := strconv.Atoi(c.DefaultQuery("interval", "240")) // 4 hours
	if err != nil || interval < 15 {
		JSONResponse(c, CodeWrongParam, "Invalid interval", nil)
		return
	}

	s := &models.Sentry{}
	s.Name = c.Query("name")
	s.RunningState = models.RSRunning
	s.UserID = c.MustGet("userId").(int64)
	s.NotificationID = notification
	s.NextCheckTime = time.Now()
	s.Interval = interval
	s.CheckCount = 0
	s.NotifyCount = 0

	trigger := models.Trigger{SimilarityThreshold: similarityThreshold}
	triggerJSON, err := json.Marshal(&trigger)
	if err != nil {
		InternalErrorResponse(c, errors.WithStack(err))
		return
	}
	s.Trigger = string(triggerJSON)

	task := gin.H{
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

	taskJSON, err := json.Marshal(&task)
	if err != nil {
		InternalErrorResponse(c, errors.WithStack(err))
		return
	}
	s.Task = string(taskJSON)

	var sid int64
	err = models.Transaction(func(tx models.TX) (err error) {
		sid, err = tx.CreateSentry(s)
		return
	})

	if err != nil {
		if errors.Is(err, models.ErrInvalidNotificationID) {
			JSONResponse(c, CodeNotExist, "notification does not exist", nil)
		} else {
			InternalErrorResponse(c, err)
		}
		return
	}

	JSONResponse(c, CodeOK, "", gin.H{
		"sentryId": strconv.FormatInt(sid, 16),
	})
}

func SentryRemove(c *gin.Context) {
	id, err := strconv.ParseInt(c.Query("id"), 16, 64)
	if err != nil {
		JSONResponse(c, CodeWrongParam, "Wrong sentry id", nil)
		return
	}

	err = models.Transaction(func(tx models.TX) (err error) {
		return tx.DeleteSentry(id, c.MustGet("userId").(int64))
	})

	if err != nil {
		if models.IsErrNoDocument(err) {
			JSONResponse(c, CodeNotExist, "", nil)
		} else {
			InternalErrorResponse(c, err)
		}
		return
	}

	JSONResponse(c, CodeOK, "", gin.H{})
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
			var sentry *models.Sentry
			var image *models.SentryImage
			err := models.Transaction(func(tx models.TX) (err error) {
				sentry, image, err = tx.GetUncheckedSentry()
				return
			})
			if err != nil {
				// TODO: log
				break
			}
			if sentry == nil {
				break
			}
			// add task
			// TODO: log
			_, _ = addSentryTask(sentry, image)
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
	if ti.baseImage == nil {

		imageFilename, err := utils.ImageSave(b)
		if err != nil {
			return errors.WithStack(err)
		}

		err = models.Transaction(func(tx models.TX) (err error) {
			return tx.UpdateSentryAfterCheck(ti.sentryID, true, imageFilename)
		})

		if err != nil {
			utils.ImageDelete(imageFilename, false)
		}

		if errors.Is(err, models.ErrSentryNotRunning) {
			log.Println(err)
			err = nil
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

	log.Printf("[compareSentryTaskImage] Info: sentry: %x, similarity: %.2f%%, changed: %v \n", ti.sentryID, similarity*100, changed)

	err = models.Transaction(func(tx models.TX) (err error) {
		return tx.UpdateSentryAfterCheck(ti.sentryID, changed, newImage)
	})

	if changed {
		if err == nil {
			// success

			// notification
			e := toggleNotification(ti.sentryID, ti.baseImage.CreatedAt, ti.baseImage.File, newImage, similarity)
			if e != nil {
				log.Printf("[toggleNotification] Error occurred in sentry: %x, err: %v \n", ti.sentryID, e)
			}

			// delete old file (keep thumb)
			utils.ImageDelete(ti.baseImage.File, true)
		} else {
			// delete new file (delete all)
			utils.ImageDelete(newImage, false)
		}
	}

	if errors.Is(err, models.ErrSentryNotRunning) {
		log.Println(err)
		err = nil
	}
	return errors.WithStack(err)
}
