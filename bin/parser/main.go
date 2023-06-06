package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"hermes/common"
	"hermes/parser"
	"hermes/storage"

	"github.com/sirupsen/logrus"
)

var (
	metadataDir string
	logDir      string
	outputDir   string
	storEngine  string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal(err)
	}

	flag.StringVar(&logDir, "log_dir", "/var/log/collector", "The path of log directory")
	flag.StringVar(&outputDir, "output_dir", homeDir+string("/view"), "The path of view directory")
	flag.StringVar(&storEngine, "storage_engine", "file", "The storage engine")
	flag.Usage = Usage
}

func Usage() {
	fmt.Println("Usage: parser [log_dir] [output_dir]")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	err := common.LoadEnv(metadataDir)
	if err != nil {
		logrus.Fatal(err)
	}

	storEngineInst, err := storage.GetStorEngine(storEngine, logDir)
	if err != nil {
		logrus.Fatal(err)
	}

	logMetas, err := storEngineInst.Load()
	if err != nil {
		logrus.Fatal(err)
	}

	timestamps := []int64{}
	for timestamp, _ := range logMetas {
		timestamps = append(timestamps, timestamp)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })
	for _, timestamp := range timestamps {
		parser, err := parser.NewParser(logDir, outputDir, timestamp, logMetas[timestamp])
		if err != nil {
			logrus.Errorf("Failed to generate parser for timestamp [%d], err [%s]", timestamp, err)
			continue
		}

		if err := parser.Parse(); err != nil {
			logrus.Errorf("Failed to parse timestamp [%d], err [%s]", timestamp, err)
		}
	}
}
