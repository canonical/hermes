package symbol

import (
	"bufio"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/golang/groupcache/lru"
	"github.com/sirupsen/logrus"
)

type KsymRecord struct {
	addr   uint64
	symbol string
}

type KsymCache struct {
	mutex       *sync.RWMutex
	ksymRecords []KsymRecord
	cache       *lru.Cache
}

const KallsymsProcEntry = "/proc/kallsyms"
const LRUCacheSize = 128

func NewKsymCache() (*KsymCache, error) {
	ksymCache := KsymCache{
		mutex:       &sync.RWMutex{},
		ksymRecords: []KsymRecord{},
		cache:       lru.New(LRUCacheSize),
	}

	if err := ksymCache.LoadSymbols(); err != nil {
		return nil, err
	}
	return &ksymCache, nil
}

func unsafeString(bytes []byte) string {
	return *((*string)(unsafe.Pointer(&bytes)))
}

func (cache *KsymCache) LoadSymbols() error {
	fp, err := os.Open(KallsymsProcEntry)
	if err != nil {
		return err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		data := scanner.Bytes()
		addr, err := strconv.ParseUint(unsafeString(data[:16]), 16, 64)
		if err != nil {
			logrus.Errorf("Failed to parse kallsym data, err [%s]", err.Error())
			continue
		}

		symbolEndIdx := len(data)
		for i := 19; i < len(data); i++ {
			if data[i] == ' ' {
				symbolEndIdx = i
				break
			}
		}

		if symbolType := strings.ToLower(string(data[17:18])); symbolType == "b" || symbolType == "d" || symbolType == "r" {
			continue
		}
		cache.ksymRecords = append(cache.ksymRecords, KsymRecord{addr: addr, symbol: string(data[19:symbolEndIdx])})
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	sort.Slice(cache.ksymRecords, func(i, j int) bool { return cache.ksymRecords[i].addr < cache.ksymRecords[j].addr })
	return nil
}

func (cache *KsymCache) searchSymbol(addr uint64) string {
	idx := sort.Search(len(cache.ksymRecords), func(i int) bool { return addr < cache.ksymRecords[i].addr })
	if idx < len(cache.ksymRecords) && idx > 0 {
		return cache.ksymRecords[idx-1].symbol
	}
	return ""
}

func (cache *KsymCache) Resolve(addr uint64) (string, error) {
	cache.mutex.RLock()
	if symbol, ok := cache.cache.Get(addr); ok {
		cache.mutex.RUnlock()
		return symbol.(string), nil
	}
	cache.mutex.RUnlock()

	symbol := cache.searchSymbol(addr)
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.cache.Add(addr, symbol)
	return symbol, nil
}
