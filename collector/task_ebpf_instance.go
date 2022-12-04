package collector

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/cilium/ebpf/rlimit"
	"github.com/sirupsen/logrus"
	"github.com/yukariatlas/hermes/backend/ebpf"
)

type EbpfContext struct {
	Timeout    uint32          `json:"timeout"`
	LoaderType ebpf.LoaderType `json:"loader_type"`
}

type TaskEbpfInstance struct{}

func NewTaskEbpfInstance() (TaskInstance, error) {
	return &TaskEbpfInstance{}, nil
}

func (instance *TaskEbpfInstance) Execute(content string, outputPath string, finish chan error) {
	ebpfContext := EbpfContext{}
	err := errors.New("")
	defer func() { finish <- err }()

	err = json.Unmarshal([]byte(content), &ebpfContext)
	if err != nil {
		logrus.Errorf("Failed to unmarshal json, content [%s]", content)
		return
	}

	loader, err := ebpf.GetLoader(ebpfContext.LoaderType)
	if err != nil {
		return
	}

	var waitGroup sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ebpfContext.Timeout)*time.Second)
	defer cancel()

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
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
	}()

	waitGroup.Wait()
}
