package perf

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

type EventType uint32

const (
	Hardware   EventType = unix.PERF_TYPE_HARDWARE
	Software   EventType = unix.PERF_TYPE_SOFTWARE
	Tracepoint EventType = unix.PERF_TYPE_TRACEPOINT
	HWCache    EventType = unix.PERF_TYPE_HW_CACHE
	Raw        EventType = unix.PERF_TYPE_RAW
	Breakpoint EventType = unix.PERF_TYPE_BREAKPOINT
)

type HardwareEvent uint64

const (
	CPUCycles             HardwareEvent = unix.PERF_COUNT_HW_CPU_CYCLES
	Instructions          HardwareEvent = unix.PERF_COUNT_HW_INSTRUCTIONS
	CacheReferences       HardwareEvent = unix.PERF_COUNT_HW_CACHE_REFERENCES
	CacheMisses           HardwareEvent = unix.PERF_COUNT_HW_CACHE_MISSES
	BranchInstructions    HardwareEvent = unix.PERF_COUNT_HW_BRANCH_INSTRUCTIONS
	BranchMisses          HardwareEvent = unix.PERF_COUNT_HW_BRANCH_MISSES
	BusCycles             HardwareEvent = unix.PERF_COUNT_HW_BUS_CYCLES
	StalledCyclesFrontend HardwareEvent = unix.PERF_COUNT_HW_STALLED_CYCLES_FRONTEND
	StalledCyclesBackend  HardwareEvent = unix.PERF_COUNT_HW_STALLED_CYCLES_BACKEND
	RefCPUCycles          HardwareEvent = unix.PERF_COUNT_HW_REF_CPU_CYCLES
)

type AttrConfigurator interface {
	Configure(attr *Attr)
}

func (event HardwareEvent) GetAttr() *Attr {
	return &Attr{
		Type:   unix.PERF_TYPE_HARDWARE,
		Config: uint64(event),
		Options: Options{
			Disabled: true,
		},
	}
}

func (event HardwareEvent) Configure(attr *Attr) {
	attr.Type = unix.PERF_TYPE_HARDWARE
	attr.Config = uint64(event)
}

type SoftwareEvent uint64

const (
	CPUClock        SoftwareEvent = unix.PERF_COUNT_SW_CPU_CLOCK
	TaskClock       SoftwareEvent = unix.PERF_COUNT_SW_TASK_CLOCK
	PageFaults      SoftwareEvent = unix.PERF_COUNT_SW_PAGE_FAULTS
	ContextSwitches SoftwareEvent = unix.PERF_COUNT_SW_CONTEXT_SWITCHES
	CPUMigrations   SoftwareEvent = unix.PERF_COUNT_SW_CPU_MIGRATIONS
	PageFaultsMin   SoftwareEvent = unix.PERF_COUNT_SW_PAGE_FAULTS_MIN
	PageFaultsMaj   SoftwareEvent = unix.PERF_COUNT_SW_PAGE_FAULTS_MAJ
	AlignmentFaults SoftwareEvent = unix.PERF_COUNT_SW_ALIGNMENT_FAULTS
	EmulationFaults SoftwareEvent = unix.PERF_COUNT_SW_EMULATION_FAULTS
	Dummy           SoftwareEvent = unix.PERF_COUNT_SW_DUMMY
	BpfOutput       SoftwareEvent = unix.PERF_COUNT_SW_BPF_OUTPUT
)

func (event SoftwareEvent) GetAttr() *Attr {
	return &Attr{
		Type:   unix.PERF_TYPE_SOFTWARE,
		Config: uint64(event),
		Options: Options{
			Disabled: true,
		},
	}
}

func (event SoftwareEvent) Configure(attr *Attr) {
	attr.Type = unix.PERF_TYPE_SOFTWARE
	attr.Config = uint64(event)
}

type SampleFormat struct {
	IP           bool
	Tid          bool
	Time         bool
	Addr         bool
	Read         bool
	Callchain    bool
	ID           bool
	CPU          bool
	Period       bool
	StreamID     bool
	Raw          bool
	BranchStack  bool
	RegsUser     bool
	StackUser    bool
	Weight       bool
	DataSrc      bool
	Identifier   bool
	Transaction  bool
	RegsIntr     bool
	PhysAddr     bool
	Aux          bool
	Cgroup       bool
	DataPageSize bool
	CodePageSize bool
	WeightStruct bool
}

