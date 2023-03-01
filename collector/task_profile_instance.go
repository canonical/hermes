package collector

import (
	"context"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"hermes/backend/perf"
	"hermes/backend/utils"
	"hermes/parser"

	"github.com/sirupsen/logrus"
)

const ProfileTask = "profile"

type ProfileContext struct {
	Timeout uint32
}

type TaskProfileInstance struct{}

func NewTaskProfileInstance() (TaskInstance, error) {
	return &TaskProfileInstance{}, nil
}

func (instance *TaskProfileInstance) profile(ctx context.Context, cpu int, attr *perf.Attr, outputPath string) {
	perfEvent, err := perf.NewPerfEvent(attr, perf.AllThreads, cpu)
	if err != nil {
		logrus.Error(err)
		return
	}

	perfEvent.Profile(ctx, outputPath)
}

func (instance *TaskProfileInstance) GetParserType(instContext interface{}) parser.ParserType {
	return parser.CpuProfile
}

func (instance *TaskProfileInstance) Process(instContext interface{}, outputPath string, result chan TaskResult) {
	profileContext := instContext.(*ProfileContext)
	taskResult := TaskResult{}
	var err error
	defer func() {
		taskResult.Err = err
		result <- taskResult
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
		path := outputPath + string(".") + strconv.Itoa(cpu)
		taskResult.OutputFiles = append(taskResult.OutputFiles, filepath.Base(path))
		go func(cpu int, path string) {
			defer waitGroup.Done()
			instance.profile(ctx, cpu, &attr, path)
		}(cpu, path)
	}

	waitGroup.Wait()
}
