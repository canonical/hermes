package collector

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type TaskType uint32

const (
	None TaskType = iota
	Binary
	Trace
	Profile
	Ebpf
	PSI
)

func (taskType TaskType) String() string {
	return [...]string{"Binary", "Trace", "Profile", "Ebpf", "PSI"}[taskType]
}

const taskConfigDir = "/root/config/tasks/"

type TaskContext struct {
	Type          TaskType `yaml:"type"`
	Param         string   `yaml:"param"`
	ParamOverride string
}

type TaskInstance interface {
	Process(param string, param_oevrride string, outputPath string, finish chan error)
}

type Task struct {
	Cond TaskContext
	Task TaskContext
}

func loadTask(taskName string, taskContext *TaskContext) error {
	taskConfigPath := string(taskConfigDir) + taskName + string(".yaml")
	if _, err := os.Stat(taskConfigPath); err != nil {
		return err
	}

	data, err := ioutil.ReadFile(taskConfigPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, taskContext)
	if err != nil {
		return err
	}
	return nil
}

func NewTask(routine Routine) (*Task, error) {
	var task Task
	if routine.Cond.Task != "" {
		if err := loadTask(routine.Cond.Task, &task.Cond); err != nil {
			return nil, err
		}
		task.Cond.ParamOverride = routine.Cond.ParamOverride
	}

	if err := loadTask(routine.Task.Task, &task.Task); err != nil {
		return nil, err
	}
	task.Task.ParamOverride = routine.Task.ParamOverride

	return &task, nil
}

func (task *Task) getInstance(taskType TaskType) (TaskInstance, error) {
	switch taskType {
	case Binary:
		return NewTaskBinaryInstance()
	case Trace:
		return NewTaskTraceInstance()
	case Profile:
		return NewTaskProfileInstance()
	case Ebpf:
		return NewTaskEbpfInstance()
	case PSI:
		return NewTaskPSIInstance()
	}

	return nil, fmt.Errorf("Unhandled task type [%d]", taskType)
}

func (task *Task) execute(context *TaskContext, outputPath string, finish chan error) {
	instance, err := task.getInstance(context.Type)
	if err != nil {
		finish <- err
		return
	}

	instance.Process(context.Param, context.ParamOverride, outputPath, finish)
}

func (task *Task) Condition(outputPath string) error {
	if task.Cond.Type == None {
		return nil
	}

	finish := make(chan error)
	go task.execute(&task.Cond, outputPath, finish)
	return <-finish
}

func (task *Task) Process(outputPath string, finish chan error) {
	if task.Task.Type == None {
		finish <- nil
		return
	}
	go task.execute(&task.Task, outputPath, finish)
}
