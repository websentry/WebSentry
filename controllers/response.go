package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeOK         = 0
	CodeAuthError  = -1
	CodeWrongParam = -2

	CodeExceededLimits = -4
	CodeNotExist       = -5
	CodeAlreadyExist   = -6

	CodeAreaTooLarge = -1001
)

var msgMap = map[int]string{
	// common
	0:  "OK",
	-1: "Authorization error",
	-2: "Invalid parameter",

	-4: "Exceeded limits",
	-5: "Record does not exist",
	-6: "Record already exists",

	// specific
	// create sentry
	-1001: "Area too large",
}

func JSONResponse(c *gin.Context, code int, detail string, data interface{}) {
	json := gin.H{}
	json["code"] = code
	json["msg"] = msgMap[code]
	if detail != "" {
		json["detail"] = detail
	}
	if data != nil {
		json["data"] = data
	}

	c.JSON(http.StatusOK, json)
}

func InternalErrorResponse(c *gin.Context, err error) {
	c.Abort()
	log.Printf("Internal error: \n %+v", err)
	c.String(http.StatusInternalServerError, "Internal Server Error")
}