func bitFieldsToUint64(bitFields []bool) uint64 {
	var val uint64

	for shift, set := range bitFields {
		if set {
			val |= (1 << uint(shift))
		}
	}

	return val
}

func (format *SampleFormat) BitFields() uint64 {
	bitFields := []bool{
		format.IP,
		format.Tid,
		format.Time,
		format.Addr,
		format.Read,
		format.Callchain,
		format.ID,
		format.CPU,
		format.Period,
		format.StreamID,
		format.Raw,
		format.BranchStack,
		format.RegsUser,
		format.StackUser,
		format.Weight,
		format.DataSrc,
		format.Identifier,
		format.Transaction,
		format.RegsIntr,
		format.PhysAddr,
		format.Aux,
		format.Cgroup,
		format.DataPageSize,
		format.CodePageSize,
		format.WeightStruct,
	}

	return bitFieldsToUint64(bitFields)
}

type ReadFormat struct {
	TotalTimeEnabled bool
	TotalTimeRunning bool
	ID               bool
	Group            bool
}

func (format *ReadFormat) BitFields() uint64 {
	bitFields := []bool{
		format.TotalTimeEnabled,
		format.TotalTimeRunning,
		format.ID,
		format.Group,
	}

	return bitFieldsToUint64(bitFields)
}

func (format *ReadFormat) CalcRequiredSize() int {
	size := 8
	if format.TotalTimeEnabled {
		size += 8
	}
	if format.TotalTimeRunning {
		size += 8
	}
	if format.ID {
		size += 8
	}
	return size
}

func (format ReadFormat) CalcGroupRequiredSize(events int) int {
	size := 8
	if format.TotalTimeEnabled {
		size += 8
	}
	if format.TotalTimeRunning {
		size += 8
	}
	valSize := 8
	if format.ID {
		valSize += 8
	}
	return size + events*valSize
}

type Options struct {
	Disabled               bool
	Inherit                bool
	Pinned                 bool
	Exclusive              bool
	ExcludeUser            bool
	ExcludeKernel          bool
	ExcludeHv              bool
	ExcludeIdle            bool
	Mmap                   bool
	Comm                   bool
	Freq                   bool
	InheritStat            bool
	EnableOnExec           bool
	Task                   bool
	Watermark              bool
	PreciseIPBit1          bool
	PreciseIPBit2          bool
	MmapData               bool
	SampleIDAll            bool
	ExcludeHost            bool
	ExcludeGuest           bool
	ExcludeCallchainKernel bool
	ExcludeCallchainUser   bool
	Mmap2                  bool
	CommExec               bool
	UseClockID             bool
	ContextWwitch          bool
	WriteBackward          bool
	Namespaces             bool
	Ksymbol                bool
	BpfEvent               bool
	AuxOutput              bool
	Cgroup                 bool
	TextPoke               bool
}

func (opt *Options) BitFields() uint64 {
	bitFields := []bool{
		opt.Disabled,
		opt.Inherit,
		opt.Pinned,
		opt.Exclusive,
		opt.ExcludeUser,
		opt.ExcludeKernel,
		opt.ExcludeHv,
		opt.ExcludeIdle,
		opt.Mmap,
		opt.Comm,
		opt.Freq,
		opt.InheritStat,
		opt.EnableOnExec,
		opt.Task,
		opt.Watermark,
		opt.PreciseIPBit1,
		opt.PreciseIPBit2,
		opt.MmapData,
		opt.SampleIDAll,
		opt.ExcludeHost,
		opt.ExcludeGuest,
		opt.ExcludeCallchainKernel,
		opt.ExcludeCallchainUser,
		opt.Mmap2,
		opt.CommExec,
		opt.UseClockID,
		opt.ContextWwitch,
		opt.WriteBackward,
		opt.Namespaces,
		opt.Ksymbol,
		opt.BpfEvent,
		opt.AuxOutput,
		opt.Cgroup,
		opt.TextPoke,
	}

	return bitFieldsToUint64(bitFields)
}

type BranchPrivilegeLevel uint64

