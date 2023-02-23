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
	CpuInfo
	MemoryInfo
)

func (taskType TaskType) String() string {
	return [...]string{
		"Binary", "Trace", "Profile",
		"Ebpf", "PSI", "CpuInfo", "MemoryInfo",
	}[taskType]
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

var taskTypeMapper = map[string]TaskType{
	CpuInfoTask:    CpuInfo,
	MemoryInfoTask: MemoryInfo,
	CPUProfileTask: Profile,
	PSITask:        PSI,
	MemoryEbpfTask: Ebpf,
}

type Task struct {
	Cond TaskContext
	Task TaskContext
}

func unmarshalTask(taskName string, param, paramOverride *[]byte, taskContext *TaskContext) error {
	taskType, isExist := taskTypeMapper[taskName]
	if !isExist {
		return fmt.Errorf("Task name [%s] doesn't define a task type", taskName)
	}

	var context interface{}
	switch taskType {
	case CpuInfo:
		context = &CpuInfoContext{}
	case MemoryInfo:
		context = &MemoryInfoContext{}
	case Profile:
		context = &ProfileContext{}
	case PSI:
		context = &PSIContext{}
	case Ebpf:
		context = &EbpfContext{}
	}

	if err := yaml.Unmarshal(*param, context); err != nil {
		return err
	}
	if paramOverride != nil {
		if err := yaml.Unmarshal(*paramOverride, context); err != nil {
			return err
		}
	}
	taskContext.Context = context
	taskContext.Type = taskType
	return nil
}

func loadTask(taskName string, paramOverride interface{}, taskContext *TaskContext) error {
	taskConfigPath := string(taskConfigDir) + taskName + string(".yaml")
	if _, err := os.Stat(taskConfigPath); err != nil {
		return err
	}

	bytes, err := ioutil.ReadFile(taskConfigPath)
	if err != nil {
		return err
	}

	if paramOverride != nil {
		_paramOverride, err := yaml.Marshal(paramOverride)
		if err != nil {
			return err
		}
		return unmarshalTask(taskName, &bytes, &_paramOverride, taskContext)
	}
	return unmarshalTask(taskName, &bytes, nil, taskContext)
}

func NewTask(routine Routine) (*Task, error) {
	var task Task
	if len(routine.Cond) == 1 {
		taskName := reflect.ValueOf(routine.Cond).MapKeys()[0].Interface().(string)
		if err := loadTask(taskName, routine.Cond[taskName], &task.Cond); err != nil {
			return nil, err
		}
	}

	if len(routine.Task) == 1 {
		taskName := reflect.ValueOf(routine.Task).MapKeys()[0].Interface().(string)
		if err := loadTask(taskName, routine.Task[taskName], &task.Task); err != nil {
			return nil, err
		}
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
	case CpuInfo:
		return NewCpuInfoInstance()
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
		go func() {
			result <- TaskResult{
				Err:         nil,
				ParserType:  parser.None,
				OutputFiles: []string{},
			}
		}()
		return
	}
	go task.execute(&task.Task, outputPath, result)
}
