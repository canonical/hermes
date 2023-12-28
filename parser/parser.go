package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"hermes/common"
	"hermes/log"
)

const (
	CpuProfileJob     = "cpu_profile"
	MemleakProfileJob = "memleak_profile"
	IoLatencyJob      = "io_latency"
)

var ParserGetMapping = map[string]map[common.TaskType]func() (ParserInstance, error){
	CpuProfileJob: {
		common.CpuInfo: GetCpuInfoParser,
		common.Profile: GetCpuProfileParser,
	},
	MemleakProfileJob: {
		common.MemoryInfo: GetMemoryInfoParser,
		common.Ebpf:       GetMemoryAllocEbpfParser,
	},
	IoLatencyJob: {
		common.PSI:  GetPSIParser,
		common.Ebpf: GetIoLatEbpfParser,
	},
}

type ParserInstance interface {
	Parse(logPathManager log.LogPathManager, timestamp int64, logDataPostfix, outputDir string) error
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
	taskMapping, isExist := ParserGetMapping[jobName]
	if !isExist {
		return nil, fmt.Errorf("Unhandled job name [%s]", jobName)
	}

	getParser, isExist := taskMapping[taskType]
	if !isExist {
		logrus.Printf("parser: %+v", getParser)
		return nil, fmt.Errorf("Unhandled task type [%d]", taskType)
	}
	return getParser()
}

func (parser *Parser) Parse() error {
	for _, meta := range parser.logMeta.Metadatas {
		logPathManager := log.NewLogPathManager(parser.logDir).SetDataLabel(parser.logMeta.DataLabel)
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
		if err := instance.Parse(*logPathManager, parser.timestamp, meta.LogDataPostfix, outputDir); err != nil {
			return err
		}
	}
	return nil
}
