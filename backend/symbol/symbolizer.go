package symbol

import (
	"sync"

	"github.com/golang/groupcache/lru"
)

type CpuMode int

const (
	KernelMode CpuMode = iota
	UserMode

	UnknownMode
)

const LRUCacheSize = 128

type Records map[uint64]string

type Symbolizer struct {
	mutex      *sync.RWMutex
	cache      *lru.Cache
	ksymParser *KsymParser
}

func NewSymbolizer(dbgDirPath string) *Symbolizer {
	return &Symbolizer{
		mutex:      &sync.RWMutex{},
		cache:      lru.New(LRUCacheSize),
		ksymParser: NewKsymParser(dbgDirPath),
	}
}

func (inst *Symbolizer) Symbolize(cpuMode CpuMode, buildID string, addr uint64) (string, error) {
	symbol := ""
	inst.mutex.RLock()
	if recs, ok := inst.cache.Get(buildID); ok {
		if rec, exist := recs.(Records)[addr]; exist {
			symbol = rec
		}
	}
	inst.mutex.RUnlock()

	if symbol != "" {
		return symbol, nil
	}

	switch cpuMode {
	case KernelMode:
		symbol = inst.ksymParser.Resolve(buildID, addr)
	}

	inst.mutex.Lock()
	defer inst.mutex.Unlock()
	recs := Records{}
	if _recs, ok := inst.cache.Get(buildID); ok {
		recs = _recs.(Records)
	}
	recs[addr] = symbol
	inst.cache.Remove(buildID)
	inst.cache.Add(buildID, recs)
	return symbol, nil
}
