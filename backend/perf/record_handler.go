package perf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"hermes/backend/symbol"
	"hermes/backend/utils"

	"golang.org/x/sys/unix"
)

const (
	AnonComm = "anon"
	DbgDir   = ".hermes.perf.dbg"
)

const (
	KernelThreadPid = 0
	KernelThreadTid = 0
)

type Map struct {
	start   uint64
	end     uint64
	buildID string
}

type Maps []Map

func (inst Maps) Len() int {
	return len(inst)
}

func (inst Maps) Swap(lhs, rhs int) {
	inst[lhs], inst[rhs] = inst[rhs], inst[lhs]
}

func (inst Maps) Less(lhs, rhs int) bool {
	return inst[lhs].start < inst[rhs].start
}

type ThreadInfo struct {
	Comm string
	maps Maps
}

func (inst *ThreadInfo) Find(ip uint64) (string, error) {
	idx := sort.Search(len(inst.maps), func(i int) bool {
		return ip >= inst.maps[i].start && ip <= inst.maps[i].end
	})
	if idx < len(inst.maps) {
		return inst.maps[idx].buildID, nil
	}
	return "", fmt.Errorf("Failed to find thread info with ip [%d]", ip)
}

type ThreadsInfo map[uint64]ThreadInfo

func (inst *ThreadsInfo) getIndex(pid, tid uint32) uint64 {
	return uint64(pid)<<32 | uint64(tid)
}

func (inst *ThreadsInfo) SetComm(comm string, pid, tid uint32) {
	index := inst.getIndex(pid, tid)
	if threadInfo, isExist := (*inst)[index]; isExist {
		threadInfo.Comm = comm
		(*inst)[index] = threadInfo
	} else {
		(*inst)[index] = ThreadInfo{
			Comm: comm,
		}
	}
}

func (inst *ThreadsInfo) Find(pid, tid uint32) *ThreadInfo {
	index := inst.getIndex(pid, tid)
	if threadInfo, isExist := (*inst)[index]; isExist {
		return &threadInfo
	}
	return nil
}

type RecordHandler struct {
	dbgDirPath     string
	flameGraphData *utils.FlameGraphData
	threadsInfo    ThreadsInfo
	symbolizer     symbol.Symbolizer
}

func NewRecordHandler() (*RecordHandler, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dbgDirPath := filepath.Join(homeDir, DbgDir)

	return &RecordHandler{
		dbgDirPath:     dbgDirPath,
		flameGraphData: utils.NewFlameGraphData(),
		threadsInfo:    ThreadsInfo{},
		symbolizer:     *symbol.NewSymbolizer(dbgDirPath),
	}, nil
}

func (inst *RecordHandler) parseCommRec(bytes []byte) error {
	var rec CommRecord
	if err := json.Unmarshal(bytes, &rec); err != nil {
		return err
	}

	inst.threadsInfo.SetComm(rec.Comm, rec.Pid, rec.Tid)
	return nil
}

func (inst *RecordHandler) parseSymbol(pid, tid uint32, cpuMode symbol.CpuMode, ip uint64) string {
	_symbol := fmt.Sprintf("0x%x", ip)
	if cpuMode == symbol.UnknownMode {
		return _symbol
	}

	var threadInfo *ThreadInfo
	switch cpuMode {
	case symbol.KernelMode:
		threadInfo = inst.threadsInfo.Find(0, 0)
	case symbol.UserMode:
		threadInfo = inst.threadsInfo.Find(pid, tid)
	}

	if threadInfo == nil {
		return _symbol
	}

	buildID, err := threadInfo.Find(ip)
	if err != nil {
		return _symbol
	}
	if __symbol, err := inst.symbolizer.Symbolize(cpuMode, buildID, ip); err == nil {
		_symbol = __symbol
	}
	return _symbol
}

func (inst *RecordHandler) parseSampleRec(bytes []byte) error {
	var rec SampleRecord
	if err := json.Unmarshal(bytes, &rec); err != nil {
		return err
	}

	stack := []string{}
	cpuMode := symbol.UnknownMode
	if rec.Misc&unix.PERF_RECORD_MISC_KERNEL == unix.PERF_RECORD_MISC_KERNEL {
		cpuMode = symbol.KernelMode
	} else if rec.Misc&unix.PERF_RECORD_MISC_USER == unix.PERF_RECORD_MISC_USER {
		cpuMode = symbol.UserMode
	}

	for _, ip := range rec.CallchainIps {
		_symbol := inst.parseSymbol(rec.Pid, rec.Tid, cpuMode, ip)
		stack = append(stack, _symbol)
	}
	if threadInfo := inst.threadsInfo.Find(rec.Pid, rec.Tid); threadInfo != nil {
		stack = append(stack, threadInfo.Comm)
	} else {
		stack = append(stack, AnonComm)
	}
	inst.flameGraphData.Add(&stack, len(stack)-1, 1)
	return nil
}

func (inst *RecordHandler) createKernelMap(_buildID string) error {
	ksymParser := symbol.NewKsymParser(inst.dbgDirPath)
	start, end, err := ksymParser.GetMapRange(_buildID)
	if err != nil {
		return err
	}
	inst.threadsInfo[0] = ThreadInfo{
		Comm: "",
		maps: Maps{
			{
				start:   start,
				end:     end,
				buildID: _buildID,
			},
		},
	}
	return nil
}

func (inst *RecordHandler) PrepareKernelSymbol(kernSymPath string) error {
	buildID, err := symbol.KernelSymPrepare(inst.dbgDirPath, kernSymPath)
	if err != nil {
		return err
	}
	return inst.createKernelMap(buildID)
}

func (inst *RecordHandler) Parse(bytes []byte) error {
	var header Header
	if err := json.Unmarshal(bytes, &header); err != nil {
		return err
	}

	switch header.Type {
	case CommRec:
		return inst.parseCommRec(bytes)
	case SampleRec:
		return inst.parseSampleRec(bytes)
	}
	return nil
}

func (inst *RecordHandler) GetFlameGraphData() *utils.FlameGraphData {
	return inst.flameGraphData
}
