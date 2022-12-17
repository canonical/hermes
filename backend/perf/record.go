package perf

import (
	"fmt"
	"math/bits"
	"sync/atomic"
	"unsafe"

	"hermes/backend/symbol"

	"golang.org/x/sys/unix"
)

type RecordType uint32

const (
	MmapRec          RecordType = unix.PERF_RECORD_MMAP
	LostRec          RecordType = unix.PERF_RECORD_LOST
	CommRec          RecordType = unix.PERF_RECORD_COMM
	ExitRec          RecordType = unix.PERF_RECORD_EXIT
	ThrottleRec      RecordType = unix.PERF_RECORD_THROTTLE
	UnthrottleRec    RecordType = unix.PERF_RECORD_UNTHROTTLE
	ForkRec          RecordType = unix.PERF_RECORD_FORK
	ReadRec          RecordType = unix.PERF_RECORD_READ
	SampleRec        RecordType = unix.PERF_RECORD_SAMPLE
	Mmap2Rec         RecordType = unix.PERF_RECORD_MMAP2
	AuxRec           RecordType = unix.PERF_RECORD_AUX
	ItraceStartRec   RecordType = unix.PERF_RECORD_ITRACE_START
	LostSamplesRec   RecordType = unix.PERF_RECORD_LOST_SAMPLES
	SwitchRec        RecordType = unix.PERF_RECORD_SWITCH
	SwitchCPUWideRec RecordType = unix.PERF_RECORD_SWITCH_CPU_WIDE
	NamespacesRec    RecordType = unix.PERF_RECORD_NAMESPACES
)

type Header struct {
	Type RecordType `json:"type"`
	Misc uint16     `json:"misc"`
	Size uint16     `json:"size"`
}

type RawRecord struct {
	Header Header
	Data   []byte
}

type Record interface {
	Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer)
}

type SampleID struct {
	Pid        uint32 `json:"pid"`
	Tid        uint32 `json:"tid"`
	Time       uint64 `json:"time"`
	ID         uint64 `json:"id"`
	StreamID   uint64 `json:"stream_id"`
	CPU        uint32 `json:"cpu"`
	_          uint32 `json:"res"`
	Identifier uint64 `json:"identifier"`
}

type MmapRecord struct {
	Header
	RecordType RecordType `json:"record_type"`
	Pid        uint32     `json:"pid"`
	Tid        uint32     `json:"tid"`
	Addr       uint64     `json:"addr"`
	Len        uint64     `json:"len"`
	Pgoff      uint64     `json:"pgoff"`
	Filename   string     `json:"filename"`
	SampleID
}

func (rec *MmapRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = MmapRec
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Tid)
	parser.Uint64(&rec.Addr)
	parser.Uint64(&rec.Len)
	parser.Uint64(&rec.Pgoff)
	parser.String(&rec.Filename)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type LostRecord struct {
	Header
	RecordType RecordType `json:"record_type"`
	ID         uint64     `json:"id"`
	Lost       uint64     `json:"lost"`
	SampleID
}

