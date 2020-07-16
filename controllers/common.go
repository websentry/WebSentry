package controllers

func Init() {
	go sentryTaskScheduler()

	// worker
	taskq.pQueue = make(chan int32, queueBuffer)
	taskq.nQueue = make(chan int32, queueBuffer)
	taskq.info = make(map[int32]*taskInfo)

	go cleanTask()
}
