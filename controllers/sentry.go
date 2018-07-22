package controllers

import (
	"net/url"
	"strings"
	"math"
	"image"
	"errors"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/websentry/websentry/middlewares"
	"github.com/websentry/websentry/models"
	"path"
	"github.com/websentry/websentry/config"
	"github.com/disintegration/imaging"
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"os"
	"io/ioutil"
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
		"timeout":  20000,
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

	id := addFullScreenshotTask(task)

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

func pixelDifference(a uint32, b uint32) float64 {
	return math.Abs(float64(a)-float64(b)) / 65535.0
}

func compareImage(a image.Image, b image.Image) (float32, error) {
	if a.Bounds() != b.Bounds() {
		return 0, errors.New("images with different size")
	}

	bounds := a.Bounds()
	total := 0
	v := 0.0
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {

			ar, ag, ab, _ := a.At(i, j).RGBA()
			br, bg, bb, _ := b.At(i, j).RGBA()
			v += pixelDifference(ar, br)
			v += pixelDifference(ag, bg)
			v += pixelDifference(ab, bb)

			total += 3
		}
	}
	return 1 - float32(v / float64(total)), nil
}

func saveImage(b []byte) string {

	// generate file name
	basePath := path.Join(config.GetFileStoragePath(), "sentry", "image")
	stime := time.Now().Format("20060102150405")
	filename := ""
	fullFilename := ""
	for {
		i := rand.Intn(200)

		filename = stime + "-" + strconv.Itoa(i) + ".png"
		fullFilename = path.Join(basePath, filename)
		_, err := os.Stat(fullFilename)
		if os.IsNotExist(err) {
			break
		}
	}

	// save
	ioutil.WriteFile(fullFilename, b, 0644)
	return filename
}

func compareSentryTaskImage(tid int32, ti *taskInfo) {
	defer func() {
		// clean up
		taskq.infoMux.Lock()
		delete(taskq.info, tid)
		taskq.infoMux.Unlock()
	}()

	file := path.Join(config.GetFileStoragePath(), "sentry", "image", ti.baseImage.File)

	a, err1 := imaging.Open(file)
	b, err2 := imaging.Decode(bytes.NewReader(ti.image))

	if err1!=nil || err2!=nil {
		// TODO: error handling
		fmt.Println("image error")
		return
	}

	v, _ := compareImage(a,b)
	changed := v < 0.9999
	imagePath := ""
	if changed {
		// changed
		// TODO: notification
		fmt.Println(ti.sentryId)
		fmt.Println("notification")
		fmt.Println(v)

		// save new image
		imagePath = saveImage(ti.image)

	}

	s := middlewares.GetDBSession()
	db := middlewares.SessionToDB(s)
	err := models.UpdateSentryAfterCheck(db, ti.sentryId, changed, imagePath, ti.version)
	s.Close()

	// make it run
	if err != nil {

	}


	//if changed {
	//	if err==nil {
	//		// delete old file
	//		os.Remove(file)
	//	} else {
	//		// delete new file
	//		os.Remove(imagePath)
	//	}
	//}

}
