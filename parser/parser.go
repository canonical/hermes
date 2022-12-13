package parser

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type ParserInstance interface {
	Parse(dir string, logs []string) error
}

type Parser struct {
	dir     string
	logMeta LogMetadata
}

func NewParser(metaPath string) (*Parser, error) {
	data, err := ioutil.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var parser Parser
	if err := yaml.Unmarshal(data, &parser.logMeta); err != nil {
		return nil, err
	}

	parser.dir = filepath.Dir(metaPath)
	return &parser, nil
}

func (parser *Parser) getInstance(meta Metadata) (ParserInstance, error) {
	switch meta.Type {
	case None:
		return nil, nil
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

		if err := instance.Parse(parser.dir, meta.Logs); err != nil {
			return err
		}
	}

	return nil
}
