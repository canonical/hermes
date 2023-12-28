package collector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"hermes/backend/dbgsym"
	"hermes/backend/perf"
	"hermes/backend/utils"
	"hermes/common"
	"hermes/log"

	"github.com/sirupsen/logrus"
)

const (
	SampleFreq   = "freq"
	SamplePeriod = "period"
)

type ProfileContext struct {
	Timeout      uint32 `yaml:"timeout"`
	SamplingType string `yaml:"sampling_type"`
	Sampling     uint64 `yaml:"sampling"`
}

func (context *ProfileContext) check() error {
	if context.Timeout == 0 {
		return fmt.Errorf("The timeout cannot be zero")
	}
	if !common.Contains([]string{SampleFreq, SamplePeriod}, context.SamplingType) {
		return fmt.Errorf("Unrecognized sampling type [%s]", context.SamplingType)
	}
	if context.Sampling == 0 {
		return fmt.Errorf("The sampling cannot be zero")
	}
	return nil
}

func (context *ProfileContext) Fill(param, paramOverride *[]byte) error {
	if err := common.FillContext(param, paramOverride, context); err != nil {
		return err
	}
	return context.check()
}

type TaskProfileInstance struct{}

func NewTaskProfileInstance(_ common.TaskType) (TaskInstance, error) {
	return &TaskProfileInstance{}, nil
}

func (instance *TaskProfileInstance) GetLogDataPathPostfix(instContext interface{}) string {
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

func (instance *TaskProfileInstance) Process(instContext interface{}, logPathManager log.LogPathManager, result chan error) {
	profileContext := instContext.(*ProfileContext)
	var err error
	defer func() {
		result <- err
	}()

	attr := perf.Attr{
		SampleFormat: perf.SampleFormat{
			Tid:       true,
			Callchain: true,
		},
		Options: perf.Options{
			Comm: true,
			Mmap: true,
		},
	}
	perf.TaskClock.SetAttr(&attr)
	if profileContext.SamplingType == SampleFreq {
		attr.SetSampleFreq(profileContext.Sampling)
	} else {
		attr.SetSamplePeriod(profileContext.Sampling)
	}
	attr.SetWakeupEvents(1)

	if synthesizeEvents, err := perf.NewSynthesizeEvents(logPathManager.DataPath(".synth_events")); err != nil {
		logrus.Errorf("Failed to generate object for synthesizing events, err [%s]", err)
	} else if err := synthesizeEvents.Synthesize(); err != nil {
		logrus.Errorf("Failed to synthesize events, err [%s]", err)
	}
	if buildIDPath, err := dbgsym.NewBuildID(dbgsym.KernelMode, "", logPathManager.DbgsymPath()).Get(); err != nil {
		logrus.Errorf("Failed to get kernel's build ID, err [%s]", err)
	} else {
		kernSymPath := logPathManager.DataPath(".kern_sym")
		if relPath, err := filepath.Rel(filepath.Dir(kernSymPath), buildIDPath); err != nil {
			logrus.Errorf("Failed to get a relative path of [%s], [%s], err [%s]", kernSymPath, buildIDPath, err)
		} else if err := os.Symlink(relPath, kernSymPath); err != nil {
			logrus.Errorf("Failed to create a symlink [%s], target [%s], err [%s]", kernSymPath, relPath, err)
		}
	}

	var waitGroup sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(profileContext.Timeout)*time.Second)
	defer cancel()
	for cpu := 0; cpu < utils.GetCpuNum(); cpu++ {
		waitGroup.Add(1)
		logDataPath := logPathManager.DataPath(".cpu_" + strconv.Itoa(cpu))
		go func(cpu int, path string) {
			defer waitGroup.Done()
			instance.profile(ctx, cpu, &attr, path)
		}(cpu, logDataPath)
	}

	waitGroup.Wait()
}
