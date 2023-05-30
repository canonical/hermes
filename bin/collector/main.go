package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"hermes/collector"
	"hermes/common"

	"github.com/sirupsen/logrus"
)

var (
	metadataDir string
	logDir      string
	storEngine  string
)

func init() {
	flag.StringVar(&logDir, "log_dir", "/var/log/collector/", "The path of log directory")
	flag.StringVar(&storEngine, "storage_engine", "file", "The storage engine")
	flag.Usage = Usage
}

func Usage() {
	fmt.Println("Usage: collector [output_dir] [storage_engine]")
	flag.PrintDefaults()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	flag.Parse()

	configDir := metadataDir + "/config/"

	err := common.LoadEnv(metadataDir)
	if err != nil {
		logrus.Fatal(err)
	}

	jobQueue, err := collector.NewJobQueue(configDir, logDir, storEngine)
	if err != nil {
		logrus.Fatal(err)
	}

	configWatcher, err := collector.NewConfigWatcher(jobQueue.Comm)
	if err != nil {
		logrus.Fatal(err)
	}

	jobQueue.Run(ctx)

	configWatcher.Run(ctx, configDir)
	defer configWatcher.Release()

	<-ctx.Done()
}
