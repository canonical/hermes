package common

type TaskType uint32

const (
	None TaskType = iota
	Binary
	Trace
	Profile
	MemoryEbpf
	PSI
	CpuInfo
	MemoryInfo
)

const PSITask = "psi"
const TraceTask = "trace"
const BinaryTask = "binary"
const ProfileTask = "profile"
const CpuInfoTask = "cpu_info"
const MemoryInfoTask = "memory_info"
const MemoryEbpfTask = "memory_ebpf"

type Context interface {
	Check() error
}

func TaskNameToType(taskName string) TaskType {
	mapper := map[string]TaskType{
		BinaryTask:     Binary,
		TraceTask:      Trace,
		ProfileTask:    Profile,
		MemoryEbpfTask: MemoryEbpf,
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

func TaskTypeToParserCategory(taskType TaskType) string {
	switch taskType {
	case CpuInfo, Profile:
		return "cpu_profile"
	case MemoryInfo, MemoryEbpf:
		return "memory_ebpf"
	}
	return ""
}
