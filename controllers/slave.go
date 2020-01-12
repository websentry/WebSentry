package controllers

import (
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/websentry/websentry/models"
	"github.com/websentry/websentry/utils"
)

const (
	// we need a timeout that is less than the timeout in proxy
	// nginx default is 60s and cloudflare is 100s
	longPollingTimeout = 42 * time.Second

	queueBuffer = 100
)

type taskStatus int8

const (
	taskStatusInQueue taskStatus = iota
	taskStatusAssigned
	taskStatusCompleted
)

type taskMode int8

const (
	taskModeFullScreen taskMode = iota
	taskModeSentry
)

type taskInfo struct {
	// common
	mode         taskMode
	status       taskStatus
	image        []byte
	feedbackCode int
	feedbackMsg  string
	task         gin.H
	expire       time.Time

	// screenshot
	user     primitive.ObjectID
	channel  chan bool
	tmpToken string // tmp token for get request for the actual image

	// sentry
	sentryID  primitive.ObjectID
	baseImage *models.SentryImage
	trigger   *models.Trigger
}

type taskQueue struct {

	// high priority queue (screenshot)
	pQueue chan int32
	// normal queue (sentry task)
	nQueue chan int32

	infoMux sync.Mutex
	info    map[int32]*taskInfo
}

var taskq taskQueue

func init() {
	rand.Seed(time.Now().UnixNano())

	taskq.pQueue = make(chan int32, queueBuffer)
	taskq.nQueue = make(chan int32, queueBuffer)
	taskq.info = make(map[int32]*taskInfo)

	go cleanTask()
}

func cleanTask() {
	for {
		time.Sleep(10 * time.Minute)
		n := time.Now()
		taskq.infoMux.Lock()
		for k, v := range taskq.info {
			if n.After(v.expire) {
				delete(taskq.info, k)
			}
		}

		taskq.infoMux.Unlock()
	}
}

func insertTaskinfo(ti *taskInfo) int32 {
	tid := rand.Int31()

	taskq.infoMux.Lock()
	var ok bool
	for {
		if _, ok = taskq.info[tid]; !ok {
			break
		}
		tid = rand.Int31()
	}

	taskq.info[tid] = ti
	taskq.infoMux.Unlock()

	return tid
}

func addSentryTask(s *models.Sentry) int32 {
	ti := new(taskInfo)
	ti.task = s.Task
	ti.mode = taskModeSentry
	ti.status = taskStatusInQueue
	ti.sentryID = s.ID
	ti.baseImage = &s.Image
	ti.trigger = &s.Trigger
	ti.expire = time.Now().Add(time.Minute * 5)

	tid := insertTaskinfo(ti)

	taskq.nQueue <- tid
	return tid
}

func addFullScreenshotTask(task gin.H, user primitive.ObjectID) int32 {
	ti := new(taskInfo)
	ti.task = task
	ti.mode = taskModeFullScreen
	ti.status = taskStatusInQueue
	ti.expire = time.Now().Add(time.Minute * 1)
	ti.channel = make(chan bool)
	ti.user = user

	tid := insertTaskinfo(ti)

	taskq.pQueue <- tid
	return tid
}

func getTask() (int32, *taskInfo) {
	var tid int32

	for {
		// select from priority queue first
		select {
		case tid = <-taskq.pQueue:
			goto assign
		default:
		}

		// select from the one that has task
		select {
		case tid = <-taskq.pQueue:
			goto assign
		case tid = <-taskq.nQueue:
			goto assign
		case <-time.After(longPollingTimeout):
			return -1, nil
		}

	assign:
		taskq.infoMux.Lock()
		// check existence
		ti, ok := taskq.info[tid]
		if !ok {
			taskq.infoMux.Unlock()
			continue
		}
		// check expiration
		if time.Now().After(ti.expire) {
			delete(taskq.info, tid)
			taskq.infoMux.Unlock()
			continue
		}
		taskq.infoMux.Unlock()

		ti.expire = time.Now().Add(time.Minute * 4)
		ti.status = taskStatusAssigned

		return tid, ti
	}
}

