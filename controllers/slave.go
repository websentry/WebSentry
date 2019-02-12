package controllers

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"io/ioutil"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/websentry/websentry/models"
)

const (
	LONG_POLLING_TIMEOUT = 60 * time.Second
	QUEUE_BUFFER = 100

	TS_IN_QUEUE = 0
	TS_ASSIGNED = 1
	TS_COMPLETE = 2

	TM_FULLSCREEN = 0
	TM_SENTRY = 1
)

type taskInfo struct {
	// common
	mode int
	status int
	image []byte
	feedbackCode int
	feedbackMsg string
	task   gin.H
	expire time.Time

	// screenshot
	user primitive.ObjectID
	channel chan bool

	// sentry
	sentryId primitive.ObjectID
	baseImage *models.SentryImage
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

	taskq.pQueue = make(chan int32, QUEUE_BUFFER)
	taskq.nQueue = make(chan int32, QUEUE_BUFFER)
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
	ti.mode = TM_SENTRY
	ti.status = TS_IN_QUEUE
	ti.sentryId = s.Id
	ti.baseImage = &s.Image
	ti.expire = time.Now().Add(time.Minute * 5)

	tid := insertTaskinfo(ti)

	taskq.nQueue <- tid
	return tid
}

func addFullScreenshotTask(task gin.H, user primitive.ObjectID) int32 {
	ti := new(taskInfo)
	ti.task = task
	ti.mode = TM_FULLSCREEN
	ti.status = TS_IN_QUEUE
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
		case <-time.After(LONG_POLLING_TIMEOUT):
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
		ti.status = TS_ASSIGNED

		return tid, ti
	}
}

func SlaveInit(c *gin.Context) {
	JsonResponse(c, CodeOK, "", nil)
}

func SlaveFetchTask(c *gin.Context) {
	tid, ti := getTask()

	if tid>=0 {
		JsonResponse(c, CodeOK, "", gin.H{
			"taskId": tid,
			"task": ti.task,
		})
	} else {
		JsonResponse(c, CodeOK, "", gin.H{
			"taskId": -1,
		})
	}
}

func SlaveSubmitTask(c *gin.Context) {
	taskq.infoMux.Lock()
	defer taskq.infoMux.Unlock()

	tid, err := strconv.ParseInt(c.Query("taskId"), 10, 32)
	if err!=nil {
		JsonResponse(c, CodeWrongParam, "", nil)
		return
	}

	ti, ok := taskq.info[int32(tid)]
	if !ok {
		JsonResponse(c, CodeNotExist, "", nil)
		return
	}

	feedbackCode, err := strconv.ParseInt(c.Query("feedback"), 10, 32)
	if feedbackCode!=0 {
		ti.feedbackCode = int(feedbackCode)
		ti.feedbackMsg = c.Query("msg")
	} else {
		ti.feedbackCode = 0

		fileH, err := c.FormFile("image")
		if err!=nil {
			JsonResponse(c, CodeWrongParam, "Image error", nil)
			return
		}

		file, _ := fileH.Open()
		ti.image, _ = ioutil.ReadAll(file)

	}

	ti.status = TS_COMPLETE
	ti.expire = time.Now().Add(time.Minute * 2)

	if ti.mode==TM_FULLSCREEN {
		close(ti.channel)
	} else {
		go compareSentryTaskImage(int32(tid), ti)
	}

	JsonResponse(c, CodeOK, "", nil)
}

func waitFullScreenshot(c *gin.Context) {

	tid, err := strconv.ParseInt(c.Query("taskId"), 10, 32)
	if err!=nil {
		JsonResponse(c, CodeWrongParam, "", nil)
		return
	}

	taskq.infoMux.Lock()

	ti, ok := taskq.info[int32(tid)]
	if !ok || ti.mode!=TM_FULLSCREEN || ti.user!=c.MustGet("userId") {
		taskq.infoMux.Unlock()
		JsonResponse(c, CodeNotExist, "", nil)
		return
	}
	incomplete := ti.status!= TS_COMPLETE
	taskq.infoMux.Unlock()

	timeoutFlag := false
	if incomplete {
		select {
		case <-ti.channel:
			timeoutFlag = false
		case <-time.After(LONG_POLLING_TIMEOUT):
			timeoutFlag = true
		}
	}

	if timeoutFlag {
		JsonResponse(c, CodeOK, "", gin.H{
			"complete": false,
		})
	} else {
		taskq.infoMux.Lock()
		JsonResponse(c, CodeOK, "", gin.H{
			"complete": true,
			"feedbackCode": ti.feedbackCode,
			"feedbackMsg": ti.feedbackMsg,
		})
		taskq.infoMux.Unlock()

	}
}

func getFullScreenshot(c *gin.Context) {
	tid, err := strconv.ParseInt(c.Query("taskId"), 10, 32)
	if err!=nil {
		JsonResponse(c, CodeWrongParam, "", nil)
		return
	}


	taskq.infoMux.Lock()

	ti, ok := taskq.info[int32(tid)]
	if !ok {
		taskq.infoMux.Unlock()
		c.String(404, "")
		return
	}
	taskq.infoMux.Unlock()

	if ti.status!=TS_COMPLETE || ti.image==nil || ti.mode!=TM_FULLSCREEN || ti.user!=c.MustGet("userId") {
		c.String(404, "")
		return
	}

	c.Data(200, "image/jpeg", ti.image)
}
