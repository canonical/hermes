package collector

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"hermes/backend/perf"
	"hermes/parser"

	"github.com/sirupsen/logrus"
)

type ProfileContext struct {
	Timeout uint32 `json:"timeout"`
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

func (instance *TaskProfileInstance) Process(param, paramOverride, outputPath string, result chan TaskResult) {
	profileContext := ProfileContext{}
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

	err = json.Unmarshal([]byte(param), &profileContext)
	if err != nil {
		logrus.Errorf("Failed to unmarshal json, param [%s]", param)
		return
	}
	if paramOverride != "" {
		err = json.Unmarshal([]byte(paramOverride), &profileContext)
		if err != nil {
			logrus.Errorf("Failed to unmarshal json, paramOverride [%s]", paramOverride)
			return
		}
	}

	attr := perf.Attr{
		SampleFormat: perf.SampleFormat{
			IP:        true,
			Callchain: true,
		},
	}
	perf.TaskClock.SetAttr(&attr)
	attr.SetSamplePeriod(1)
	attr.SetWakeupEvents(1)

	var waitGroup sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(profileContext.Timeout)*time.Second)
	defer cancel()
	for cpu := 0; cpu < runtime.NumCPU(); cpu++ {
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
