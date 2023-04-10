package collector

import (
	"fmt"
	"hermes/parser"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v2"
)

type TaskType uint32

const taskTypeKey = "task_type"
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

func parseTaskType(val string) TaskType {
	mapper := map[string]TaskType{
		BinaryTask:     Binary,
		TraceTask:      Trace,
		ProfileTask:    Profile,
		EbpfTask:       Ebpf,
		PSITask:        PSI,
		CpuInfoTask:    CpuInfo,
		MemoryInfoTask: MemoryInfo,
	}

	taskType, isExist := mapper[val]
	if !isExist {
		return None
	}
	return taskType
}

const tasksDir = "tasks"

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
	GetParserType(context interface{}) parser.ParserType
	Process(context interface{}, outputPath string, result chan TaskResult)
}

type Task struct {
	Cond TaskContext
	Task TaskContext
}

func unmarshalTask(taskType string, param, paramOverride *[]byte, taskContext *TaskContext) error {
	var context interface{}
	switch taskType {
	case BinaryTask:
		context = &BinaryContext{}
	case TraceTask:
		context = &TraceContext{}
	case ProfileTask:
		context = &ProfileContext{}
	case EbpfTask:
		context = &EbpfContext{}
	case PSITask:
		context = &PSIContext{}
	case CpuInfoTask:
		context = &CpuInfoContext{}
	case MemoryInfoTask:
		context = &MemoryInfoContext{}
	}

	if param != nil {
		if err := yaml.Unmarshal(*param, context); err != nil {
			return err
		}
	}
	if paramOverride != nil {
		if err := yaml.Unmarshal(*paramOverride, context); err != nil {
			return err
		}
	}
	taskContext.Context = context
	taskContext.Type = parseTaskType(taskType)
	return nil
}

func loadTask(configDir, taskName string, paramOverride interface{}, taskContext *TaskContext) error {
	var _paramOverride *[]byte = nil
	taskConfigPath := filepath.Join(configDir, tasksDir, taskName+string(".yaml"))

	if paramOverride != nil {
		val, err := yaml.Marshal(paramOverride)
		if err != nil {
			return err
		}
		_paramOverride = &val
	}

	bytes, err := ioutil.ReadFile(taskConfigPath)
	if os.IsNotExist(err) {
		return unmarshalTask(taskName, nil, _paramOverride, taskContext)
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(bytes, &data); err != nil {
		return err
	}
	taskType, isExist := data[taskTypeKey]
	if !isExist {
		return fmt.Errorf("Entry [%s] does not exist", taskTypeKey)
	}
	return unmarshalTask(taskType.(string), &bytes, _paramOverride, taskContext)
}

func NewTask(configDir string, routine Routine) (*Task, error) {
	var task Task
	if len(routine.Cond) == 1 {
		taskName := reflect.ValueOf(routine.Cond).MapKeys()[0].Interface().(string)
		if err := loadTask(configDir, taskName, routine.Cond[taskName], &task.Cond); err != nil {
			return nil, err
		}
	}

	if len(routine.Task) == 1 {
		taskName := reflect.ValueOf(routine.Task).MapKeys()[0].Interface().(string)
		if err := loadTask(configDir, taskName, routine.Task[taskName], &task.Task); err != nil {
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
		result <- TaskResult{}
		return
	}

	instance.Process(context.Context, outputPath, result)
}

func (task *Task) getParserType(context *TaskContext) parser.ParserType {
	instance, err := task.getInstance(context.Type)
	if err != nil {
		return parser.None
	}
	return instance.GetParserType(context.Context)
}

func (task *Task) Condition(outputPath string) TaskResult {
	if task.Cond.Type == None {
		return TaskResult{}
	}

	result := make(chan TaskResult)
	go task.execute(&task.Cond, outputPath, result)
	return <-result
}

func (task *Task) GetCondParserType() parser.ParserType {
	if task.Cond.Type == None {
		return parser.None
	}
	return task.getParserType(&task.Cond)
}

func (task *Task) Process(outputPath string, result chan TaskResult) {
	if task.Task.Type == None {
		go func() {
			result <- TaskResult{}
		}()
		return
	}
	go task.execute(&task.Task, outputPath, result)
}

func (task *Task) GetTaskParserType() parser.ParserType {
	if task.Task.Type == None {
		return parser.None
	}
	return task.getParserType(&task.Task)
}
