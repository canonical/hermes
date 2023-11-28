package collector

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"hermes/common"
	"hermes/log"

	"gopkg.in/yaml.v2"
)

const (
	TaskTypeKey = "task_type"
	TasksDir    = "tasks"
)

var InstGetMapping = map[common.TaskType]func(common.TaskType) (TaskInstance, error){
	common.Binary:     NewTaskBinaryInstance,
	common.Trace:      NewTaskTraceInstance,
	common.Profile:    NewTaskProfileInstance,
	common.Ebpf:       NewTaskEbpfInstance,
	common.PSI:        NewTaskPSIInstance,
	common.CpuInfo:    NewCpuInfoInstance,
	common.MemoryInfo: NewMemoryInfoInstance,
}

type TaskContext struct {
	Type    common.TaskType
	Context common.Context
}

type TaskInstance interface {
	GetLogDataPathPostfix(context interface{}) string
	Process(context interface{}, logDataPathGenerator log.LogDataPathGenerator, result chan error)
}

type Task struct {
	Cond TaskContext
	Task TaskContext
}

func unmarshalTask(taskType string, param, paramOverride *[]byte, taskContext *TaskContext) error {
	var context common.Context
	switch taskType {
	case common.BinaryTask:
		context = &BinaryContext{}
	case common.TraceTask:
		context = &TraceContext{}
	case common.ProfileTask:
		context = &ProfileContext{}
	case common.EbpfTask:
		context = &EbpfContext{}
	case common.PSITask:
		context = &PSIContext{}
	case common.CpuInfoTask:
		context = &CpuInfoContext{}
	case common.MemoryInfoTask:
		context = &MemoryInfoContext{}
	}

	if err := context.Fill(param, paramOverride); err != nil {
		return err
	}

	taskContext.Context = context
	taskContext.Type = common.TaskNameToType(taskType)
	return nil
}

func loadTask(configDir, taskName string, paramOverride interface{}, taskContext *TaskContext) error {
	var _paramOverride *[]byte = nil
	taskConfigPath := filepath.Join(configDir, TasksDir, taskName+string(".yaml"))

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
	taskType, isExist := data[TaskTypeKey]
	if !isExist {
		return fmt.Errorf("Entry [%s] does not exist", TaskTypeKey)
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

func (task *Task) getInstance(taskType common.TaskType) (TaskInstance, error) {
	getInst, isExist := InstGetMapping[taskType]
	if !isExist {
		return nil, fmt.Errorf("Unhandled task type [%d]", taskType)
	}
	return getInst(taskType)
}

func (task *Task) execute(context *TaskContext, logDir, logDataLabel string, errChan chan error) {
	instance, err := task.getInstance(context.Type)
	if err != nil {
		errChan <- err
		return
	}

	logDataPathGenerator := log.GetLogDataPathGenerator(logDir, logDataLabel)
	instance.Process(context.Context, logDataPathGenerator, errChan)
}

func (task *Task) GetCondLogDataPathPostfix() string {
	instance, err := task.getInstance(task.Cond.Type)
	if err != nil {
		return "*"
	}
	return ".cond" + instance.GetLogDataPathPostfix(task.Cond.Context)
}

func (task *Task) Condition(logDir, logDataLabel string) error {
	if task.Cond.Type == common.None {
		return nil
	}

	errChan := make(chan error)
	go task.execute(&task.Cond, logDir, logDataLabel+".cond", errChan)
	return <-errChan
}

func (task *Task) GetTaskLogDataPathPostfix() string {
	instance, err := task.getInstance(task.Task.Type)
	if err != nil {
		return "*"
	}
	return ".task" + instance.GetLogDataPathPostfix(task.Task.Context)
}

func (task *Task) Process(logDir, logDataLabel string, errChan chan error) {
	var err error
	if task.Task.Type == common.None {
		go func() {
			errChan <- err
		}()
		return
	}
	go task.execute(&task.Task, logDir, logDataLabel+".task", errChan)
}
