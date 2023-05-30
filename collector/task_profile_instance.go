package collector

import (
	"context"
	"strconv"
	"sync"
	"time"

	"hermes/backend/perf"
	"hermes/backend/utils"
	"hermes/common"
	"hermes/log"

	"github.com/sirupsen/logrus"
)

type ProfileContext struct {
	Timeout uint32
}

type TaskProfileInstance struct{}

func NewTaskProfileInstance(_ common.TaskType) (TaskInstance, error) {
	return &TaskProfileInstance{}, nil
}

func (instance *TaskProfileInstance) GetLogDataPathPostfix() string {
	return ".cpu_*"
}

func (instance *TaskProfileInstance) profile(ctx context.Context, cpu int, attr *perf.Attr, logDataPath string) {
	perfEvent, err := perf.NewPerfEvent(attr, perf.AllThreads, cpu)
	if err != nil {
		logrus.Error(err)
		return
	}

	perfEvent.Profile(ctx, logDataPath)
}

func (instance *TaskProfileInstance) Process(instContext interface{}, logDataPathGenerator log.LogDataPathGenerator, result chan error) {
	profileContext := instContext.(*ProfileContext)
	var err error
	defer func() {
		result <- err
	}()

	attr := perf.Attr{
		SampleFormat: perf.SampleFormat{
			IP:        true,
			Callchain: true,
		},
	}
	perf.TaskClock.SetAttr(&attr)
	attr.SetSamplePeriod(1000)
	attr.SetWakeupEvents(1)

	var waitGroup sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(profileContext.Timeout)*time.Second)
	defer cancel()
	for cpu := 0; cpu < utils.GetCpuNum(); cpu++ {
		waitGroup.Add(1)
		logDataPath := logDataPathGenerator(".cpu_" + strconv.Itoa(cpu))
		go func(cpu int, path string) {
			defer waitGroup.Done()
			instance.profile(ctx, cpu, &attr, path)
		}(cpu, logDataPath)
	}

	waitGroup.Wait()
}
