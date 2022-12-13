package collector

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/yukariatlas/hermes/parser"
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

type TaskResult struct {
	Err         error
	ParserType  parser.ParserType
	OutputFiles []string
}

type TaskInstance interface {
	Process(param string, param_oevrride string, outputPath string, result chan TaskResult)
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

func (task *Task) execute(context *TaskContext, outputPath string, result chan TaskResult) {
	instance, err := task.getInstance(context.Type)
	if err != nil {
		result <- TaskResult{
			Err:         err,
			ParserType:  parser.None,
			OutputFiles: []string{},
		}
		return
	}

	instance.Process(context.Param, context.ParamOverride, outputPath, result)
}

func (task *Task) Condition(outputPath string) TaskResult {
	if task.Cond.Type == None {
		return TaskResult{
			Err:         nil,
			ParserType:  parser.None,
			OutputFiles: []string{},
		}
	}

	result := make(chan TaskResult)
	go task.execute(&task.Cond, outputPath, result)
	return <-result
}

func (task *Task) Process(outputPath string, result chan TaskResult) {
	if task.Task.Type == None {
		result <- TaskResult{
			Err:         nil,
			ParserType:  parser.None,
			OutputFiles: []string{},
		}
		return
	}
	go task.execute(&task.Task, outputPath, result)
}
