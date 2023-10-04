package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"hermes/collector"
	"hermes/common"
	"hermes/log"
	"hermes/parser"

	"github.com/sirupsen/logrus"
)

const (
	JobCompleteChanSize = 16
)

var (
	metadataDir  string
	configDir    string
	logDir       string
	storEngine   string
	viewDir      string
	instantParse bool
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal(err)
	}

	flag.StringVar(&configDir, "config_dir", metadataDir+common.ConfigDirDefault, "The path of config directory")
	flag.StringVar(&logDir, "log_dir", common.LogDirDefault, "The path of log directory")
	flag.StringVar(&storEngine, "storage_engine", "file", "The storage engine")
	flag.StringVar(&viewDir, "view_dir", homeDir+common.ViewDirDefault, "The path of view directory")
	flag.BoolVar(&instantParse, "instant_parse", true, "Instant parse")
	flag.Usage = usage
}

func usage() {
	fmt.Println("Usage: collector [config_dir] [log_dir] [storage_engine] [instant_parse]")
	flag.PrintDefaults()
}

func logParseRoutine(ctx context.Context, jobCompleteSub chan log.LogMetaPubFormat) {
	for {
		select {
		case <-ctx.Done():
			return
		case logMetaPub := <-jobCompleteSub:
			parser, err := parser.NewParser(logDir, viewDir, logMetaPub.Timestamp, logMetaPub.LogMetadata)
			if err != nil {
				logrus.Errorf("Failed to generate parser for timestamp [%d], err [%s]", logMetaPub.Timestamp, err)
				continue
			}

			if err := parser.Parse(); err != nil {
				logrus.Errorf("Failed to parse timestamp [%d], err [%s]", logMetaPub.Timestamp, err)
			}

		}
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	flag.Parse()

	err := common.LoadEnv(metadataDir)
	if err != nil {
		logrus.Fatal(err)
	}

	var jobCompleteSub chan log.LogMetaPubFormat
	if instantParse {
		jobCompleteSub = make(chan log.LogMetaPubFormat, JobCompleteChanSize)
		go logParseRoutine(ctx, jobCompleteSub)
	}

	jobQueue, err := collector.NewJobQueue(configDir, logDir, storEngine, jobCompleteSub)
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
