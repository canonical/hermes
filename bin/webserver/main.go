package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var contentParser *ContentParser

var (
	frontendDir string
	viewDir     string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal(err)
	}

	flag.StringVar(&frontendDir, "frontend_dir", homeDir+string("/frontend"), "The path of frontend directory")
	flag.StringVar(&viewDir, "view_dir", homeDir+string("/view"), "The path of view directory")
	flag.Usage = usage
}

func usage() {
	fmt.Println("Usage: webserver [frontend_dir] [view_dir]")
	flag.PrintDefaults()
}

func main() {
	router := gin.Default()

	flag.Parse()

	contentParser = NewContentParser(viewDir)
	router.LoadHTMLGlob(frontendDir + string("/*.html"))
	router.Static("/assets", frontendDir+string("/assets"))
	router.Static("/css", frontendDir+string("/css"))
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
