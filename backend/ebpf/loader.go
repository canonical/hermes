package ebpf

import (
	"context"
	"fmt"

	"hermes/log"

	memory "hermes/backend/ebpf/memory_alloc"
	iolat "hermes/backend/ebpf/io_latency"
)

const (
	MemoryEbpf = "memory"
	IoLatEbpf = "io_latency"
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
	case IoLatEbpf:
		return iolat.GetLoader()
	}
	return nil, fmt.Errorf("Unahndled ebpf type [%s]", ebpfType)
}
