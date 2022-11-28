package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/yukariatlas/hermes/collector"
)

var (
	configDir string
	outputDir string
)

func init() {
	flag.StringVar(&configDir, "config_dir", "/root/config/", "The path of config directory")
	flag.StringVar(&outputDir, "output_dir", "/var/log/collector/", "The path of output directory")
	flag.Usage = usage
}

func usage() {
	fmt.Println("Usage: collector [config_dir] [output_dir]")
	flag.PrintDefaults()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	flag.Parse()

	taskQueue, err := collector.NewTaskQueue()
	if err != nil {
		logrus.Fatal(err)
	}

	configWatcher, err := collector.NewConfigWatcher(taskQueue.Comm)
	if err != nil {
		logrus.Fatal(err)
	}

	taskQueue.Run(ctx, outputDir)

	configWatcher.Run(ctx, configDir)
	defer configWatcher.Release()

	<-ctx.Done()
}
