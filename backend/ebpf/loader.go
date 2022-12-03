package ebpf

import (
	"context"
	"fmt"

	memory "github.com/yukariatlas/hermes/backend/ebpf/memory_alloc"
)

type LoaderType uint32

const (
	Memory LoaderType = iota
)

type Loader interface {
	Load(context context.Context) error
	StoreData(outputPath string) error
	Close()
}

func GetLoader(loaderType LoaderType) (Loader, error) {
	switch loaderType {
	case Memory:
		return memory.GetLoader()
	}

	return nil, fmt.Errorf("Unahndled loader type [%d]", loaderType)
}
