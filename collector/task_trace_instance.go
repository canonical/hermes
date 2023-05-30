package collector

import (
	"time"

	"hermes/backend/ftrace"
	"hermes/common"
	"hermes/log"
)

type TraceContext struct {
	Timeout         uint32
	CurrentTracer   string
	TraceOptions    []string
	SetEvent        []string
	SetFtraceFilter []string
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

func (instance *TaskTraceInstance) GetLogDataPathPostfix() string {
	return ".trace"
}

func (instance *TaskTraceInstance) Process(instContext interface{}, logDataPathGenerator log.LogDataPathGenerator, result chan error) {
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

	go instance.ftrace.Trace(logDataPathGenerator(".trace"), timeout, ack)

	timer := time.NewTimer(time.Duration(traceContext.Timeout) * time.Second)
	defer timer.Stop()
	select {
	case err = <-ack:
		return
	case <-timer.C:
		timeout <- true
	}
}
