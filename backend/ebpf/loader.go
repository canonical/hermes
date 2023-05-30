package ebpf

import (
	"context"
	"fmt"

	"hermes/common"
	"hermes/log"

	memory "hermes/backend/ebpf/memory_alloc"
)

type Loader interface {
	GetLogDataPathPostfix() string
	Load(context context.Context) error
	StoreData(logDataPathGenerator log.LogDataPathGenerator) error
	Close()
}

func GetLoader(taskType common.TaskType) (Loader, error) {
	switch taskType {
	case common.MemoryEbpf:
		return memory.GetLoader()
	}

	return nil, fmt.Errorf("Unahndled task type [%d]", taskType)
}
