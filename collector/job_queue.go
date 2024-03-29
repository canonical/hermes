package collector

import (
	"context"
	"fmt"
	"hermes/log"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	Disposable = "disposable"
	Periodic   = "periodic"
	Enabled    = "enabled"
	Disabled   = "disabled"
)

type Routine struct {
	Cond     map[string]interface{} `yaml:"condition"`
	Task     map[string]interface{} `yaml:"content"`
	CondSucc string                 `yaml:"cond_succ"`
	CondFail string                 `yaml:"cond_fail"`
}

type Job struct {
	Name       string             `yaml:"-"`
	Class      string             `yaml:"class"`
	Interval   uint32             `yaml:"interval"`
	Status     string             `yaml:"status"`
	AptInstall []string           `yaml:"apt_install"`
	Routines   map[string]Routine `yaml:"routines"`
	Start      string             `yaml:"start"`
}

func getFileNameWithoutExt(configPath string) string {
	l := strings.LastIndexByte(configPath, '/') + 1
	if r := strings.LastIndexByte(configPath, '.'); r != -1 {
		return configPath[l:r]
	}
	return configPath[l:]
}

func NewJob(configPath string) (*Job, error) {
	var job Job

	job.Name = getFileNameWithoutExt(configPath)
	if _, err := os.Stat(configPath); err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(bytes, &job)
	if err != nil {
		return nil, err
	}

	if len(job.Status) == 0 || (job.Status != Enabled && job.Status != Disabled) {
		job.Status = Enabled
	}

	for _, routine := range job.Routines {
		if len(routine.Cond) > 1 || len(routine.Task) > 1 {
			return nil, fmt.Errorf("Unexpected config format")
		}
	}

	return &job, nil
}

type JobQueueComm struct {
	AddJob    chan Job
	ModifyJob chan Job
	RemoveJob chan string
	Ack       chan error
}

type JobQueue struct {
	Comm      JobQueueComm
	jobProtos map[string]Job
	ticker    *JobTicker
	runner    *JobRunner
}

func NewJobQueue(configDir, logDir, storEngine string, jobCompleteSub chan log.LogMetaPubFormat) (*JobQueue, error) {
	ticker, err := NewJobTicker()
	if err != nil {
		return nil, err
	}

	runner, err := NewJobRunner(configDir, logDir, storEngine, jobCompleteSub)
	if err != nil {
		return nil, err
	}

	return &JobQueue{
		Comm: JobQueueComm{
			AddJob:    make(chan Job),
			ModifyJob: make(chan Job),
			RemoveJob: make(chan string),
			Ack:       make(chan error)},
		jobProtos: make(map[string]Job),
		ticker:    ticker,
		runner:    runner}, nil
}

func aptInstall(pkgs []string) error {
	for _, pkg := range pkgs {
		if err := exec.Command("/usr/bin/dpkg", "-s", pkg).Run(); err == nil {
			continue
		}
		if err := exec.Command("/usr/bin/apt", "install", pkg, "-y").Run(); err != nil {
			return err
		}
	}

	return nil
}

func (jobQueue *JobQueue) initJob(job *Job) error {
	return aptInstall(job.AptInstall)
}

func (jobQueue *JobQueue) addJob(job Job) error {
	if _, isExist := jobQueue.jobProtos[job.Name]; isExist {
		if reflect.DeepEqual(job, jobQueue.jobProtos[job.Name]) {
			return nil
		}
		if err := jobQueue.removeJob(job.Name); err != nil {
			return err
		}
	}

	if job.Status == Disabled {
		return nil
	}

	jobQueue.jobProtos[job.Name] = job

	if err := jobQueue.initJob(&job); err != nil {
		return err
	}

	jobQueue.ticker.AddJob(job)
	return nil
}

func (jobQueue *JobQueue) modifyJob(job Job) error {
	if _, isExist := jobQueue.jobProtos[job.Name]; isExist {
		if err := jobQueue.removeJob(job.Name); err != nil {
			return err
		}
	}

	if job.Status == Disabled {
		return nil
	}
	return jobQueue.addJob(job)
}

func (jobQueue *JobQueue) removeJob(jobName string) error {
	if _, isExist := jobQueue.jobProtos[jobName]; !isExist {
		return fmt.Errorf("Job [%s] does not exist", jobName)
	}

	delete(jobQueue.jobProtos, jobName)
	jobQueue.ticker.RemoveJob(jobName)
	return nil
}

func (jobQueue *JobQueue) handleJobInstance(jobName string) {
	if _, isExist := jobQueue.jobProtos[jobName]; !isExist {
		logrus.Errorf("Job [%s] does not exist, cannot trigger.", jobName)
		return
	}

	err := jobQueue.runner.Add(jobQueue.jobProtos[jobName])
	if err != nil {
		logrus.Errorf(err.Error())
	}
}

func (jobQueue *JobQueue) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-jobQueue.Comm.AddJob:
			jobQueue.Comm.Ack <- jobQueue.addJob(job)
		case job := <-jobQueue.Comm.ModifyJob:
			jobQueue.Comm.Ack <- jobQueue.modifyJob(job)
		case jobName := <-jobQueue.Comm.RemoveJob:
			jobQueue.Comm.Ack <- jobQueue.removeJob(jobName)
		case jobName := <-jobQueue.ticker.ReadyJobs:
			jobQueue.handleJobInstance(jobName)
		}
	}
}

func (jobQueue *JobQueue) Run(ctx context.Context) {
	jobQueue.ticker.Run(ctx)
	jobQueue.runner.Run(ctx)
	go jobQueue.run(ctx)
}