func SlaveInit(c *gin.Context) {
	JSONResponse(c, CodeOK, "", nil)
}

func SlaveFetchTask(c *gin.Context) {
	tid, ti := getTask()

	if tid >= 0 {
		JSONResponse(c, CodeOK, "", gin.H{
			"taskId": tid,
			"task":   ti.task,
		})
	} else {
		JSONResponse(c, CodeOK, "", gin.H{
			"taskId": -1,
		})
	}
}

func SlaveSubmitTask(c *gin.Context) {
	taskq.infoMux.Lock()
	defer taskq.infoMux.Unlock()

	tid, err := strconv.ParseInt(c.Query("taskId"), 10, 32)
	if err != nil {
		JSONResponse(c, CodeWrongParam, "", nil)
		return
	}

	ti, ok := taskq.info[int32(tid)]
	if !ok {
		JSONResponse(c, CodeNotExist, "", nil)
		return
	}

	feedbackCode, err := strconv.ParseInt(c.Query("feedback"), 10, 32)
	if feedbackCode != 0 {
		ti.feedbackCode = int(feedbackCode)
		ti.feedbackMsg = c.Query("msg")
	} else {
		ti.feedbackCode = 0

		fileH, err := c.FormFile("image")
		if err != nil {
			JSONResponse(c, CodeWrongParam, "Image error", nil)
			return
		}

		file, _ := fileH.Open()
		ti.image, _ = ioutil.ReadAll(file)

	}

	ti.status = taskStatusCompleted
	ti.expire = time.Now().Add(time.Minute * 2)

	if ti.mode == taskModeFullScreen {
		ti.tmpToken = utils.RandStringBytes(16)
		close(ti.channel)
	} else {
		// taskModeSentry
		go func() {
			err = compareSentryTaskImage(int32(tid), ti)
			if err != nil {
				log.Printf("[compareSentryTaskImage] Error occurred in task: %d, err: %v", tid, err)
			}
		}()
	}

	JSONResponse(c, CodeOK, "", nil)
}

func waitFullScreenshot(c *gin.Context) {

	tid, err := strconv.ParseInt(c.Query("taskId"), 10, 32)
	if err != nil {
		JSONResponse(c, CodeWrongParam, "", nil)
		return
	}

	taskq.infoMux.Lock()

	ti, ok := taskq.info[int32(tid)]
	if !ok || ti.mode != taskModeFullScreen || ti.user != c.MustGet("userId") {
		taskq.infoMux.Unlock()
		JSONResponse(c, CodeNotExist, "", nil)
		return
	}
	incomplete := ti.status != taskStatusCompleted
	taskq.infoMux.Unlock()

	timeoutFlag := false
	if incomplete {
		select {
		case <-ti.channel:
			timeoutFlag = false
		case <-time.After(longPollingTimeout):
			timeoutFlag = true
		}
	}

	if timeoutFlag {
		JSONResponse(c, CodeOK, "", gin.H{
			"complete": false,
		})
	} else {
		taskq.infoMux.Lock()
		JSONResponse(c, CodeOK, "", gin.H{
			"complete":     true,
			"imageToken":   ti.tmpToken,
			"feedbackCode": ti.feedbackCode,
			"feedbackMsg":  ti.feedbackMsg,
		})
		taskq.infoMux.Unlock()

	}
}

func GetFullScreenshotImage(c *gin.Context) {
	tid, err := strconv.ParseInt(c.Query("taskId"), 10, 32)
	if err != nil {
		c.String(400, "")
		return
	}

	taskq.infoMux.Lock()

	ti, ok := taskq.info[int32(tid)]
	// use imageToken as auth, not WS-User-Token
	if !ok || ti.tmpToken != c.Query("imageToken") {
		taskq.infoMux.Unlock()
		c.String(404, "")
		return
	}
	taskq.infoMux.Unlock()

	if ti.status != taskStatusCompleted || ti.image == nil || ti.mode != taskModeFullScreen {
		c.String(404, "")
		return
	}

	c.Data(200, "image/jpeg", ti.image)
}
