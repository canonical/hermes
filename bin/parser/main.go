package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/yukariatlas/hermes/parser"
)

var (
	dir string
)

func init() {
	flag.StringVar(&dir, "dir", "/var/log/collector", "The path of log directory")
	flag.Usage = usage
}

func usage() {
	fmt.Println("Usage: parser [dir]")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	matches, err := filepath.Glob(dir + string("/*/*/") + parser.MetadataFilename)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, file := range matches {
		parser, err := parser.NewParser(file)
		if err != nil {
			logrus.Errorf("Failed to generate parser for metadata [%s], err [%s]", file, err)
			continue
		}

		if err := parser.Parse(); err != nil {
			logrus.Errorf("Failed to parse metadata [%s], err [%s]", file, err)
		}
	}
}
