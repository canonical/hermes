package collector

import (
	"context"
	"fmt"
	"hermes/log"
	"hermes/storage"
	"sync"
	"time"

	"hermes/common"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type JobRunner struct {
	configDir      string
	logDir         string
	storEngine     storage.StorEngine
	jobCompleteSub chan log.LogMetaPubFormat
	quit           chan bool
	jobsInProcess  sync.Map
}

func NewJobRunner(configDir, logDir, storEngine string, jobCompleteSub chan log.LogMetaPubFormat) (*JobRunner, error) {
	err := log.PrepareLogDataDir(logDir)
	if err != nil {
		return nil, err
	}

	storEngineInst, err := storage.GetStorEngine(storEngine, logDir)
	if err != nil {
		return nil, err
	}

	return &JobRunner{
		configDir:      configDir,
		logDir:         logDir,
		storEngine:     storEngineInst,
		jobCompleteSub: jobCompleteSub,
		quit:           make(chan bool),
		jobsInProcess:  sync.Map{}}, nil
}

func (runner *JobRunner) newJob(job Job) {
	logMeta := log.LogMetadata{
		JobName:   job.Name,
		DataLabel: uuid.NewString(),
	}
	timestamp := time.Now().Unix()
	runner.jobsInProcess.Store(job.Name, timestamp)
	defer runner.jobsInProcess.Delete(job.Name)

	routineName := job.Start

	for routineName != "" {
		routine, isExist := job.Routines[routineName]
		if !isExist {
			logrus.Errorf("Routine [%s] does not exist", routineName)
			return
		}

		task, err := NewTask(runner.configDir, routine)
		if err != nil {
			logrus.Errorf(err.Error())
			return
		}

		if task.Cond.Type == common.None && task.Task.Type == common.None {
			break
		}

		if task.Cond.Type != common.None {
			err = task.Condition(runner.logDir, logMeta.DataLabel)
			logMeta.AddMetadata(log.Metadata{
				TaskType:       int(task.Cond.Type),
				LogDataPostfix: task.GetCondLogDataPathPostfix(),
			})
			if err != nil {
				routineName = routine.CondFail
				continue
			}
		}

		if task.Task.Type != common.None {
			errChan := make(chan error)
			task.Process(runner.logDir, logMeta.DataLabel, errChan)
			select {
			case <-runner.quit:
				return
			case err := <-errChan:
				if err != nil {
					logrus.Errorf("Task [%s] failed, err [%s].", routineName, err)
					return
				}
				logMeta.AddMetadata(log.Metadata{
					TaskType:       int(task.Task.Type),
					LogDataPostfix: task.GetTaskLogDataPathPostfix(),
				})
			}
		}
		routineName = routine.CondSucc
	}

	if err := runner.storEngine.Save(timestamp, logMeta); err != nil {
		logrus.Errorf("Failed to save log metadata, err [%s]", err)
	}

	if runner.jobCompleteSub != nil {
		runner.jobCompleteSub <- log.LogMetaPubFormat{
			Timestamp:   timestamp,
			LogMetadata: logMeta,
		}
	}
}

func (runner *JobRunner) Add(job Job) error {
	if _, isExist := runner.jobsInProcess.Load(job.Name); isExist {
		return fmt.Errorf("Job [%s] is still processing", job.Name)
	}

	go runner.newJob(job)
	return nil
}

func (runner *JobRunner) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			runner.quit <- true
		}
	}
}

func (runner *JobRunner) Run(ctx context.Context) {
	go runner.run(ctx)
}
