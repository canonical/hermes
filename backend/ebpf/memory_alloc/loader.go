package ebpf

import (
	"context"
	"encoding/json"
	"os"
	"sort"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"github.com/sirupsen/logrus"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS bpf memory_alloc.c -- -I../header

type MemoryLoader struct {
	objs *bpfObjects
}

func GetLoader() (*MemoryLoader, error) {
	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, err
	}

	// Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		return nil, err
	}

	return &MemoryLoader{
		objs: &objs,
	}, nil
}

func (loader *MemoryLoader) Load(ctx context.Context) error {
	tpKmalloc, err := link.Tracepoint("kmem", "kmalloc", loader.objs.Kmalloc, nil)
	if err != nil {
		logrus.Errorf("Failed to open kmalloc tracepoint, err [%s]", err)
		return err
	}
	defer tpKmalloc.Close()

	tpKmallocNode, err := link.Tracepoint("kmem", "kmalloc_node", loader.objs.KmallocNode, nil)
	if err != nil {
		logrus.Errorf("Failed to open kmalloc_node tracepoint, err [%s]", err)
		return err
	}
	defer tpKmallocNode.Close()

	tpKfree, err := link.Tracepoint("kmem", "kfree", loader.objs.Kfree, nil)
	if err != nil {
		logrus.Errorf("Failed to open kfree tracepoint, err [%s]", err)
		return err
	}
	defer tpKfree.Close()

	tpKmemCacheAlloc, err := link.Tracepoint("kmem", "kmem_cache_alloc", loader.objs.KmemCacheAlloc, nil)
	if err != nil {
		logrus.Errorf("Failed to open kmem_cache_alloc tracepoint, err [%s]", err)
		return err
	}
	defer tpKmemCacheAlloc.Close()

	tpKmemCacheAllocNode, err := link.Tracepoint("kmem", "kmem_cache_alloc_node", loader.objs.KmemCacheAllocNode, nil)
	if err != nil {
		logrus.Errorf("Failed to open kmem_cache_alloc_node tracepoint, err [%s]", err)
		return err
	}
	defer tpKmemCacheAllocNode.Close()

	tpKmemCacheFree, err := link.Tracepoint("kmem", "kmem_cache_free", loader.objs.KmemCacheFree, nil)
	if err != nil {
		logrus.Errorf("Failed to open kmem_cache_free tracepoint, err [%s]", err)
		return err
	}
	defer tpKmemCacheFree.Close()

	tpMmPageAlloc, err := link.Tracepoint("kmem", "mm_page_alloc", loader.objs.MmPageAlloc, nil)
	if err != nil {
		logrus.Errorf("Failed to open mm_page_alloc tracepoint, err [%s]", err)
		return err
	}
	defer tpMmPageAlloc.Close()

	tpMmPageFree, err := link.Tracepoint("kmem", "mm_page_free", loader.objs.MmPageFree, nil)
	if err != nil {
		logrus.Errorf("Failed to open mm_page_free tracepoint, err [%s]", err)
		return err
	}
	defer tpMmPageFree.Close()

	<-ctx.Done()
	return nil
}

const CallStackSize = 127

type MemoryType uint8

const (
	Slab MemoryType = iota
	Page
)

type AllocRecord struct {
	BytesAlloc uint64   `json:"bytes_alloc"`
	CallStack  []uint64 `json:"call_stack"`
}

type DataRecord struct {
	BytesOwned uint64        `json:"bytes_owned"`
	AllocRecs  []AllocRecord `json:"alloc_records"`
}

func (loader *MemoryLoader) getDataRec(infoMap, statsMap *ebpf.Map, recs *map[uint64]DataRecord) {
	var addr uint64
	var info bpfInfoValue
	infoIter := infoMap.Iterate()
	for infoIter.Next(&addr, &info) {
		rec, _ := (*recs)[info.TgidPid]
		if err := statsMap.Lookup(info.StackId, &rec.BytesOwned); err != nil {
			logrus.Errorf("Failed to lookup stack id [%d] in stats map", info.StackId)
			continue
		}
		callStack := make([]uint64, CallStackSize)
		if err := loader.objs.StackTrace.Lookup(info.StackId, &callStack); err != nil {
			logrus.Errorf("Failed to lookup stack id [%d] in stack trace", info.StackId)
			continue
		}
		idx := sort.Search(len(callStack), func(idx int) bool { return callStack[idx] == 0 })
		rec.AllocRecs = append(rec.AllocRecs, AllocRecord{
			BytesAlloc: info.Size,
			CallStack:  callStack[:idx],
		})
		(*recs)[info.TgidPid] = rec
	}
}

func (loader *MemoryLoader) getDataRecByType(memoryType MemoryType) *map[uint64]DataRecord {
	recs := map[uint64]DataRecord{}
	switch memoryType {
	case Slab:
		loader.getDataRec(loader.objs.SlabInfo, loader.objs.SlabStats, &recs)
	case Page:
		loader.getDataRec(loader.objs.PageInfo, loader.objs.PageStats, &recs)
	}
	return &recs
}

func (loader *MemoryLoader) writeToFile(outputPath string, recs *map[uint64]DataRecord) error {
	bytes, err := json.Marshal(recs)
	if err != nil {
		return err
	}
	fp, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString(string(bytes)); err != nil {
		return err
	}
	return nil
}

func (loader *MemoryLoader) StoreData(outputPath string) error {
	recs := loader.getDataRecByType(Slab)
	if err := loader.writeToFile(outputPath+string(".slab"), recs); err != nil {
		logrus.Errorf("Failed to write slab records to file, err [%s]", err.Error())
		return err
	}

	recs = loader.getDataRecByType(Page)
	if err := loader.writeToFile(outputPath+string(".page"), recs); err != nil {
		logrus.Errorf("Failed to write page records to file, err [%s]", err.Error())
		return err
	}
	return nil
}

func (loader *MemoryLoader) Close() {
	loader.objs.Close()
}
