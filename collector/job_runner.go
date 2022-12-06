package collector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type JobRunner struct {
	outputDir     string
	quit          chan bool
	jobsInProcess sync.Map
}

func NewJobRunner() (*JobRunner, error) {
	return &JobRunner{
		outputDir:     "",
		quit:          make(chan bool),
		jobsInProcess: sync.Map{}}, nil
}

func (runner *JobRunner) prepareOutputDir(timestamp int64, jobName string) (string, error) {
	outputDir := filepath.Join(runner.outputDir, strconv.FormatInt(timestamp, 10), jobName)
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	return outputDir, nil
}

func (runner *JobRunner) newJob(job Job) {
	timestamp := time.Now().Unix()
	runner.jobsInProcess.Store(job.Name, timestamp)
	defer runner.jobsInProcess.Delete(job.Name)

	routineName := job.Start

	outputDir, err := runner.prepareOutputDir(timestamp, job.Name)
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	for routineName != "" {
		outputPath := filepath.Join(outputDir, routineName)
		routine, isExist := job.Routines[routineName]
		if !isExist {
			logrus.Errorf("Routine [%s] does not exist", routineName)
			return
		}

		task, err := NewTask(routine)
		if err != nil {
			logrus.Errorf(err.Error())
			return
		}

		if err := task.Condition(outputPath); err != nil {
			routineName = routine.CondFail
			continue
		}

		finish := make(chan error)
		task.Process(outputPath, finish)
		select {
		case <-runner.quit:
			return
		case err := <-finish:
			if err != nil {
				logrus.Errorf("Task [%s] failed, err [%s].", routineName, err.Error())
				return
			}
		}
		routineName = routine.CondSucc
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

func (runner *JobRunner) Run(ctx context.Context, outputDir string) {
	runner.outputDir = outputDir
	go runner.run(ctx)
}
