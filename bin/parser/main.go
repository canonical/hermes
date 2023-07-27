package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"hermes/common"
	"hermes/log"
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

var (
	modeMap = map[string]func(){
		"oneshot": OneshotParser,
		"daemon":  DaemonizedParser,
	}
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal(err)
	}

	flag.StringVar(&logDir, "log_dir", "/var/log/collector", "The path of log directory")
	flag.StringVar(&outputDir, "output_dir", homeDir+string("/view"), "The path of view directory")
	flag.StringVar(&storEngine, "storage_engine", "file", "The storage engine (file)")
	flag.StringVar(&mode, "mode", "oneshot", "Mode (oneshot|daemon)")
	flag.Usage = Usage
}

func Usage() {
	fmt.Println("Usage: parser [log_dir] [output_dir] [storage_engine] [mode]")
	flag.PrintDefaults()
}

func OneshotParser() {
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

func DaemonizedParser() {
	pubsub, err := common.NewPubSub(common.Sub, common.JobCompleteTopic)
	if err != nil {
		logrus.Fatalf("Failed to create pubsub, err [%s]", err)
	}

	for {
		var logMetaPub log.LogMetaPubFormat
		bytes := pubsub.Recv()
		if err := json.Unmarshal(bytes, &logMetaPub); err != nil {
			logrus.Errorf("Failed to unmarshal data, err [%s]", err)
			continue
		}
		parser, err := parser.NewParser(logDir, outputDir, logMetaPub.Timestamp, logMetaPub.LogMetadata)
		if err != nil {
			logrus.Errorf("Failed to generate parser for timestamp [%d], err [%s]", logMetaPub.Timestamp, err)
			continue
		}

		if err := parser.Parse(); err != nil {
			logrus.Errorf("Failed to parse timestamp [%d], err [%s]", logMetaPub.Timestamp, err)
		}
	}
}

func main() {
	flag.Parse()

	err := common.LoadEnv(metadataDir)
	if err != nil {
		logrus.Fatal(err)
	}

	modeFunc, isExist := modeMap[mode]
	if !isExist {
		logrus.Fatalf("Unsupported mode [%s]", mode)
	}
	modeFunc()
}
