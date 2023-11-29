package perf

import (
	"encoding/json"

	"hermes/backend/utils"
)

const ANON_COMM = "anon"

type ThreadInfo struct {
	Comm string
}

type ThreadsInfo struct {
	threads map[uint64]ThreadInfo
}

func (info *ThreadsInfo) getIndex(pid, tid uint32) uint64 {
	return uint64(pid)<<32 | uint64(tid)
}

func (info *ThreadsInfo) SetComm(comm string, pid, tid uint32) {
	index := info.getIndex(pid, tid)
	if threadInfo, isExist := info.threads[index]; isExist {
		threadInfo.Comm = comm
		info.threads[index] = threadInfo
	} else {
		info.threads[index] = ThreadInfo{
			Comm: comm,
		}
	}
}

func (info *ThreadsInfo) Find(pid, tid uint32) *ThreadInfo {
	index := info.getIndex(pid, tid)
	if threadInfo, isExist := info.threads[index]; isExist {
		return &threadInfo
	}
	return nil
}

type RecordHandler struct {
	flameGraphData *utils.FlameGraphData
	threadsInfo    ThreadsInfo
}

func GetRecordHandler() *RecordHandler {
	return &RecordHandler{
		flameGraphData: utils.NewFlameGraphData(),
		threadsInfo: ThreadsInfo{
			threads: make(map[uint64]ThreadInfo),
		},
	}
}

func (handler *RecordHandler) parseCommRec(bytes []byte) error {
	var rec CommRecord
	if err := json.Unmarshal(bytes, &rec); err != nil {
		return err
	}

	handler.threadsInfo.SetComm(rec.Comm, rec.Pid, rec.Tid)
	return nil
}

func (handler *RecordHandler) parseSampleRec(bytes []byte) error {
	var rec SampleRecord
	if err := json.Unmarshal(bytes, &rec); err != nil {
		return err
	}
	stack := []string{}
	for _, inst := range rec.CallchainInsts {
		if inst.Symbol == "" {
			continue
		}
		stack = append(stack, inst.Symbol)
	}
	if threadInfo := handler.threadsInfo.Find(rec.Pid, rec.Tid); threadInfo != nil {
		stack = append(stack, threadInfo.Comm)
	} else {
		stack = append(stack, ANON_COMM)
	}
	handler.flameGraphData.Add(&stack, len(stack)-1, 1)
	return nil
}

func (handler *RecordHandler) Parse(bytes []byte) error {
	var header Header
	if err := json.Unmarshal(bytes, &header); err != nil {
		return err
	}

	switch header.Type {
	case CommRec:
		return handler.parseCommRec(bytes)
	case SampleRec:
		return handler.parseSampleRec(bytes)
	}
	return nil
}

func (handler *RecordHandler) GetFlameGraphData() *utils.FlameGraphData {
	return handler.flameGraphData
}
