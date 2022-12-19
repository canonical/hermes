package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type ParserInstance interface {
	Parse(logDir string, logs []string, outputDir string) error
}

type Parser struct {
	logDir    string
	outputDir string
	logMeta   LogMetadata
}

func NewParser(metaPath string, outputDir string) (*Parser, error) {
	data, err := ioutil.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var parser Parser
	if err := yaml.Unmarshal(data, &parser.logMeta); err != nil {
		return nil, err
	}

	parser.logDir = filepath.Dir(metaPath)
	taskName := filepath.Base(parser.logDir)
	timestamp := filepath.Base(filepath.Dir(parser.logDir))
	parser.outputDir = outputDir + string("/") + taskName + string("/") + timestamp
	return &parser, nil
}

func (parser *Parser) getInstance(meta Metadata) (ParserInstance, error) {
	switch meta.Type {
	case None:
		return nil, nil
	case MemoryInfo:
		return GetMemoryInfoParser()
	case MemoryEbpf:
		return GetEbpfParser(meta.Type)
	}

	return nil, fmt.Errorf("Unhandled parser type [%d]", meta.Type)
}

func (parser *Parser) Parse() error {
	for _, meta := range parser.logMeta.Metas {
		instance, err := parser.getInstance(meta)
		if err != nil {
			return err
		}

		if instance == nil {
			return nil
		}

		if err := os.MkdirAll(parser.outputDir, os.ModePerm); err != nil {
			return err
		}
		if err := instance.Parse(parser.logDir, meta.Logs, parser.outputDir); err != nil {
			return err
		}
	}

	return nil
}
