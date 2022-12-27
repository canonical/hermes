package collector

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"hermes/parser"

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
	MemoryInfo
)

func (taskType TaskType) String() string {
	return [...]string{"Binary", "Trace", "Profile", "Ebpf", "PSI", "MemoryInfo"}[taskType]
}

const taskConfigDir = "/root/config/tasks/"

type TaskContext struct {
	Type    TaskType
	Context interface{}
}

type TaskResult struct {
	Err         error
	ParserType  parser.ParserType
	OutputFiles []string
}

type TaskInstance interface {
	Process(context interface{}, outputPath string, result chan TaskResult)
}

type Task struct {
	Cond TaskContext
	Task TaskContext
}

func unmarshalTask(taskName string, param, paramOverride *[]byte, taskContext *TaskContext) error {
	switch taskName {
	case MemoryInfoTask:
		var context MemoryInfoContext
		if err := yaml.Unmarshal(*param, &context); err != nil {
			return err
		}
		if paramOverride != nil {
			if err := yaml.Unmarshal(*paramOverride, &context); err != nil {
				return err
			}
		}
		taskContext.Type = MemoryInfo
		taskContext.Context = context
	case CPUProfileTask:
		var context ProfileContext
		if err := yaml.Unmarshal(*param, &context); err != nil {
			return err
		}
		if paramOverride != nil {
			if err := yaml.Unmarshal(*paramOverride, &context); err != nil {
				return err
			}
		}
		taskContext.Type = Profile
		taskContext.Context = context
	case PSITask:
		var context PSIContext
		if err := yaml.Unmarshal(*param, &context); err != nil {
			return err
		}
		if paramOverride != nil {
			if err := yaml.Unmarshal(*paramOverride, &context); err != nil {
				return err
			}
		}
		taskContext.Type = PSI
		taskContext.Context = context
	case MemoryEbpfTask:
		var context EbpfContext
		if err := yaml.Unmarshal(*param, &context); err != nil {
			return err
		}
		if paramOverride != nil {
			if err := yaml.Unmarshal(*paramOverride, &context); err != nil {
				return err
			}
		}
		taskContext.Type = Ebpf
		taskContext.Context = context
	}

	return nil
}

func loadTask(taskName string, paramOverride interface{}, taskContext *TaskContext) error {
	taskConfigPath := string(taskConfigDir) + taskName + string(".yaml")
	if _, err := os.Stat(taskConfigPath); err != nil {
		return err
	}

	param, err := ioutil.ReadFile(taskConfigPath)
	if err != nil {
		return err
	}

	if paramOverride != nil {
		_paramOverride, err := yaml.Marshal(paramOverride)
		if err != nil {
			return err
		}
		return unmarshalTask(taskName, &param, &_paramOverride, taskContext)
	}
	return unmarshalTask(taskName, &param, nil, taskContext)
}

func NewTask(routine Routine) (*Task, error) {
	var task Task
	if len(routine.Cond) == 1 {
		taskName := reflect.ValueOf(routine.Cond).MapKeys()[0].Interface().(string)
		if err := loadTask(taskName, routine.Cond[taskName], &task.Cond); err != nil {
			return nil, err
		}
	}

	taskName := reflect.ValueOf(routine.Task).MapKeys()[0].Interface().(string)
	if err := loadTask(taskName, routine.Task[taskName], &task.Task); err != nil {
		return nil, err
	}

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
	case MemoryInfo:
		return NewMemoryInfoInstance()
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

	instance.Process(context.Context, outputPath, result)
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
