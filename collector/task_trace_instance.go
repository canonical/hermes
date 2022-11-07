package collector

import (
	"encoding/json"
	"errors"
	"time"
)

type TraceContext struct {
	Timeout         uint32   `json:"timeout"`
	CurrentTracer   string   `json:"currentTracer"`
	TraceOptions    []string `json:"traceOptions"`
	SetEvent        []string `json:"setEvent"`
	SetFtraceFilter []string `json:"setFtraceFilter"`
}

type TaskTraceInstance struct {
	ftrace *Ftrace
}

func NewTaskTraceInstance() (TaskInstance, error) {
	ftrace, err := NewFtrace()
	if err != nil {
		return nil, err
	}

	return &TaskTraceInstance{
		ftrace: ftrace}, nil
}

func (instance *TaskTraceInstance) Execute(content string, outputPath string, finish chan error) {
	context := TraceContext{}
	err := errors.New("")
	defer func() { finish <- err }()

	err = json.Unmarshal([]byte(content), &context)
	if err != nil {
		return
	}

	err = instance.ftrace.Enable(context.CurrentTracer, context.TraceOptions, context.SetEvent, context.SetFtraceFilter)
	if err != nil {
		return
	}
	defer func() { instance.ftrace.Disable() }()

	timeout := make(chan bool)
	ack := make(chan error)

	go instance.ftrace.Trace(outputPath, timeout, ack)

	timer := time.NewTimer(time.Duration(context.Timeout) * time.Second)
	defer timer.Stop()
	select {
	case err = <-ack:
		return
	case <-timer.C:
		timeout <- true
	}
}
