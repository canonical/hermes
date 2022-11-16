package ftrace

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	traceFsDir      = "/sys/kernel/tracing"
	currentTracer   = "current_tracer"
	traceOptions    = "trace_options"
	setEvent        = "set_event"
	setFtraceFilter = "set_ftrace_filter"
	tracingOn       = "tracing_on"
	tracePipe       = "trace_pipe"
)

type Ftrace struct{}

func NewFtrace() (*Ftrace, error) {
	return &Ftrace{}, nil
}

func (ftrace *Ftrace) writeEntry(entry string, data string) error {
	return os.WriteFile(filepath.Join(traceFsDir, entry), []byte(data), 0755)
}

func (ftrace *Ftrace) Enable(tracer string, options []string, events []string, funcs []string) error {
	if tracer != "" {
		if err := ftrace.writeEntry(currentTracer, tracer); err != nil {
			return err
		}
	}

	if len(options) > 0 {
		if err := ftrace.writeEntry(traceOptions, strings.Join(options, " ")); err != nil {
			return err
		}
	}

	if len(events) > 0 {
		if err := ftrace.writeEntry(setEvent, strings.Join(events, " ")); err != nil {
			return err
		}
	}

	if len(funcs) > 0 {
		if err := ftrace.writeEntry(setFtraceFilter, strings.Join(funcs, " ")); err != nil {
			return err
		}
	}

	return nil
}

func (ftrace *Ftrace) redirectTracePipe(outputPath string) (chan bool, error) {
	fp, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}

	pipe, err := os.Open(filepath.Join(traceFsDir, "trace_pipe"))
	if err != nil {
		return nil, err
	}

	eof := make(chan bool)
	go func() {
		defer func() {
			pipe.Close()
			fp.Close()
		}()
		scanner := bufio.NewScanner(pipe)

		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			if _, err := fp.WriteString(scanner.Text() + "\n"); err != nil {
				logrus.Error(err.Error())
				break
			}
		}
		eof <- true
	}()

	return eof, nil
}

func (ftrace *Ftrace) tracingOn(isOn bool) error {
	val := "0"
	if isOn {
		val = "1"
	}

	return ftrace.writeEntry(tracingOn, val)
}

func (ftrace *Ftrace) Trace(outputPath string, timeout chan bool, ack chan error) {
	err := errors.New("")
	defer func() { ack <- err }()

	err = ftrace.tracingOn(true)
	if err != nil {
		return
	}
	defer ftrace.tracingOn(false)

	eof, err := ftrace.redirectTracePipe(outputPath)
	if err != nil {
		return
	}

	for {
		select {
		case <-timeout:
			err = ftrace.tracingOn(false)
			if err != nil {
				return
			}
		case <-eof:
			return
		}
	}
}

func (ftrace *Ftrace) Disable() error {
	err := ftrace.tracingOn(false)

	if _err := ftrace.writeEntry(currentTracer, "nop"); err != nil {
		err = _err
	}

	if _err := ftrace.writeEntry(traceOptions, ""); err != nil {
		err = _err
	}

	if _err := ftrace.writeEntry(setEvent, ""); err != nil {
		err = _err
	}

	if _err := ftrace.writeEntry(setFtraceFilter, ""); err != nil {
		err = _err
	}

	return err
}
