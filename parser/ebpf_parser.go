package parser

import (
	"fmt"

	memory "github.com/yukariatlas/hermes/backend/ebpf/memory_alloc"
)

func GetEbpfParser(parserType ParserType) (ParserInstance, error) {
	switch parserType {
	case MemoryEbpf:
		return memory.GetMemoryEbpfParser()
	}

	return nil, fmt.Errorf("Unahndled parser type [%d]", parserType)
}
