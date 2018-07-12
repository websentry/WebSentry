package controllers

import (
	"sync"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	QUEUE_BUFFER = 100

	IN_QUEUE = 0
	ASSIGNED = 1
	COMPLETE = 2
)

type taskInfo struct {
	task gin.H
	expire time.Time
	status int
}

type taskQueue struct {

	// high priority queue (screenshot)
	pQueue chan int32
	// normal queue (sencry task)
	nQueue chan int32

	infoMux sync.Mutex
	info map[int32] *taskInfo
}

var taskq taskQueue

func init() {
    rand.Seed(time.Now().UnixNano())

	taskq.pQueue = make(chan int32, QUEUE_BUFFER)
	taskq.nQueue = make(chan int32, QUEUE_BUFFER)
	taskq.info = make(map[int32] *taskInfo)
}

func addFullScreenshotTask(task gin.H) int32 {
	ti := new(taskInfo)
	ti.task = task
	ti.status = IN_QUEUE
	ti.expire = time.Now().Add(time.Minute * 5)

	tid := rand.Int31()

	taskq.infoMux.Lock()
	var ok bool
	for true {
		if _, ok = taskq.info[tid]; !ok {
			break
		}
		tid = rand.Int31()
	}

	taskq.info[tid] = ti
	taskq.infoMux.Unlock()

	taskq.pQueue <- tid
	return tid
}


func SlaveInit(c *gin.Context) {
}

func SlaveFetchTask(c *gin.Context) {
}

func SlaveSubmitTask(c *gin.Context) {
}
