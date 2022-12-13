package ebpf

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/sirupsen/logrus"
	"github.com/yukariatlas/hermes/backend/symbol"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS bpf memory_alloc.c -- -I../header

type MemoryLoader struct {
	objs        *bpfObjects
	symbolizer  *symbol.Symbolizer
	outputFiles []string
}

func GetLoader() (*MemoryLoader, error) {
	// Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		return nil, err
	}
	symbolizer, err := symbol.NewSymbolizer()
	if err != nil {
		return nil, err
	}

	return &MemoryLoader{
		objs:        &objs,
		symbolizer:  symbolizer,
		outputFiles: []string{},
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
	BytesAlloc     uint64   `json:"bytes_alloc"`
	CallchainInsts []string `json:"callchain_insts"`
}

type DataRecord struct {
	BytesOwned uint64        `json:"bytes_owned"`
	Comm       string        `json:"comm"`
	AllocRecs  []AllocRecord `json:"alloc_records"`
}

func uint8ToString(val []uint8) string {
	bytes := []byte{}
	for _, _byte := range val {
		if _byte == 0 {
			break
		}
		bytes = append(bytes, byte(_byte))
	}

	if len(bytes) == 0 {
		return string("unknown")
	}
	return string(bytes)
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

		ips := make([]uint64, CallStackSize)
		if err := loader.objs.StackTrace.Lookup(info.StackId, &ips); err != nil {
			logrus.Errorf("Failed to lookup stack id [%d] in stack trace", info.StackId)
			continue
		}

		allocRec := AllocRecord{
			BytesAlloc:     info.Size,
			CallchainInsts: []string{},
		}
		for _, ip := range ips {
			if ip == 0 {
				break
			}
			symbol, err := loader.symbolizer.Symbolize(ip)
			if err != nil {
				symbol = ""
			}
			allocRec.CallchainInsts = append(allocRec.CallchainInsts, symbol)
		}
		rec.Comm = uint8ToString(info.Comm[:])
		rec.AllocRecs = append(rec.AllocRecs, allocRec)
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
	slabOutputFile := outputPath + string(".slab")
	if err := loader.writeToFile(slabOutputFile, recs); err != nil {
		logrus.Errorf("Failed to write slab records to file, err [%s]", err)
		return err
	}
	loader.outputFiles = append(loader.outputFiles, filepath.Base(slabOutputFile))

	recs = loader.getDataRecByType(Page)
	pageOutputFile := outputPath + string(".page")
	if err := loader.writeToFile(pageOutputFile, recs); err != nil {
		logrus.Errorf("Failed to write page records to file, err [%s]", err)
		return err
	}
	loader.outputFiles = append(loader.outputFiles, filepath.Base(pageOutputFile))

	return nil
}

func (loader *MemoryLoader) GetOutputFiles() []string {
	return loader.outputFiles
}

func (loader *MemoryLoader) Close() {
	loader.objs.Close()
}
