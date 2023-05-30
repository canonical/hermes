package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"hermes/common"
	"hermes/log"
)

type ParserInstance interface {
	Parse(logDataPathGenerator log.LogDataPathGenerator, timestamp int64, logDataPostfix, outputDir string) error
}

type Parser struct {
	logDir    string
	outputDir string
	timestamp int64
	logMeta   log.LogMetadata
}

func NewParser(logDir, outputDir string, timestamp int64, logMeta log.LogMetadata) (*Parser, error) {
	return &Parser{
		logDir:    logDir,
		outputDir: outputDir,
		timestamp: timestamp,
		logMeta:   logMeta,
	}, nil
}

func (parser *Parser) GetTaskParser(taskType common.TaskType) (ParserInstance, error) {
	parserGetMapping := map[common.TaskType]func() (ParserInstance, error){
		common.CpuInfo:    GetCpuInfoParser,
		common.MemoryInfo: GetMemoryInfoParser,
		common.Profile:    GetCpuProfileParser,
		common.MemoryEbpf: GetMemoryAllocEbpfParser,
	}

	getParser, isExist := parserGetMapping[taskType]
	if !isExist {
		return nil, fmt.Errorf("Unhandled task type [%d]", taskType)
	}
	return getParser()
}

func (parser *Parser) Parse() error {
	for _, meta := range parser.logMeta.Metadatas {
		logDataPathGenerator := log.GetLogDataPathGenerator(parser.logDir, parser.logMeta.LogDataLabel)
		instance, err := parser.GetTaskParser(common.TaskType(meta.TaskType))
		if err != nil {
			return err
		}

		if instance == nil {
			return nil
		}

		category := common.TaskTypeToParserCategory(common.TaskType(meta.TaskType))
		if category == "" {
			return fmt.Errorf("Failed to parse task type [%d] to task name", meta.TaskType)
		}

		outputDir := filepath.Join(parser.outputDir, category)
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return err
		}
		if err := instance.Parse(logDataPathGenerator, parser.timestamp, meta.LogDataPostfix, outputDir); err != nil {
			return err
		}
	}
	return nil
}
