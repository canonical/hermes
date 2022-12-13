package ebpf

import (
	"context"
	"fmt"

	memory "github.com/yukariatlas/hermes/backend/ebpf/memory_alloc"
)

type EbpfType uint32

const (
	Memory EbpfType = iota
)

type Loader interface {
	Load(context context.Context) error
	StoreData(outputPath string) error
	GetOutputFiles() []string
	Close()
}

func GetLoader(ebpfType EbpfType) (Loader, error) {
	switch ebpfType {
	case Memory:
		return memory.GetLoader()
	}

	return nil, fmt.Errorf("Unahndled ebpf type [%d]", ebpfType)
}
