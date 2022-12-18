package collector

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"hermes/backend/ebpf"
	"hermes/parser"

	"github.com/cilium/ebpf/rlimit"
	"github.com/sirupsen/logrus"
)

type EbpfContext struct {
	Timeout  uint32        `json:"timeout"`
	EbpfType ebpf.EbpfType `json:"ebpf_type"`
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

func (instance *TaskEbpfInstance) Process(param, paramOverride, outputPath string, result chan TaskResult) {
	ebpfContext := EbpfContext{}
	taskResult := TaskResult{
		Err:         nil,
		ParserType:  parser.None,
		OutputFiles: []string{},
	}
	err := errors.New("")
	defer func() {
		taskResult.Err = err
		result <- taskResult
	}()

	err = json.Unmarshal([]byte(param), &ebpfContext)
	if err != nil {
		logrus.Errorf("Failed to unmarshal json, param [%s]", param)
		return
	}
	if paramOverride != "" {
		err = json.Unmarshal([]byte(paramOverride), &ebpfContext)
		if err != nil {
			logrus.Errorf("Failed to unmarshal json, paramOverride [%s]", paramOverride)
			return
		}
	}

	loader, err := ebpf.GetLoader(ebpfContext.EbpfType)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ebpfContext.Timeout)*time.Second)
	defer cancel()

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		return
	}
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
