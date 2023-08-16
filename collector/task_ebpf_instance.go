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
	Timeout uint32
}

func (context *EbpfContext) check() error {
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
	taskType common.TaskType
}

func NewTaskEbpfInstance(taskType common.TaskType) (TaskInstance, error) {
	return &TaskEbpfInstance{
		taskType: taskType,
	}, nil
}

func (instance *TaskEbpfInstance) GetLogDataPathPostfix() string {
	loader, err := ebpf.GetLoader(instance.taskType)
	if err != nil {
		return ""
	}
	return loader.GetLogDataPathPostfix()
}

func (instance *TaskEbpfInstance) Process(instContext interface{}, logDataPathGenerator log.LogDataPathGenerator, result chan error) {
	ebpfContext := instContext.(*EbpfContext)
	var err error
	defer func() {
		result <- err
	}()

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		return
	}

	loader, err := ebpf.GetLoader(instance.taskType)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ebpfContext.Timeout)*time.Second)
	defer cancel()

	if err := loader.Load(ctx); err != nil {
		return
	}
	if err := loader.StoreData(logDataPathGenerator); err != nil {
		return
	}
	loader.Close()
}
