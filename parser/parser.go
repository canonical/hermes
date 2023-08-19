package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"hermes/common"
	"hermes/log"
)

const (
	CpuProfileJob     = "cpu_profile"
	MemleakProfileJob = "memleak_profile"
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

func (parser *Parser) getTaskParser(jobName string, taskType common.TaskType) (ParserInstance, error) {
	parserGetMapping := map[string]map[common.TaskType]func() (ParserInstance, error){
		CpuProfileJob: {
			common.CpuInfo: GetCpuInfoParser,
			common.Profile: GetCpuProfileParser,
		},
		MemleakProfileJob: {
			common.MemoryInfo: GetMemoryInfoParser,
			common.Ebpf:       GetMemoryAllocEbpfParser,
		},
	}

	taskMapping, isExist := parserGetMapping[jobName]
	if !isExist {
		return nil, fmt.Errorf("Unhandled job name [%s]", jobName)
	}

	getParser, isExist := taskMapping[taskType]
	if !isExist {
		return nil, fmt.Errorf("Unhandled task type [%d]", taskType)
	}
	return getParser()
}

func (parser *Parser) Parse() error {
	for _, meta := range parser.logMeta.Metadatas {
		logDataPathGenerator := log.GetLogDataPathGenerator(parser.logDir, parser.logMeta.LogDataLabel)
		instance, err := parser.getTaskParser(parser.logMeta.JobName, common.TaskType(meta.TaskType))
		if err != nil {
			return err
		}

		if instance == nil {
			return nil
		}

		outputDir := filepath.Join(parser.outputDir, parser.logMeta.JobName)
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return err
		}
		if err := instance.Parse(logDataPathGenerator, parser.timestamp, meta.LogDataPostfix, outputDir); err != nil {
			return err
		}
	}
	return nil
}
