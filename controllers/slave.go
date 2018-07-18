package controllers

import (
	"math/rand"
	"sync"
	"time"
	"strconv"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
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
	channel chan bool

	// sentry
	sentryId bson.ObjectId
	version int
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
	ti.version = s.Version
	ti.expire = time.Now().Add(time.Minute * 5)

	tid := insertTaskinfo(ti)

	taskq.nQueue <- tid
	return tid
}

func addFullScreenshotTask(task gin.H) int32 {
	ti := new(taskInfo)
	ti.task = task
	ti.mode = TM_FULLSCREEN
	ti.status = TS_IN_QUEUE
	ti.expire = time.Now().Add(time.Minute * 1)
	ti.channel = make(chan bool)

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
	c.JSON(200, gin.H{
		"code":   0,
		"msg":    "OK",
	})
}

func SlaveFetchTask(c *gin.Context) {
	tid, ti := getTask()

	if tid>=0 {
		c.JSON(200, gin.H{
			"code":   0,
			"msg":    "OK",
			"taskId": tid,
			"task": ti.task,
		})
	} else {
		c.JSON(200, gin.H{
			"code":   0,
			"msg":    "OK",
			"taskId": -1,
		})
	}
}

func SlaveSubmitTask(c *gin.Context) {
	taskq.infoMux.Lock()
	defer taskq.infoMux.Unlock()

	tid, err := strconv.ParseInt(c.Query("taskid"), 10, 32)
	if err!=nil {
		c.JSON(200, gin.H{
			"code": -2,
			"msg":  "Wrong parameter",
		})
		return
	}

	ti, ok := taskq.info[int32(tid)]
	if !ok {
		c.JSON(200, gin.H{
			"code": -3,
			"msg":  "Record not exists",
		})
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
			c.JSON(200, gin.H{
				"code": -2,
				"msg":  "Wrong parameter",
			})
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


	c.JSON(200, gin.H{
		"code":   0,
		"msg":    "OK",
	})
}

func waitFullScreenshot(c *gin.Context) {

	tid, err := strconv.ParseInt(c.Query("taskid"), 10, 32)
	if err!=nil {
		c.JSON(200, gin.H{
			"code": -2,
			"msg":  "Wrong parameter",
		})
		return
	}

	taskq.infoMux.Lock()

	ti, ok := taskq.info[int32(tid)]
	if !ok || ti.mode!=TM_FULLSCREEN {
		taskq.infoMux.Unlock()
		c.JSON(200, gin.H{
			"code": -3,
			"msg":  "Record not exists",
		})
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
		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "OK",
			"complete": false,
		})
	} else {
		taskq.infoMux.Lock()
		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "OK",
			"complete": true,
			"feedbackCode": ti.feedbackCode,
			"feedbackMsg": ti.feedbackMsg,
		})
		taskq.infoMux.Unlock()

	}
}

func getFullScreenshot(c *gin.Context) {
	tid, err := strconv.ParseInt(c.Query("taskid"), 10, 32)
	if err!=nil {
		c.JSON(200, gin.H{
			"code": -2,
			"msg":  "Wrong parameter",
		})
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

	if ti.status!=TS_COMPLETE || ti.image==nil || ti.mode!=TM_FULLSCREEN {
		c.String(404, "")
		return
	}

	c.Data(200, "image/jpeg", ti.image)
}
