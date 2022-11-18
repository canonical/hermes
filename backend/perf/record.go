package perf

import (
	"fmt"
	"math/bits"
	"sync/atomic"
	"unsafe"

	"github.com/sirupsen/logrus"
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
	Type RecordType
	Misc uint16
	Size uint16
}

type RawRecord struct {
	Header Header
	Data   []byte
}

type Record interface {
	Decode(raw *RawRecord, attr *Attr)
	Process()
}

type SampleID struct {
	Pid        uint32
	Tid        uint32
	Time       uint64
	ID         uint64
	StreamID   uint64
	CPU        uint32
	_          uint32 // reserved
	Identifier uint64
}

type MmapRecord struct {
	Header
	Pid      uint32
	Tid      uint32
	Addr     uint64
	Len      uint64
	Pgoff    uint64
	Filename string
	SampleID
}

func (rec *MmapRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Tid)
	parser.Uint64(&rec.Addr)
	parser.Uint64(&rec.Len)
	parser.Uint64(&rec.Pgoff)
	parser.String(&rec.Filename)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *MmapRecord) Process() {
	logrus.Errorf("[MmapRecord] Pid [%d]", rec.Pid)
}

type LostRecord struct {
	Header
	ID   uint64
	Lost uint64
	SampleID
}

func (rec *LostRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint64(&rec.ID)
	parser.Uint64(&rec.Lost)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *LostRecord) Process() {
	logrus.Errorf("[LostRecord] ID [%d]", rec.ID)
}

type CommRecord struct {
	Header
	Pid  uint32
	Tid  uint32
	Comm string
	SampleID
}

func (rec *CommRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Tid)
	parser.String(&rec.Comm)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *CommRecord) Process() {
	logrus.Errorf("[CommRecord] Pid [%d]", rec.Pid)
}

type ExitRecord struct {
	Header
	Pid  uint32
	Ppid uint32
	Tid  uint32
	Ptid uint32
	Time uint64
	SampleID
}

func (rec *ExitRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Ppid)
	parser.Uint32(&rec.Tid)
	parser.Uint32(&rec.Ptid)
	parser.Uint64(&rec.Time)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *ExitRecord) Process() {
	logrus.Errorf("[ExitRecord] Pid [%d]", rec.Pid)
}

type ThrottleRecord struct {
	Header
	Time     uint64
	ID       uint64
	StreamID uint64
	SampleID
}

func (rec *ThrottleRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint64(&rec.Time)
	parser.Uint64(&rec.ID)
	parser.Uint64(&rec.StreamID)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *ThrottleRecord) Process() {
	logrus.Errorf("[ThrottleRecord] ID [%d]", rec.ID)
}

type UnthrottleRecord struct {
	Header
	Time     uint64
	ID       uint64
	StreamID uint64
	SampleID
}

func (rec *UnthrottleRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint64(&rec.Time)
	parser.Uint64(&rec.ID)
	parser.Uint64(&rec.StreamID)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *UnthrottleRecord) Process() {
	logrus.Errorf("[UnthrottleRecord] ID [%d]", rec.ID)
}

type ForkRecord struct {
	Header
	Pid  uint32
	Ppid uint32
	Tid  uint32
	Ptid uint32
	Time uint64
	SampleID
}

func (rec *ForkRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Ppid)
	parser.Uint32(&rec.Tid)
	parser.Uint32(&rec.Ptid)
	parser.Uint64(&rec.Time)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *ForkRecord) Process() {
	logrus.Errorf("[ForkRecord] Pid [%d]", rec.Pid)
}

type ReadContent struct {
	Value       uint64
	TimeEnabled uint64
	TimeRunning uint64
	ID          uint64
}

type ReadRecord struct {
	Header
	Pid         uint32
	Tid         uint32
	ReadContent ReadContent
	SampleID
}

func (rec *ReadRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Tid)
	parser.ParseReadContent(attr.ReadFormat, &rec.ReadContent)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *ReadRecord) Process() {
	logrus.Errorf("[ReadRecord] Pid [%d]", rec.Pid)
}

