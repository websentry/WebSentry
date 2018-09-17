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