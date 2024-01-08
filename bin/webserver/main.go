package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"hermes/common"
	"hermes/parser"
	hprom "hermes/prometheus"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var contentParser *ContentParser

var (
	metadataDir           string
	frontendDir           string
	viewDir               string
	rawDataDir            string
	runPrometheusExporter bool
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal(err)
	}

	flag.StringVar(&frontendDir, "frontend_dir", metadataDir+common.FrontendDirDefault, "The path of frontend directory")
	flag.StringVar(&viewDir, "view_dir", homeDir+common.ViewDirDefault, "The path of view directory")
	flag.StringVar(&rawDataDir, "raw_dir", common.LogDirDefault+"/data", "The path of raw data directory (prometheus only)")
	flag.Usage = usage
	flag.BoolVar(&runPrometheusExporter, "prometheus", false, "Whether to serve the prometheus exporter at /metrics")
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

	page := router.Group("/")
	{
		page.GET("", func(ctx *gin.Context) {
			ctx.HTML(http.StatusOK, "index.html", nil)
		})
	}

	api := router.Group("/api")
	{
		api.GET("/routines", func(ctx *gin.Context) {
			ctx.ProtoBuf(http.StatusOK, contentParser.GetRoutines())
		})
	}

	cpu := router.Group("/cpu")
	{
		cpu.GET("/cpu_profile", func(ctx *gin.Context) {
			path := filepath.Join(viewDir, "cpu_profile", "overview")
			ctx.File(path)
		})
		cpu.GET("/cpu_profile/:timestamp", func(ctx *gin.Context) {
			timestamp := ctx.Param("timestamp")
			path := filepath.Join(viewDir, "cpu_profile", timestamp, parser.ParsedPostfix[parser.CpuProfileJob])
			ctx.File(path)
		})
	}

	mem := router.Group("/memory")
	{
		mem.GET("/memleak_profile", func(ctx *gin.Context) {
			path := filepath.Join(viewDir, "memleak_profile", "overview")
			ctx.File(path)
		})
		mem.GET("/memleak_profile/:timestamp", func(ctx *gin.Context) {
			timestamp := ctx.Param("timestamp")
			path := filepath.Join(viewDir, "memleak_profile", timestamp, parser.ParsedPostfix[parser.CpuProfileJob])
			ctx.File(path)
		})
	}

	io := router.Group("/io")
	{
		io.GET("/io_latency", func(ctx *gin.Context) {
			path := filepath.Join(viewDir, "io_latency", "overview")
			ctx.File(path)
		})
		io.GET("/io_latency/:timestamp", func(ctx *gin.Context) {
			timestamp := ctx.Param("timestamp")
			path := filepath.Join(viewDir, "io_latency", timestamp, parser.ParsedPostfix[parser.IoLatencyJob])
			ctx.File(path)
		})
	}

	if runPrometheusExporter {
		reg := prometheus.NewRegistry()
		he := hprom.NewHermesExporter(viewDir, rawDataDir)
		reg.MustRegister(he)
		hh := hprom.HermesPrometheusHandler(reg)
		router.GET("/metrics", hh)
	}

	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{})
	})
	router.Run()
}
