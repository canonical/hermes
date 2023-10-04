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
	mode        string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal(err)
	}

	flag.StringVar(&logDir, "log_dir", common.LogDirDefault, "The path of log directory")
	flag.StringVar(&outputDir, "output_dir", homeDir+common.ViewDirDefault, "The path of view directory")
	flag.StringVar(&storEngine, "storage_engine", "file", "The storage engine (file)")
	flag.StringVar(&mode, "mode", "oneshot", "Mode (oneshot|daemon)")
	flag.Usage = usage
}

func usage() {
	fmt.Println("Usage: parser [log_dir] [output_dir] [storage_engine] [mode]")
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
		logrus.Fatalf("Failed to get storage engine, err [%s]", err)
	}

	logMetas, err := storEngineInst.Load()
	if err != nil {
		logrus.Fatalf("Failed to load meta by storage engine, err [%s]", err)
	}

	timestamps := []int64{}
	for timestamp := range logMetas {
		timestamps = append(timestamps, timestamp)
	}
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })
	for _, timestamp := range timestamps {
		for _, logMeta := range logMetas[timestamp] {
			parser, err := parser.NewParser(logDir, outputDir, timestamp, logMeta)
			if err != nil {
				logrus.Errorf("Failed to generate parser for timestamp [%d], err [%s]", timestamp, err)
				continue
			}

			if err := parser.Parse(); err != nil {
				logrus.Errorf("Failed to parse timestamp [%d], err [%s]", timestamp, err)
			}
		}
	}
}
