package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/yukariatlas/hermes/parser"
)

var (
	logDir    string
	outputDir string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal(err)
	}

	flag.StringVar(&logDir, "log_dir", "/var/log/collector", "The path of log directory")
	flag.StringVar(&outputDir, "output_dir", homeDir+string("/view"), "The path of view directory")
	flag.Usage = usage
}

func usage() {
	fmt.Println("Usage: parser [log_dir] [output_dir]")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	matches, err := filepath.Glob(logDir + string("/*/*/") + parser.MetadataFilename)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, file := range matches {
		parser, err := parser.NewParser(file, outputDir)
		if err != nil {
			logrus.Errorf("Failed to generate parser for metadata [%s], err [%s]", file, err)
			continue
		}

		if err := parser.Parse(); err != nil {
			logrus.Errorf("Failed to parse metadata [%s], err [%s]", file, err)
		}
	}
}
