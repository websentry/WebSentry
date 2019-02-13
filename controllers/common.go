package controllers

func Init() {
	go sentryTaskScheduler()
}
