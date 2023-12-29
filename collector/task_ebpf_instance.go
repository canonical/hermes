package collector

import (
	"context"
	"fmt"
	"time"

	"hermes/backend/ebpf"
	"hermes/common"
	"hermes/log"

	"github.com/cilium/ebpf/rlimit"
)

type EbpfContext struct {
	EbpfType string `yaml:"ebpf_type"`
	Timeout  uint32 `yaml:"timeout"`
}

func (context *EbpfContext) check() error {
	if context.EbpfType == "" {
		return fmt.Errorf("The ebpf type is empty")
	}
	if context.Timeout == 0 {
		return fmt.Errorf("The timeout cannot be zero")
	}
	return nil
}

func (context *EbpfContext) Fill(param, paramOverride *[]byte) error {
	if err := common.FillContext(param, paramOverride, context); err != nil {
		return err
	}
	return context.check()
}

type TaskEbpfInstance struct {
	ebpfType string
	taskType common.TaskType
}

func NewTaskEbpfInstance(taskType common.TaskType) (TaskInstance, error) {
	return &TaskEbpfInstance{
		taskType: taskType,
	}, nil
}

func (instance *TaskEbpfInstance) GetLogDataPathPostfix(instContext interface{}) string {
	ebpfContext := instContext.(*EbpfContext)
	loader, err := ebpf.GetLoader(ebpfContext.EbpfType)
	if err != nil {
		return ""
	}
	return loader.GetLogDataPathPostfix()
}

func (instance *TaskEbpfInstance) Process(instContext interface{}, logPathManager log.LogPathManager, result chan error) {
	ebpfContext := instContext.(*EbpfContext)
	var loader ebpf.Loader
	var err error
	defer func() {
		result <- err
	}()

	instance.ebpfType = ebpfContext.EbpfType

	// Allow the current process to lock memory for eBPF resources.
	if err = rlimit.RemoveMemlock(); err != nil {
		return
	}

	loader, err = ebpf.GetLoader(instance.ebpfType)
	if err != nil {
		return
	}

	if err = loader.Prepare(logPathManager); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ebpfContext.Timeout)*time.Second)
	defer cancel()

	if err = loader.Load(ctx); err != nil {
		return
	}
	if err = loader.StoreData(logPathManager); err != nil {
		return
	}
	loader.Close()
}
