package collector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type TaskInstance interface {
	Execute(content string, outputPath string, finish chan error)
}

func getTaskInstance(taskType TaskType) (TaskInstance, error) {
	switch taskType {
	case Binary:
		return NewTaskBinaryInstance()
	case Trace:
		return NewTaskTraceInstance()
	case Profile:
		return NewTaskProfileInstance()
	}

	return nil, fmt.Errorf("Unhandled task type [%d]", taskType)
}

type TaskRunner struct {
	outputDir      string
	quit           chan bool
	tasksInProcess sync.Map
}

func NewTaskRunner() (*TaskRunner, error) {
	return &TaskRunner{
		outputDir:      "",
		quit:           make(chan bool),
		tasksInProcess: sync.Map{}}, nil
}

func (runner *TaskRunner) prepareOutputDir(timestamp int64, taskName string) (string, error) {
	outputDir := filepath.Join(runner.outputDir, strconv.FormatInt(timestamp, 10), taskName)
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	return outputDir, nil
}

func (runner *TaskRunner) check(content string, env map[string]string) bool {
	tokens := strings.Split(content, "|")
	return PrepareCmd(tokens, env).Run() == nil
}

func (runner *TaskRunner) preProcess(content string) bool {
	if len(content) == 0 {
		return true
	}
	return runner.check(content, map[string]string{})
}

func (runner *TaskRunner) postProcess(subTaskName string, outputPath string, task *Task) (string, error) {
	subTask, isExist := task.SubTasks[subTaskName]
	if !isExist {
		return "", fmt.Errorf("Subtask [%s] does not exist", subTaskName)
	}

	if len(subTask.PostCheck) == 0 {
		return "", nil
	}

	env := map[string]string{
		"OUTPUT_PATH": outputPath,
	}
	if runner.check(subTask.PostCheck, env) {
		return subTask.SuccessCond, nil
	}
	return subTask.FailCond, nil
}

func (runner *TaskRunner) newTask(task Task) {
	timestamp := time.Now().Unix()
	runner.tasksInProcess.Store(task.Name, timestamp)
	defer runner.tasksInProcess.Delete(task.Name)

	subTaskName := task.Start
	finish := make(chan error)

	outputDir, err := runner.prepareOutputDir(timestamp, task.Name)
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	for subTaskName != "" {
		outputPath := filepath.Join(outputDir, subTaskName)
		subTask, isExist := task.SubTasks[subTaskName]
		if !isExist {
			logrus.Errorf("Subtask [%s] does not exist", subTaskName)
			return
		}

		instance, err := getTaskInstance(subTask.Type)
		if err != nil {
			logrus.Errorf(err.Error())
			return
		}

		if !runner.preProcess(subTask.PreCheck) {
			continue
		}

		go instance.Execute(subTask.Content, outputPath, finish)

		select {
		case <-runner.quit:
			return
		case err := <-finish:
			if err != nil {
				logrus.Errorf("Task [%s] failed, err [%s].", subTaskName, err.Error())
				return
			}
			subTaskName, err = runner.postProcess(subTaskName, outputPath, &task)
			if err != nil {
				logrus.Errorf("Postprocess of task [%s] failed, err [%s].", subTaskName, err.Error())
				return
			}
		}
	}
}

func (runner *TaskRunner) Add(task Task) error {
	if _, isExist := runner.tasksInProcess.Load(task.Name); isExist {
		return fmt.Errorf("Task [%s] is still processing", task.Name)
	}

	go runner.newTask(task)
	return nil
}

func (runner *TaskRunner) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			runner.quit <- true
		}
	}
}

func (runner *TaskRunner) Run(ctx context.Context, outputDir string) {
	runner.outputDir = outputDir
	go runner.run(ctx)
}
