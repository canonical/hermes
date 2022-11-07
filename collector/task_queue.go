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

type TaskClass uint8

const (
	Disposable TaskClass = iota
	Periodic
)

func (taskClass TaskClass) String() string {
	return [...]string{"Disposable", "Periodic"}[taskClass]
}

type TaskType uint32

const (
	Binary TaskType = iota
	Trace
)

func (taskType TaskType) String() string {
	return [...]string{"Binary", "Trace"}[taskType]
}

type SubTask struct {
	Type        TaskType `yaml:"type"`
	PreCheck    string   `yaml:"preCheck"`
	Content     string   `yaml:"content"`
	PostCheck   string   `yaml:"postCheck"`
	SuccessCond string   `yaml:"successCond"`
	FailCond    string   `yaml:"failCond"`
}

type Task struct {
	Name       string
	Class      TaskClass          `yaml:"class"`
	Interval   uint32             `yaml:"interval"`
	AptInstall []string           `yaml:"aptInstall"`
	SubTasks   map[string]SubTask `yaml:"subTasks"`
	Start      string             `yaml:"start"`
}

func getFileNameWithoutExtension(configPath string) string {
	l := strings.LastIndexByte(configPath, '/') + 1
	if r := strings.LastIndexByte(configPath, '.'); r != -1 {
		return configPath[l:r]
	}
	return configPath[l:]
}

func NewTask(configPath string) (*Task, error) {
	var task Task

	task.Name = getFileNameWithoutExtension(configPath)
	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(data, &task)
		if err != nil {
			return nil, err
		}
	}

	return &task, nil
}

type TaskQueueComm struct {
	AddTask    chan Task
	RemoveTask chan string
	Ack        chan error
}

type TaskQueue struct {
	Comm       TaskQueueComm
	taskProtos map[string]Task
	ticker     *TaskTicker
	runner     *TaskRunner
}

func NewTaskQueue() (*TaskQueue, error) {
	ticker, err := NewTaskTicker()
	if err != nil {
		return nil, err
	}

	runner, err := NewTaskRunner()
	if err != nil {
		return nil, err
	}

	return &TaskQueue{
		Comm: TaskQueueComm{
			AddTask:    make(chan Task),
			RemoveTask: make(chan string),
			Ack:        make(chan error)},
		taskProtos: make(map[string]Task),
		ticker:     ticker,
		runner:     runner}, nil
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

func (taskQueue *TaskQueue) initTask(task *Task) error {
	return aptInstall(task.AptInstall)
}

func (taskQueue *TaskQueue) addTask(task Task) error {
	if _, isExist := taskQueue.taskProtos[task.Name]; isExist {
		return errors.New("Duplicated task name")
	}
	taskQueue.taskProtos[task.Name] = task

	if err := taskQueue.initTask(&task); err != nil {
		return err
	}

	taskQueue.ticker.AddTask(task)
	return nil
}

func (taskQueue *TaskQueue) removeTask(taskName string) error {
	if _, isExist := taskQueue.taskProtos[taskName]; !isExist {
		return fmt.Errorf("Task [%s] does not exist", taskName)
	}

	delete(taskQueue.taskProtos, taskName)
	taskQueue.ticker.RemoveTask(taskName)
	return nil
}

func (taskQueue *TaskQueue) handleTaskInstance(taskName string) {
	if _, isExist := taskQueue.taskProtos[taskName]; !isExist {
		logrus.Errorf("Task [%s] does not exist, cannot trigger.", taskName)
		return
	}

	err := taskQueue.runner.Add(taskQueue.taskProtos[taskName])
	if err != nil {
		logrus.Errorf(err.Error())
	}
}

func (taskQueue *TaskQueue) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-taskQueue.Comm.AddTask:
			taskQueue.Comm.Ack <- taskQueue.addTask(task)
		case taskName := <-taskQueue.Comm.RemoveTask:
			taskQueue.Comm.Ack <- taskQueue.removeTask(taskName)
		case taskName := <-taskQueue.ticker.ReadyTasks:
			taskQueue.handleTaskInstance(taskName)
		}
	}
}

func (taskQueue *TaskQueue) Run(ctx context.Context, outputDir string) {
	taskQueue.ticker.Run(ctx)
	taskQueue.runner.Run(ctx, outputDir)
	go taskQueue.run(ctx)
}
