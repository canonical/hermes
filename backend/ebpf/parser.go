package ebpf

import (
	"fmt"

	memory "github.com/yukariatlas/hermes/backend/ebpf/memory_alloc"
)

type Parser interface {
	Parse(path string) error
}

func GetParser(ebpfType EbpfType) (Parser, error) {
	switch ebpfType {
	case Memory:
		return memory.GetParser()
	}

	return nil, fmt.Errorf("Unahndled ebpf type [%d]", ebpfType)
}
