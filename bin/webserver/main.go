package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var contentParser *ContentParser

func main() {
	router := gin.Default()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal(err)
	}

	viewDir := homeDir + string("/view")
	contentParser = NewContentParser(viewDir)
	router.LoadHTMLGlob(homeDir + string("/frontend/*.html"))
	router.Static("/assets", homeDir+string("/frontend/assets"))
	router.Static("/view", viewDir)

	page := router.Group("/")
	{
		page.GET("", func(ctx *gin.Context) {
			ctx.HTML(http.StatusOK, "index.html", nil)
		})
	}

	api := router.Group("/api")
	{
		api.GET("/tasks", func(ctx *gin.Context) {
			ctx.ProtoBuf(http.StatusOK, contentParser.GetTasks())
		})
	}

	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{})
	})
	router.Run()
}
