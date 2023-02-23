package collector

import (
	"time"

	"hermes/backend/ftrace"
	"hermes/parser"
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

func NewTaskTraceInstance() (TaskInstance, error) {
	ftrace, err := ftrace.NewFtrace()
	if err != nil {
		return nil, err
	}

	return &TaskTraceInstance{
		ftrace: ftrace}, nil
}

func (instance *TaskTraceInstance) Process(instContext interface{}, outputPath string, result chan TaskResult) {
	traceContext := instContext.(*TraceContext)
	taskResult := TaskResult{
		Err:         nil,
		ParserType:  parser.None,
		OutputFiles: []string{},
	}
	var err error
	defer func() {
		taskResult.Err = err
		result <- taskResult
	}()

	err = instance.ftrace.Enable(traceContext.CurrentTracer,
		traceContext.TraceOptions, traceContext.SetEvent, traceContext.SetFtraceFilter)
	if err != nil {
		return
	}
	defer func() { instance.ftrace.Disable() }()

	timeout := make(chan bool)
	ack := make(chan error)

	go instance.ftrace.Trace(outputPath, timeout, ack)

	timer := time.NewTimer(time.Duration(traceContext.Timeout) * time.Second)
	defer timer.Stop()
	select {
	case err = <-ack:
		return
	case <-timer.C:
		timeout <- true
	}
}