const (
	UserLevel   BranchPrivilegeLevel = unix.PERF_SAMPLE_BRANCH_USER
	KernelLevel BranchPrivilegeLevel = unix.PERF_SAMPLE_BRANCH_KERNEL
	HVLevel     BranchPrivilegeLevel = unix.PERF_SAMPLE_BRANCH_HV
	PLMAll      BranchPrivilegeLevel = unix.PERF_SAMPLE_BRANCH_PLM_ALL
)

type BranchType uint64

const (
	Any       BranchType = unix.PERF_SAMPLE_BRANCH_ANY
	AnyCall   BranchType = unix.PERF_SAMPLE_BRANCH_ANY_CALL
	AnyReturn BranchType = unix.PERF_SAMPLE_BRANCH_ANY_RETURN
	IndCall   BranchType = unix.PERF_SAMPLE_BRANCH_IND_CALL
	AbortTx   BranchType = unix.PERF_SAMPLE_BRANCH_ABORT_TX
	InTx      BranchType = unix.PERF_SAMPLE_BRANCH_IN_TX
	NoTx      BranchType = unix.PERF_SAMPLE_BRANCH_NO_TX
	Cond      BranchType = unix.PERF_SAMPLE_BRANCH_COND
	CallStack BranchType = unix.PERF_SAMPLE_BRANCH_CALL_STACK
	IndJump   BranchType = unix.PERF_SAMPLE_BRANCH_IND_JUMP
	Call      BranchType = unix.PERF_SAMPLE_BRANCH_CALL
	NoFlags   BranchType = unix.PERF_SAMPLE_BRANCH_NO_FLAGS
	NoCycles  BranchType = unix.PERF_SAMPLE_BRANCH_NO_CYCLES
	TypeSave  BranchType = unix.PERF_SAMPLE_BRANCH_TYPE_SAVE
	HWIndex   BranchType = unix.PERF_SAMPLE_BRANCH_HW_INDEX
)

type BranchSampleType struct {
	PrivilegeLevel BranchPrivilegeLevel
	Type           BranchType
}

func (branchSampleType *BranchSampleType) BitFields() uint64 {
	return uint64(branchSampleType.PrivilegeLevel) | uint64(branchSampleType.Type)
}

type Attr struct {
	Label            string
	Type             uint32
	Config           uint64
	Sample           uint64
	SampleFormat     SampleFormat
	ReadFormat       ReadFormat
	Options          Options
	Wakeup           uint32
	BreakpointType   uint32
	Config1          uint64
	Config2          uint64
	BranchSampleType BranchSampleType
	SampleRegsUser   uint64
	SampleStackUser  uint32
	ClockID          int32
	SampleRegsIntr   uint64
	AuxWatermark     uint32
	SampleMaxStack   uint16
}

func (_attr *Attr) Configure(attr *Attr) {
	*attr = *_attr
}

func (attr *Attr) ToUnixPerfEventAttr() *unix.PerfEventAttr {
	return &unix.PerfEventAttr{
		Type:               uint32(attr.Type),
		Size:               uint32(unsafe.Sizeof(unix.PerfEventAttr{})),
		Config:             attr.Config,
		Sample:             attr.Sample,
		Sample_type:        attr.SampleFormat.BitFields(),
		Read_format:        attr.ReadFormat.BitFields(),
		Bits:               attr.Options.BitFields(),
		Wakeup:             attr.Wakeup,
		Bp_type:            attr.BreakpointType,
		Ext1:               attr.Config1,
		Ext2:               attr.Config2,
		Branch_sample_type: attr.BranchSampleType.BitFields(),
		Sample_regs_user:   attr.SampleRegsUser,
		Sample_stack_user:  attr.SampleStackUser,
		Clockid:            attr.ClockID,
		Sample_regs_intr:   attr.SampleRegsIntr,
		Aux_watermark:      attr.AuxWatermark,
		Sample_max_stack:   attr.SampleMaxStack,
	}
}

func (attr *Attr) SetSamplePeriod(period uint64) {
	attr.Sample = period
	attr.Options.Freq = false
}

func (attr *Attr) SetSampleFreq(freq uint64) {
	attr.Sample = freq
	attr.Options.Freq = true
}

func (attr *Attr) SetWakeupEvents(events uint32) {
	attr.Wakeup = events
	attr.Options.Watermark = false
}

func (attr *Attr) SetWakeupWatermark(watermark uint32) {
	attr.Wakeup = watermark
	attr.Options.Watermark = true
}
