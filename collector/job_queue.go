package collector

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type JobClass uint8

const (
	Disposable JobClass = iota
	Periodic
)

func (jobClass JobClass) String() string {
	return [...]string{"Disposable", "Periodic"}[jobClass]
}

type RoutineTask struct {
	Task          string `yaml:"task"`
	ParamOverride string `yaml:"param_override"`
}

type Routine struct {
	Cond     RoutineTask `yaml:"condition"`
	Task     RoutineTask `yaml:"content"`
	CondSucc string      `yaml:"condSucc"`
	CondFail string      `yaml:"condFail"`
}

type Job struct {
	Name       string
	Class      JobClass           `yaml:"class"`
	Interval   uint32             `yaml:"interval"`
	AptInstall []string           `yaml:"aptInstall"`
	Routines   map[string]Routine `yaml:"routines"`
	Start      string             `yaml:"start"`
}

func getFileNameWithoutExtension(configPath string) string {
	l := strings.LastIndexByte(configPath, '/') + 1
	if r := strings.LastIndexByte(configPath, '.'); r != -1 {
		return configPath[l:r]
	}
	return configPath[l:]
}

func NewJob(configPath string) (*Job, error) {
	var job Job

	job.Name = getFileNameWithoutExtension(configPath)
	if _, err := os.Stat(configPath); err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

type JobQueueComm struct {
	AddJob    chan Job
	RemoveJob chan string
	Ack       chan error
}

type JobQueue struct {
	Comm      JobQueueComm
	jobProtos map[string]Job
	ticker    *JobTicker
	runner    *JobRunner
}

func NewJobQueue() (*JobQueue, error) {
	ticker, err := NewJobTicker()
	if err != nil {
		return nil, err
	}

	runner, err := NewJobRunner()
	if err != nil {
		return nil, err
	}

	return &JobQueue{
		Comm: JobQueueComm{
			AddJob:    make(chan Job),
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
		return errors.New("Duplicated job name")
	}
	jobQueue.jobProtos[job.Name] = job

	if err := jobQueue.initJob(&job); err != nil {
		return err
	}

	jobQueue.ticker.AddJob(job)
	return nil
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
		case jobName := <-jobQueue.Comm.RemoveJob:
			jobQueue.Comm.Ack <- jobQueue.removeJob(jobName)
		case jobName := <-jobQueue.ticker.ReadyJobs:
			jobQueue.handleJobInstance(jobName)
		}
	}
}

func (jobQueue *JobQueue) Run(ctx context.Context, outputDir string) {
	jobQueue.ticker.Run(ctx)
	jobQueue.runner.Run(ctx, outputDir)
	go jobQueue.run(ctx)
}