func (rec *LostRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = LostRec
	parser.Uint64(&rec.ID)
	parser.Uint64(&rec.Lost)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type CommRecord struct {
	Header
	RecordType RecordType `json:"record_type"`
	Pid        uint32     `json:"pid"`
	Tid        uint32     `json:"tid"`
	Comm       string     `json:"comm"`
	SampleID
}

func (rec *CommRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = CommRec
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Tid)
	parser.String(&rec.Comm)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type ExitRecord struct {
	Header
	RecordType RecordType `json:"record_type"`
	Pid        uint32     `json:"pid"`
	Ppid       uint32     `json:"ppid"`
	Tid        uint32     `json:"tid"`
	Ptid       uint32     `json:"ptid"`
	Time       uint64     `json:"time"`
	SampleID
}

func (rec *ExitRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = ExitRec
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Ppid)
	parser.Uint32(&rec.Tid)
	parser.Uint32(&rec.Ptid)
	parser.Uint64(&rec.Time)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type ThrottleRecord struct {
	Header
	RecordType RecordType `json:"record_type"`
	Time       uint64     `json:"time"`
	ID         uint64     `json:"id"`
	StreamID   uint64     `json:"stream_id"`
	SampleID
}

func (rec *ThrottleRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = ThrottleRec
	parser.Uint64(&rec.Time)
	parser.Uint64(&rec.ID)
	parser.Uint64(&rec.StreamID)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type UnthrottleRecord struct {
	Header
	RecordType RecordType `json:"record_type"`
	Time       uint64     `json:"time"`
	ID         uint64     `json:"id"`
	StreamID   uint64     `json:"stream_id"`
	SampleID
}

func (rec *UnthrottleRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = UnthrottleRec
	parser.Uint64(&rec.Time)
	parser.Uint64(&rec.ID)
	parser.Uint64(&rec.StreamID)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type ForkRecord struct {
	Header
	RecordType RecordType `json:"record_type"`
	Pid        uint32     `json:"pid"`
	Ppid       uint32     `json:"ppid"`
	Tid        uint32     `json:"tid"`
	Ptid       uint32     `json:"ptid"`
	Time       uint64     `json:"time"`
	SampleID
}

func (rec *ForkRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = ForkRec
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Ppid)
	parser.Uint32(&rec.Tid)
	parser.Uint32(&rec.Ptid)
	parser.Uint64(&rec.Time)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type ReadContentValue struct {
	Value uint64 `json:"value"`
	ID    uint64 `json:"id"`
}

type ReadContent struct {
	TimeEnabled uint64 `json:"time_enabled"`
	TimeRunning uint64 `json:"time_running"`
	ReadContentValue
}

type ReadRecord struct {
	Header
	RecordType  RecordType  `json:"record_type"`
	Pid         uint32      `json:"pid"`
	Tid         uint32      `json:"tid"`
	ReadContent ReadContent `json:"values"`
	SampleID
}

func (rec *ReadRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = ReadRec
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Tid)
	parser.ParseReadContent(attr.ReadFormat, &rec.ReadContent)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type GroupReadRecord struct {
	Header
	RecordType       RecordType `json:"record_type"`
	Pid              uint32     `json:"pid"`
	Tid              uint32     `json:"tid"`
	GroupReadContent GroupReadContent
	SampleID
}

func (rec *GroupReadRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = ReadRec
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Tid)
	parser.ParseGroupReadContent(attr.ReadFormat, &rec.GroupReadContent)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type BranchEntry struct {
	From         uint64 `json:"from"`
	To           uint64 `json:"to"`
	Mispredicted bool   `json:"mispred"`
	Predicted    bool   `json:"predicted"`
	InTrans      bool   `json:"in_tx"`
	Abort        bool   `json:"abort"`
	Cycles       uint16 `json:"cycles"`
	BranchType   uint8  `json:"branch_type"`
}

func (entry *BranchEntry) Decode(from, to, flags uint64) {
	*entry = BranchEntry{
		From:         from,
		To:           to,
		Mispredicted: (flags & (1 << 0)) != 0,
		Predicted:    (flags & (1 << 1)) != 0,
		InTrans:      (flags & (1 << 2)) != 0,
		Abort:        (flags & (1 << 3)) != 0,
		Cycles:       uint16((flags << 44) >> 48),
		BranchType:   uint8((flags << 40) >> 44),
	}
}

type Instruction struct {
	IP     uint64 `json:"ip"`
	Symbol string `json:"symbol"`
}

type SampleRecord struct {
	Header
	RecordType       RecordType    `json:"record_type"`
	Identifier       uint64        `json:"identifier"`
	Inst             Instruction   `json:"instruction"`
	Pid              uint32        `json:"pid"`
	Tid              uint32        `json:"tid"`
	Time             uint64        `json:"time"`
	Addr             uint64        `json:"addr"`
	ID               uint64        `json:"id"`
	StreamID         uint64        `json:"stream_id"`
	CPU              uint32        `json:"cpu"`
	_                uint32        `json:"res"`
	Period           uint64        `json:"period"`
	ReadContent      ReadContent   `json:"values"`
	CallchainInsts   []Instruction `json:"callchain_insts"`
	RawData          []byte        `json:"raw_data"`
	BranchStack      []BranchEntry `json:"lbr"`
	RegsUserABI      uint64        `json:"regs_user_abi"`
	RegsUserRegs     []uint64      `json:"regs_user_regs"`
	StackUserData    []byte        `josn:"stack_user_data"`
	StackUserDynSize uint64        `json:"stack_user_dyn_size"`
	WeightFull       uint64        `json:"weight_full"`
	DataSrc          uint64        `json:"data_src"`
	Transaction      uint64        `json:"transaction"`
	RegsIntrABI      uint64        `json:"regs_intr_abi"`
	RegsIntrRegs     []uint64      `json:"regs_intr_regs"`
	PhysAddr         uint64        `json:"phys_addr"`
	AuxData          []byte        `json:"aux_data"`
	DataPageSize     uint64        `json:"data_page_size"`
	CodePageSize     uint64        `json:"code_page_size"`
}

func getSymbol(symbolizer *symbol.Symbolizer, inst *Instruction) {
	if inst.IP == 0 {
		return
	}
	if symbol, err := symbolizer.Symbolize(inst.IP); err == nil {
		inst.Symbol = symbol
	}
}

func (rec *SampleRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = SampleRec
	parser.Uint64Cond(attr.SampleFormat.Identifier, &rec.Identifier)
	parser.Uint64Cond(attr.SampleFormat.IP, &rec.Inst.IP)
	getSymbol(symbolizer, &rec.Inst)
	parser.Uint32Cond(attr.SampleFormat.Tid, &rec.Pid)
	parser.Uint32Cond(attr.SampleFormat.Tid, &rec.Tid)
	parser.Uint64Cond(attr.SampleFormat.Time, &rec.Time)
	parser.Uint64Cond(attr.SampleFormat.Addr, &rec.Addr)
	parser.Uint64Cond(attr.SampleFormat.ID, &rec.ID)
	parser.Uint64Cond(attr.SampleFormat.StreamID, &rec.StreamID)

	var reserved uint32
	parser.Uint32Cond(attr.SampleFormat.CPU, &rec.CPU)
	parser.Uint32Cond(attr.SampleFormat.CPU, &reserved)
	parser.Uint64Cond(attr.SampleFormat.Period, &rec.Period)
	if attr.SampleFormat.Read {
		parser.ParseReadContent(attr.ReadFormat, &rec.ReadContent)
	}
	if attr.SampleFormat.Callchain {
		var nr uint64
		parser.Uint64(&nr)
		rec.CallchainInsts = make([]Instruction, nr)
		for i := 0; i < int(nr); i++ {
			parser.Uint64(&rec.CallchainInsts[i].IP)
			getSymbol(symbolizer, &rec.CallchainInsts[i])
		}
	}
	if attr.SampleFormat.Raw {
		parser.BytesByUint32Size(&rec.RawData)
	}
	if attr.SampleFormat.BranchStack {
		var nr uint64
		parser.Uint64(&nr)
		rec.BranchStack = make([]BranchEntry, nr)
		for i := 0; i < int(nr); i++ {
			var from, to, flags uint64
			parser.Uint64(&from)
			parser.Uint64(&to)
			parser.Uint64(&flags)
			rec.BranchStack[i].Decode(from, to, flags)
		}
	}
	if attr.SampleFormat.RegsUser {
		parser.Uint64(&rec.RegsUserABI)
		nr := bits.OnesCount64(attr.SampleRegsUser)
		rec.RegsUserRegs = make([]uint64, nr)
		for i := 0; i < nr; i++ {
			parser.Uint64(&rec.RegsUserRegs[i])
		}
	}
	if attr.SampleFormat.StackUser {
		parser.BytesByUint64Size(&rec.StackUserData)
		if len(rec.StackUserData) > 0 {
			parser.Uint64(&rec.StackUserDynSize)
		}
	}
	parser.Uint64Cond(attr.SampleFormat.Weight, &rec.WeightFull)
	parser.Uint64Cond(attr.SampleFormat.DataSrc, &rec.DataSrc)
	parser.Uint64Cond(attr.SampleFormat.Transaction, &rec.Transaction)
	if attr.SampleFormat.RegsIntr {
		parser.Uint64(&rec.RegsIntrABI)
		nr := bits.OnesCount64(attr.SampleRegsIntr)
		rec.RegsIntrRegs = make([]uint64, nr)
		for i := 0; i < int(nr); i++ {
			parser.Uint64(&rec.RegsIntrRegs[i])
		}
	}
	parser.Uint64Cond(attr.SampleFormat.PhysAddr, &rec.PhysAddr)
	if attr.SampleFormat.Aux {
		parser.BytesByUint64Size(&rec.AuxData)
	}
	parser.Uint64Cond(attr.SampleFormat.DataPageSize, &rec.DataPageSize)
	parser.Uint64Cond(attr.SampleFormat.CodePageSize, &rec.CodePageSize)
}

type GroupSampleRecord struct {
	Header
	RecordType       RecordType       `json:"record_type"`
	Identifier       uint64           `json:"identifier"`
	Inst             Instruction      `json:"instruction"`
	Pid              uint32           `json:"pid"`
	Tid              uint32           `json:"tid"`
	Time             uint64           `json:"time"`
	Addr             uint64           `json:"addr"`
	ID               uint64           `json:"id"`
	StreamID         uint64           `json:"stream_id"`
	CPU              uint32           `json:"cpu"`
	_                uint32           `json:"res"`
	Period           uint64           `json:"period"`
	GroupReadContent GroupReadContent `json:"values"`
	CallchainInsts   []Instruction    `json:"callchain_insts"`
	RawData          []byte           `json:"raw_data"`
	BranchStack      []BranchEntry    `json:"lbr"`
	RegsUserABI      uint64           `json:"regs_user_abi"`
	RegsUserRegs     []uint64         `json:"regs_user_regs"`
	StackUserData    []byte           `josn:"stack_user_data"`
	StackUserDynSize uint64           `json:"stack_user_dyn_size"`
	WeightFull       uint64           `json:"weight_full"`
	DataSrc          uint64           `json:"data_src"`
	Transaction      uint64           `json:"transaction"`
	RegsIntrABI      uint64           `json:"regs_intr_abi"`
	RegsIntrRegs     []uint64         `json:"regs_intr_regs"`
	PhysAddr         uint64           `json:"phys_addr"`
	AuxData          []byte           `json:"aux_data"`
	DataPageSize     uint64           `json:"data_page_size"`
	CodePageSize     uint64           `json:"code_page_size"`
}

func (rec *GroupSampleRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = SampleRec
	parser.Uint64Cond(attr.SampleFormat.Identifier, &rec.Identifier)
	parser.Uint64Cond(attr.SampleFormat.IP, &rec.Inst.IP)
	getSymbol(symbolizer, &rec.Inst)
	parser.Uint32Cond(attr.SampleFormat.Tid, &rec.Pid)
	parser.Uint32Cond(attr.SampleFormat.Tid, &rec.Tid)
	parser.Uint64Cond(attr.SampleFormat.Time, &rec.Time)
	parser.Uint64Cond(attr.SampleFormat.Addr, &rec.Addr)
	parser.Uint64Cond(attr.SampleFormat.ID, &rec.ID)
	parser.Uint64Cond(attr.SampleFormat.StreamID, &rec.StreamID)

	var reserved uint32
	parser.Uint32Cond(attr.SampleFormat.CPU, &rec.CPU)
	parser.Uint32Cond(attr.SampleFormat.CPU, &reserved)
	parser.Uint64Cond(attr.SampleFormat.Period, &rec.Period)
	if attr.SampleFormat.Read {
		parser.ParseGroupReadContent(attr.ReadFormat, &rec.GroupReadContent)
	}
	if attr.SampleFormat.Callchain {
		var nr uint64
		parser.Uint64(&nr)
		rec.CallchainInsts = make([]Instruction, nr)
		for i := 0; i < int(nr); i++ {
			parser.Uint64(&rec.CallchainInsts[i].IP)
			getSymbol(symbolizer, &rec.CallchainInsts[i])
		}
	}
	if attr.SampleFormat.Raw {
		parser.BytesByUint32Size(&rec.RawData)
	}
	if attr.SampleFormat.BranchStack {
		var nr uint64
		parser.Uint64(&nr)
		rec.BranchStack = make([]BranchEntry, nr)
		for i := 0; i < int(nr); i++ {
			var from, to, flags uint64
			parser.Uint64(&from)
			parser.Uint64(&to)
			parser.Uint64(&flags)
			rec.BranchStack[i].Decode(from, to, flags)
		}
	}
	if attr.SampleFormat.RegsUser {
		parser.Uint64(&rec.RegsUserABI)
		nr := bits.OnesCount64(attr.SampleRegsUser)
		rec.RegsUserRegs = make([]uint64, nr)
		for i := 0; i < nr; i++ {
			parser.Uint64(&rec.RegsUserRegs[i])
		}
	}
	if attr.SampleFormat.StackUser {
		parser.BytesByUint64Size(&rec.StackUserData)
		if len(rec.StackUserData) > 0 {
			parser.Uint64(&rec.StackUserDynSize)
		}
	}
	parser.Uint64Cond(attr.SampleFormat.Weight, &rec.WeightFull)
	parser.Uint64Cond(attr.SampleFormat.DataSrc, &rec.DataSrc)
	parser.Uint64Cond(attr.SampleFormat.Transaction, &rec.Transaction)
	if attr.SampleFormat.RegsIntr {
		parser.Uint64(&rec.RegsIntrABI)
		nr := bits.OnesCount64(attr.SampleRegsIntr)
		rec.RegsIntrRegs = make([]uint64, nr)
		for i := 0; i < int(nr); i++ {
			parser.Uint64(&rec.RegsIntrRegs[i])
		}
	}
	parser.Uint64Cond(attr.SampleFormat.PhysAddr, &rec.PhysAddr)
	if attr.SampleFormat.Aux {
		parser.BytesByUint64Size(&rec.AuxData)
	}
	parser.Uint64Cond(attr.SampleFormat.DataPageSize, &rec.DataPageSize)
	parser.Uint64Cond(attr.SampleFormat.CodePageSize, &rec.CodePageSize)
}

type Mmap2Record struct {
	Header
	RecordType    RecordType `json:"record_type"`
	Pid           uint32     `json:"pid"`
	Tid           uint32     `json:"tid"`
	Addr          uint64     `json:"addr"`
	Len           uint64     `json:"len"`
	Pgoff         uint64     `json:"pgoff"`
	MajorID       uint32     `json:"maj"`
	MinorID       uint32     `json:"min"`
	Ino           uint64     `json:"ino"`
	InoGeneration uint64     `json:"ino_generation"`
	Prot          uint32     `json:"prot"`
	Flags         uint32     `json:"flags"`
	Filename      string     `json:"filename"`
	SampleID
}

func (rec *Mmap2Record) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = Mmap2Rec
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Tid)
	parser.Uint64(&rec.Addr)
	parser.Uint64(&rec.Len)
	parser.Uint64(&rec.Pgoff)
	parser.Uint32(&rec.MajorID)
	parser.Uint32(&rec.MinorID)
	parser.Uint64(&rec.Ino)
	parser.Uint64(&rec.InoGeneration)
	parser.Uint32(&rec.Prot)
	parser.Uint32(&rec.Flags)
	parser.String(&rec.Filename)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type AuxRecord struct {
	Header
	RecordType RecordType `json:"record_type"`
	Offset     uint64     `json:"aux_offset"`
	Size       uint64     `json:"aux_size"`
	Flags      uint64     `json:"flags"`
	SampleID
}

func (rec *AuxRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = AuxRec
	parser.Uint64(&rec.Offset)
	parser.Uint64(&rec.Size)
	parser.Uint64(&rec.Flags)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type ItraceStartRecord struct {
	Header
	RecordType RecordType `json:"record_type"`
	Pid        uint32     `json:"pid"`
	Tid        uint32     `json:"tid"`
	SampleID
}

func (rec *ItraceStartRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = ItraceStartRec
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Tid)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type LostSamplesRecord struct {
	Header
	RecordType RecordType `json:"record_type"`
	Lost       uint64     `json:"lost"`
	SampleID
}

func (rec *LostSamplesRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = LostSamplesRec
	parser.Uint64(&rec.Lost)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type SwitchRecord struct {
	Header
	RecordType RecordType `json:"record_type"`
	SampleID
}

func (rec *SwitchRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = SwitchRec
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type SwitchCPUWideRecord struct {
	Header
	RecordType  RecordType `json:"record_type"`
	NextPrevPid uint32     `json:"next_prev_pid"`
	NextPrevTid uint32     `json:"next_prev_tid"`
	SampleID
}

func (rec *SwitchCPUWideRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = SwitchCPUWideRec
	parser.Uint32(&rec.NextPrevPid)
	parser.Uint32(&rec.NextPrevTid)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type NamespacesContent struct {
	Dev uint64 `json:"dev"`
	Ino uint64 `json:"inode"`
}

type NamespacesRecord struct {
	Header
	RecordType RecordType          `json:"record_type"`
	Pid        uint32              `json:"pid"`
	Tid        uint32              `json:"tid"`
	Namespaces []NamespacesContent `json:"namespaces"`
	SampleID
}

func (rec *NamespacesRecord) Decode(raw *RawRecord, attr *Attr, symbolizer *symbol.Symbolizer) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	rec.RecordType = NamespacesRec
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Tid)

	var num uint64
	parser.Uint64(&num)
	rec.Namespaces = make([]NamespacesContent, num)
	for i := 0; i < int(num); i++ {
		parser.Uint64(&rec.Namespaces[i].Dev)
		parser.Uint64(&rec.Namespaces[i].Ino)
	}
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

type RecordParser struct {
	ringBuf    *RingBuf
	attr       *Attr
	symbolizer *symbol.Symbolizer
}

func NewRecordParser(ringBuf *RingBuf, attr *Attr) (*RecordParser, error) {
	symbolizer, err := symbol.NewSymbolizer()
	if err != nil {
		return nil, err
	}
	return &RecordParser{
		ringBuf:    ringBuf,
		attr:       attr,
		symbolizer: symbolizer,
	}, nil
}

func (parser *RecordParser) getRawRecord() *RawRecord {
	var raw RawRecord
	buf := parser.ringBuf
	head := atomic.LoadUint64(&buf.MetaPage.Data_head)
	tail := atomic.LoadUint64(&buf.MetaPage.Data_tail)
	if head == tail {
		return nil
	}

	start := tail % uint64(len(buf.RingData))
	raw.Header = *(*Header)(unsafe.Pointer(&buf.RingData[start]))
	end := (tail + uint64(raw.Header.Size)) % uint64(len(buf.RingData))

	start = (start + uint64(unsafe.Sizeof(raw.Header))) % uint64(len(buf.RingData))
	raw.Data = make([]byte, raw.Header.Size-uint16(unsafe.Sizeof(raw.Header)))
	if start > end {
		n := copy(raw.Data, buf.RingData[start:])
		copy(raw.Data[n:], buf.RingData[:int(raw.Header.Size)-n])
	} else {
		copy(raw.Data, buf.RingData[start:end])
	}

	atomic.AddUint64(&buf.MetaPage.Data_tail, uint64(raw.Header.Size))
	return &raw
}

func (parser *RecordParser) newRecord(raw *RawRecord) (Record, error) {
	var rec Record
	switch raw.Header.Type {
	case MmapRec:
		rec = &MmapRecord{}
	case LostRec:
		rec = &LostRecord{}
	case CommRec:
		rec = &CommRecord{}
	case ExitRec:
		rec = &ExitRecord{}
	case ThrottleRec:
		rec = &ThrottleRecord{}
	case UnthrottleRec:
		rec = &UnthrottleRecord{}
	case ForkRec:
		rec = &ForkRecord{}
	case ReadRec:
		if parser.attr.ReadFormat.Group {
			rec = &GroupReadRecord{}
		} else {
			rec = &ReadRecord{}
		}
	case SampleRec:
		if parser.attr.ReadFormat.Group {
			rec = &GroupSampleRecord{}
		} else {
			rec = &SampleRecord{}
		}
	case Mmap2Rec:
		rec = &Mmap2Record{}
	case AuxRec:
		rec = &AuxRecord{}
	case ItraceStartRec:
		rec = &ItraceStartRecord{}
	case LostSamplesRec:
		rec = &LostSamplesRecord{}
	case SwitchRec:
		rec = &SwitchRecord{}
	case SwitchCPUWideRec:
		rec = &SwitchCPUWideRecord{}
	case NamespacesRec:
		rec = &NamespacesRecord{}
	default:
		return nil, fmt.Errorf("Unhandled type [%d]", int(raw.Header.Type))
	}
	rec.Decode(raw, parser.attr, parser.symbolizer)
	return rec, nil
}

func (parser *RecordParser) GetRecord() (Record, error) {
	raw := parser.getRawRecord()
	if raw == nil {
		return nil, nil
	}

	return parser.newRecord(raw)
}
