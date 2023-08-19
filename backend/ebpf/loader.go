package ebpf

import (
	"context"
	"fmt"

	"hermes/log"

	memory "hermes/backend/ebpf/memory_alloc"
)

const (
	MemoryEbpf = "memory"
)

type Loader interface {
	GetLogDataPathPostfix() string
	Load(context context.Context) error
	StoreData(logDataPathGenerator log.LogDataPathGenerator) error
	Close()
}

func GetLoader(ebpfType string) (Loader, error) {
	switch ebpfType {
	case MemoryEbpf:
		return memory.GetLoader()
	}
	return nil, fmt.Errorf("Unahndled ebpf type [%s]", ebpfType)
}
