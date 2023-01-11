package ebpf

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"hermes/backend/symbol"
	"hermes/backend/utils"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/sirupsen/logrus"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target amd64 -cc $BPF_CLANG -cflags $BPF_CFLAGS bpf memory_alloc.c -- -I../header

const SlabInfoFilePostfix = ".slab.info"
const SlabRecFilePostfix = ".slab.rec"

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

	kpKmemCacheAlloc, err := link.Kprobe("kmem_cache_alloc", loader.objs.KmemCacheAllocKprobe, nil)
	if err != nil {
		logrus.Errorf("Failed to open kmem_cache_alloc kprobe, err [%s]", err)
		return err
	}
	defer kpKmemCacheAlloc.Close()

	tpKmemCacheAlloc, err := link.Tracepoint("kmem", "kmem_cache_alloc", loader.objs.KmemCacheAlloc, nil)
	if err != nil {
		logrus.Errorf("Failed to open kmem_cache_alloc tracepoint, err [%s]", err)
		return err
	}
	defer tpKmemCacheAlloc.Close()

	kpKmemCacheAllocNode, err := link.Kprobe("kmem_cache_alloc_node", loader.objs.KmemCacheAllocNodeKprobe, nil)
	if err != nil {
		logrus.Errorf("Failed to open kmem_cache_alloc_node kprobe, err [%s]", err)
		return err
	}
	defer kpKmemCacheAllocNode.Close()

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

	<-ctx.Done()
	return nil
}

const CallStackSize = 127

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

type AllocDetail struct {
	BytesAlloc     int64    `json:"bytes_alloc"`
	CallchainInsts []string `json:"callchain_insts"`
}

type AllocRecord struct {
	Comm         string        `json:"comm"`
	AllocDetails []AllocDetail `json:"alloc_details"`
}

type SlabRecord map[uint64]AllocRecord

func (loader *MemoryLoader) getSlabRec(infoMap *ebpf.Map, recs *map[string]SlabRecord) {
	var taskKey bpfTaskKey
	var taskInfo bpfTaskInfo
	infoIter := infoMap.Iterate()
	for infoIter.Next(&taskKey, &taskInfo) {
		slab := uint8ToString(taskInfo.Slab[:])
		rec := (*recs)[slab][taskKey.TgidPid]
		ips := make([]uint64, CallStackSize)
		if err := loader.objs.StackTrace.Lookup(taskInfo.StackId, &ips); err != nil {
			continue
		}

		allocDetail := AllocDetail{
			BytesAlloc:     int64(taskInfo.BytesAlloc),
			CallchainInsts: []string{},
		}
		/* skip first entry (duplicated) */
		for _, ip := range ips[1:] {
			if ip == 0 {
				break
			}
			symbol, err := loader.symbolizer.Symbolize(ip)
			if err != nil {
				symbol = ""
			}
			allocDetail.CallchainInsts = append(allocDetail.CallchainInsts, symbol)
		}
		rec.Comm = uint8ToString(taskInfo.Comm[:])
		rec.AllocDetails = append(rec.AllocDetails, allocDetail)
		if _, isExist := (*recs)[slab]; !isExist {
			(*recs)[slab] = make(map[uint64]AllocRecord)
		}
		(*recs)[slab][taskKey.TgidPid] = rec
	}
}

func (loader *MemoryLoader) writeToFile(outputPath string, bytes *[]byte) error {
	fp, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString(string(*bytes)); err != nil {
		return err
	}
	return nil
}

func (loader *MemoryLoader) writeSlabInfo(outputPath string, info *utils.SlabInfo) error {
	bytes, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return loader.writeToFile(outputPath, &bytes)
}

func (loader *MemoryLoader) writeSlabRec(outputPath string, recs *map[string]SlabRecord) error {
	bytes, err := json.Marshal(recs)
	if err != nil {
		return err
	}

	return loader.writeToFile(outputPath, &bytes)
}

func (loader *MemoryLoader) StoreData(outputPath string) error {
	slabInfoFile := outputPath + SlabInfoFilePostfix
	slabInfo, err := utils.GetSlabInfo()
	if err != nil {
		logrus.Errorf("Failed to get slab info, err [%s]", err)
		return err
	}
	if err := loader.writeSlabInfo(slabInfoFile, slabInfo); err != nil {
		logrus.Errorf("Failed to write slab info to file, err [%s]", err)
		return err
	}
	loader.outputFiles = append(loader.outputFiles, filepath.Base(slabInfoFile))

	slabRecFile := outputPath + SlabRecFilePostfix
	slabRecs := map[string]SlabRecord{}
	loader.getSlabRec(loader.objs.SlabInfo, &slabRecs)
	if err := loader.writeSlabRec(slabRecFile, &slabRecs); err != nil {
		logrus.Errorf("Failed to write slab records to file, err [%s]", err)
		return err
	}
	loader.outputFiles = append(loader.outputFiles, filepath.Base(slabRecFile))

	return nil
}

func (loader *MemoryLoader) GetOutputFiles() []string {
	return loader.outputFiles
}

func (loader *MemoryLoader) Close() {
	loader.objs.Close()
}
