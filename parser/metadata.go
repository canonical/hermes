package parser

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const MetadataFilename = "metadata"

type ParserType uint32

const (
	None ParserType = iota
	CpuInfo
	CpuProfile
	MemoryInfo
	MemoryAllocEbpf
)

type Metadata struct {
	Type ParserType `yaml:"parser_type"`
	Logs []string   `yaml:"logs"`
}

type LogMetadata struct {
	Metas []Metadata `yaml:"meta"`
}

func GetParser(parserType ParserType) (ParserInstance, error) {
	parserGetMapping := map[ParserType]func() (ParserInstance, error){
		CpuInfo:         GetCpuInfoParser,
		MemoryInfo:      GetMemoryInfoParser,
		CpuProfile:      GetCpuProfileParser,
		MemoryAllocEbpf: GetMemoryAllocEbpfParser,
	}

	getParser, isExist := parserGetMapping[parserType]
	if !isExist {
		return nil, fmt.Errorf("Unhandled parser type [%d]", parserType)
	}
	return getParser()
}

func (logMeta *LogMetadata) ToFile(dir string) error {
	bytes, err := yaml.Marshal(*logMeta)
	if err != nil {
		return err
	}
	fp, err := os.OpenFile(dir+string("/")+MetadataFilename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString(string(bytes)); err != nil {
		return err
	}
	return nil
}