type BranchEntry struct {
	From         uint64
	To           uint64
	Mispredicted bool
	Predicted    bool
	InTrans      bool
	Abort        bool
	Cycles       uint16
	BranchType   uint8
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

type SampleRecord struct {
	Header
	Identifier       uint64
	IP               uint64
	Pid              uint32
	Tid              uint32
	Time             uint64
	Addr             uint64
	ID               uint64
	StreamID         uint64
	CPU              uint32
	_                uint32
	Period           uint64
	ReadContent      ReadContent
	Callchain        []uint64
	Raw              []byte
	BranchStack      []BranchEntry
	RegsABI          uint64
	RegsUser         []uint64
	StackUser        []byte
	StackUserDynSize uint64
	Weight           uint64
	DataSrc          uint64
	Transaction      uint64
	RegsIntrABI      uint64
	RegsIntr         []uint64
	PhysAddr         uint64
	Aux              []byte
	DataPageSize     uint64
	CodePageSize     uint64
}

func (rec *SampleRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint64Cond(attr.SampleFormat.Identifier, &rec.Identifier)
	parser.Uint64Cond(attr.SampleFormat.IP, &rec.IP)
	parser.Uint32Cond(attr.SampleFormat.Tid, &rec.Pid)
	parser.Uint32Cond(attr.SampleFormat.Tid, &rec.Tid)
	parser.Uint64Cond(attr.SampleFormat.Time, &rec.Time)
	parser.Uint64Cond(attr.SampleFormat.Addr, &rec.Addr)
	parser.Uint64Cond(attr.SampleFormat.ID, &rec.ID)
	parser.Uint64Cond(attr.SampleFormat.StreamID, &rec.StreamID)

	var reserved uint32
	parser.Uint32Cond(attr.SampleFormat.CPU, &rec.CPU)
	parser.Uint32Cond(attr.SampleFormat.CPU, &reserved)
	if attr.SampleFormat.Read {
		parser.ParseReadContent(attr.ReadFormat, &rec.ReadContent)
	}
	if attr.SampleFormat.Callchain {
		var nr uint64
		parser.Uint64(&nr)
		rec.Callchain = make([]uint64, nr)
		for i := 0; i < int(nr); i++ {
			parser.Uint64(&rec.Callchain[i])
		}
	}
	if attr.SampleFormat.Raw {
		parser.BytesByUint32Size(&rec.Raw)
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
		parser.Uint64(&rec.RegsABI)
		nr := bits.OnesCount64(attr.SampleRegsUser)
		rec.RegsUser = make([]uint64, nr)
		for i := 0; i < nr; i++ {
			parser.Uint64(&rec.RegsUser[i])
		}
	}
	if attr.SampleFormat.StackUser {
		parser.BytesByUint64Size(&rec.StackUser)
		if len(rec.StackUser) > 0 {
			parser.Uint64(&rec.StackUserDynSize)
		}
	}
	parser.Uint64Cond(attr.SampleFormat.Weight, &rec.Weight)
	parser.Uint64Cond(attr.SampleFormat.DataSrc, &rec.DataSrc)
	parser.Uint64Cond(attr.SampleFormat.Transaction, &rec.Transaction)
	if attr.SampleFormat.RegsIntr {
		parser.Uint64(&rec.RegsIntrABI)
		nr := bits.OnesCount64(attr.SampleRegsIntr)
		rec.RegsIntr = make([]uint64, nr)
		for i := 0; i < int(nr); i++ {
			parser.Uint64(&rec.RegsIntr[i])
		}
	}
	parser.Uint64Cond(attr.SampleFormat.PhysAddr, &rec.PhysAddr)
	if attr.SampleFormat.Aux {
		parser.BytesByUint64Size(&rec.Aux)
	}
	parser.Uint64Cond(attr.SampleFormat.DataPageSize, &rec.DataPageSize)
	parser.Uint64Cond(attr.SampleFormat.CodePageSize, &rec.CodePageSize)
}

func (rec *SampleRecord) Process() {
	logrus.Errorf("[SampleRecord] IP [%d]", rec.IP)
}

type Mmap2Record struct {
	Header
	Pid           uint32
	Tid           uint32
	Addr          uint64
	Len           uint64
	Pgoff         uint64
	MajorID       uint32
	MinorID       uint32
	Ino           uint64
	InoGeneration uint64
	Prot          uint32
	Flags         uint32
	Filename      string
	SampleID
}

func (rec *Mmap2Record) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
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

func (rec *Mmap2Record) Process() {
	logrus.Errorf("[Mmap2Record] Pid [%d]", rec.Pid)
}

type AuxRecord struct {
	Header
	Offset uint64
	Size   uint64
	Flags  uint64
	SampleID
}

func (rec *AuxRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint64(&rec.Offset)
	parser.Uint64(&rec.Size)
	parser.Uint64(&rec.Flags)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *AuxRecord) Process() {
	logrus.Errorf("[AuxRecord] Offset [%d]", rec.Offset)
}

type ItraceStartRecord struct {
	Header
	Pid uint32
	Tid uint32
	SampleID
}

func (rec *ItraceStartRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint32(&rec.Pid)
	parser.Uint32(&rec.Tid)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *ItraceStartRecord) Process() {
	logrus.Errorf("[ItraceStartRecord] Pid [%d]", rec.Pid)
}

type LostSamplesRecord struct {
	Header
	Lost uint64
	SampleID
}

func (rec *LostSamplesRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint64(&rec.Lost)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *LostSamplesRecord) Process() {
	logrus.Errorf("[LostSamplesRecord] Lost [%d]", rec.Lost)
}

type SwitchRecord struct {
	Header
	SampleID
}

func (rec *SwitchRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *SwitchRecord) Process() {
	logrus.Errorf("[SwitchRecord]")
}

type SwitchCPUWideRecord struct {
	Header
	NextPrevPid uint32
	NextPrevTid uint32
	SampleID
}

func (rec *SwitchCPUWideRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
	parser.Uint32(&rec.NextPrevPid)
	parser.Uint32(&rec.NextPrevTid)
	parser.ParseSampleID(attr.Options.SampleIDAll, attr.SampleFormat, &rec.SampleID)
}

func (rec *SwitchCPUWideRecord) Process() {
	logrus.Errorf("[SwitchCPUWideRecord]")
}

type NamespacesContent struct {
	Dev uint64
	Ino uint64
}

type NamespacesRecord struct {
	Header
	Pid        uint32
	Tid        uint32
	Namespaces []NamespacesContent
	SampleID
}

func (rec *NamespacesRecord) Decode(raw *RawRecord, attr *Attr) {
	parser := FieldParser(raw.Data)
	rec.Header = raw.Header
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

func (rec *NamespacesRecord) Process() {
	logrus.Errorf("[NamespacesRecord] Pid [%d]", rec.Pid)
}

type RecordParser struct {
	ringBuf *RingBuf
	attr    *Attr
}

func NewRecordParser(ringBuf *RingBuf, attr *Attr) *RecordParser {
	return &RecordParser{
		ringBuf: ringBuf,
		attr:    attr,
	}
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

	var data []byte
	if start > end {
		data = make([]byte, raw.Header.Size)
		n := copy(data, buf.RingData[start:])
		copy(data[n:], buf.RingData[:int(raw.Header.Size)-n])
	} else {
		data = buf.RingData[start:end]
	}

	raw.Data = data[unsafe.Sizeof(raw.Header):]
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
		rec = &ReadRecord{}
	case SampleRec:
		rec = &SampleRecord{}
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
	rec.Decode(raw, parser.attr)
	return rec, nil
}

func (parser *RecordParser) GetRecord() (Record, error) {
	raw := parser.getRawRecord()
	if raw == nil {
		return nil, nil
	}

	return parser.newRecord(raw)
}
