package collector

import (
	"fmt"
	"time"

	"hermes/backend/ftrace"
	"hermes/common"
	"hermes/log"
)

type TraceContext struct {
	Timeout         uint32   `yaml:"timeout"`
	CurrentTracer   string   `yaml:"current_tracer"`
	TraceOptions    []string `yaml:"trace_options"`
	SetEvent        []string `yaml:"set_event"`
	SetFtraceFilter []string `yaml:"set_ftrace_filter"`
}

func (context *TraceContext) check() error {
	if context.Timeout == 0 {
		return fmt.Errorf("The timeout cannot be zero")
	}
	return nil
}

func (context *TraceContext) Fill(param, paramOverride *[]byte) error {
	if err := common.FillContext(param, paramOverride, context); err != nil {
		return err
	}
	return context.check()
}

type TaskTraceInstance struct {
	ftrace *ftrace.Ftrace
}

func NewTaskTraceInstance(_ common.TaskType) (TaskInstance, error) {
	ftrace, err := ftrace.NewFtrace()
	if err != nil {
		return nil, err
	}

	return &TaskTraceInstance{
		ftrace: ftrace}, nil
}

func (instance *TaskTraceInstance) GetLogDataPathPostfix(instContext interface{}) string {
	return ".trace"
}

func (instance *TaskTraceInstance) Process(instContext interface{}, logPathManager log.LogPathManager, result chan error) {
	traceContext := instContext.(*TraceContext)
	var err error
	defer func() {
		result <- err
	}()

	err = instance.ftrace.Enable(traceContext.CurrentTracer,
		traceContext.TraceOptions, traceContext.SetEvent, traceContext.SetFtraceFilter)
	if err != nil {
		return
	}
	defer func() { instance.ftrace.Disable() }()

	timeout := make(chan bool)
	ack := make(chan error)

	go instance.ftrace.Trace(logPathManager.DataPath(".trace"), timeout, ack)

	timer := time.NewTimer(time.Duration(traceContext.Timeout) * time.Second)
	defer timer.Stop()
	select {
	case err = <-ack:
		return
	case <-timer.C:
		timeout <- true
	}
}
