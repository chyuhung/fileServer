package main

import (
	"uploadServer/controler"
	"uploadServer/global"

	"github.com/gin-gonic/gin"
)

func main() {
	global.InitPath()
	e := gin.Default()
	e.POST("/fileServer/uploadAppend", controler.AppendHandle)
	e.POST("/fileServer/uploadNewFile", controler.UploadNewFile)
	e.GET("/fileServer/getProgress", controler.GetProgress)
	e.POST("/fileServer/uploadDelete", controler.UploadDelete)
	e.GET("/fileServer/download", controler.GetFile)
	e.Run(":27149")
}
