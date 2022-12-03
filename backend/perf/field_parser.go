package perf

import (
	"unsafe"
)

type FieldParser []byte

func (parser *FieldParser) advance(c int) {
	*parser = (*parser)[c:]
}

func (parser *FieldParser) Uint64(val *uint64) {
	*val = *(*uint64)(unsafe.Pointer(&(*parser)[0]))
	parser.advance(8)
}

func (parser *FieldParser) Uint64Cond(cond bool, val *uint64) {
	if cond {
		parser.Uint64(val)
	}
}

func (parser *FieldParser) Uint32(val *uint32) {
	*val = *(*uint32)(unsafe.Pointer(&(*parser)[0]))
	parser.advance(4)
}

func (parser *FieldParser) Uint32Cond(cond bool, val *uint32) {
	if cond {
		parser.Uint32(val)
	}
}

func (parser *FieldParser) String(val *string) {
	for i := 0; i < len(*parser); i++ {
		if (*parser)[i] == 0 {
			*val = string((*parser)[:i])
			if i+1 <= len(*parser) {
				parser.advance(i + 1)
			}
			return
		}
	}
}

func (parser *FieldParser) BytesByUint32Size(val *[]byte) {
	size := *(*uint32)(unsafe.Pointer(&(*parser)[0]))
	parser.advance(4)
	data := make([]byte, size)
	copy(data, *parser)
	*val = data
	parser.advance(int(size))
}

func (parser *FieldParser) BytesByUint64Size(val *[]byte) {
	size := *(*uint64)(unsafe.Pointer(&(*parser)[0]))
	parser.advance(8)
	data := make([]byte, size)
	copy(data, *parser)
	*parser = data
	parser.advance(int(size))
}

func (parser *FieldParser) ParseSampleID(sampleIDAll bool, sampleFormat SampleFormat, sampleID *SampleID) {
	if !sampleIDAll {
		return
	}
	parser.Uint32Cond(sampleFormat.Tid, &sampleID.Pid)
	parser.Uint32Cond(sampleFormat.Tid, &sampleID.Tid)
	parser.Uint64Cond(sampleFormat.Time, &sampleID.Time)
	parser.Uint64Cond(sampleFormat.ID, &sampleID.ID)
	parser.Uint64Cond(sampleFormat.StreamID, &sampleID.StreamID)
	parser.Uint32Cond(sampleFormat.CPU, &sampleID.CPU)
	parser.advance(4)
	parser.Uint64Cond(sampleFormat.Identifier, &sampleID.Identifier)
}

func (parser *FieldParser) ParseReadContent(readFormat ReadFormat, readContent *ReadContent) {
	parser.Uint64(&readContent.Value)
	parser.Uint64Cond(readFormat.TotalTimeEnabled, &readContent.TimeEnabled)
	parser.Uint64Cond(readFormat.TotalTimeRunning, &readContent.TimeRunning)
	parser.Uint64Cond(readFormat.ID, &readContent.ID)
}

func (parser *FieldParser) ParseGroupReadContent(readFormat ReadFormat, groupReadContent *GroupReadContent) {
	var nr uint64
	parser.Uint64(&nr)
	parser.Uint64Cond(readFormat.TotalTimeEnabled, &groupReadContent.TimeEnabled)
	parser.Uint64Cond(readFormat.TotalTimeRunning, &groupReadContent.TimeRunning)
	groupReadContent.Values = make([]ReadContentValue, nr)
	for i := 0; i < int(nr); i++ {
		parser.Uint64(&groupReadContent.Values[i].Value)
		parser.Uint64Cond(readFormat.ID, &groupReadContent.Values[i].ID)
	}
}
