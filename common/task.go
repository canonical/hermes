package common

import (
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

const (
	PSITask        = "psi"
	TraceTask      = "trace"
	BinaryTask     = "binary"
	ProfileTask    = "profile"
	CpuInfoTask    = "cpu_info"
	MemoryInfoTask = "memory_info"
	EbpfTask       = "ebpf"
)

type Context interface {
	Fill(param, paramOverride *[]byte) error
}

func FillContext(param, paramOverride *[]byte, context Context) error {
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
	return nil
}

func TaskNameToType(taskName string) TaskType {
	mapper := map[string]TaskType{
		BinaryTask:     Binary,
		TraceTask:      Trace,
		ProfileTask:    Profile,
		EbpfTask:       Ebpf,
		PSITask:        PSI,
		CpuInfoTask:    CpuInfo,
		MemoryInfoTask: MemoryInfo,
	}

	taskType, isExist := mapper[taskName]
	if !isExist {
		return None
	}
	return taskType
}
