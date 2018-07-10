package controllers

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)


// [url] the url of the page that needs screenshot
func SentryRequestFullScreenshot(c *gin.Context) {
	u, err := url.ParseRequestURI(c.Query("url"))
	if (err != nil || !(strings.EqualFold(u.Scheme, "http") || strings.EqualFold(u.Scheme, "https"))) {
		c.JSON(200, gin.H{
				"code": -3,
				"msg": "Wrong parameter",
			})
		return
	}

	// TODO: handle extreme long page

	task := gin.H{
		"url": u.String(),
		"timeout": 20000,
		"fullPage": true,
		"viewport": gin.H{
			"width": 900,
			"isMobile": false,
		},
		"output": gin.H{
			"type": "jpg",
			"progressive": true,
			"quality": 20,
		},
	}

	id := addFullScreenshotTask(task)

	c.JSON(200, gin.H{
			"code": 1,
			"msg": "OK",
			"taskId": id,
		})
}

func SentryGetFullScreenshot(c *gin.Context) {

}
