package parser

import (
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
	MemoryEbpf
)

type Metadata struct {
	Type ParserType `yaml:"type"`
	Logs []string   `yaml:"logs"`
}

type LogMetadata struct {
	Metas []Metadata `yaml:"meta"`
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
