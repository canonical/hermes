package collector

import (
	"context"
	"time"

	"hermes/backend/ebpf"
	"hermes/parser"

	"github.com/cilium/ebpf/rlimit"
)

const EbpfTask = "ebpf"

type EbpfContext struct {
	Timeout  uint32
	EbpfType ebpf.EbpfType
}

type TaskEbpfInstance struct{}

func NewTaskEbpfInstance() (TaskInstance, error) {
	return &TaskEbpfInstance{}, nil
}

func (instance *TaskEbpfInstance) getParserType(ebpfType ebpf.EbpfType) parser.ParserType {
	switch ebpfType {
	case ebpf.Memory:
		return parser.MemoryEbpf
	}
	return parser.None
}

func (instance *TaskEbpfInstance) Process(instContext interface{}, outputPath string, result chan TaskResult) {
	ebpfContext := instContext.(*EbpfContext)
	taskResult := TaskResult{
		Err:         nil,
		ParserType:  parser.None,
		OutputFiles: []string{},
	}
	var err error
	defer func() {
		taskResult.Err = err
		result <- taskResult
	}()

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		return
	}

	loader, err := ebpf.GetLoader(ebpfContext.EbpfType)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ebpfContext.Timeout)*time.Second)
	defer cancel()

	if err := loader.Load(ctx); err != nil {
		return
	}
	if err := loader.StoreData(outputPath); err != nil {
		return
	}
	loader.Close()

	taskResult.ParserType = instance.getParserType(ebpfContext.EbpfType)
	taskResult.OutputFiles = loader.GetOutputFiles()
}
