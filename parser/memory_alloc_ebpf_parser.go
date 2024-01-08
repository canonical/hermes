package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	memoryAlloc "hermes/backend/ebpf/memory_alloc"
	"hermes/backend/symbol"
	"hermes/backend/utils"
	"hermes/log"
)

type MemoryEbpfParser struct {
	dbgDirPath    string
	symbolizer    symbol.Symbolizer
	kernelBuildID string
}

const (
	UnrecordedLabel = "Unrecorded"
	RecordedLabel   = "Recorded"
	DbgDir          = ".hermes.memory_alloc_ebpf.dbg"
)

func GetMemoryAllocEbpfParser() (ParserInstance, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dbgDirPath := filepath.Join(homeDir, DbgDir)

	return &MemoryEbpfParser{
		dbgDirPath: dbgDirPath,
		symbolizer: *symbol.NewSymbolizer(dbgDirPath),
	}, nil
}

func (parser *MemoryEbpfParser) getSlabInfo(path string) (*utils.SlabInfo, error) {
	var slabInfo utils.SlabInfo

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &slabInfo); err != nil {
		return nil, err
	}
	return &slabInfo, nil
}

func (parser *MemoryEbpfParser) getSlabRec(path string) (*map[string]memoryAlloc.SlabRecord, error) {
	slabRec := make(map[string]memoryAlloc.SlabRecord)

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &slabRec); err != nil {
		return nil, err
	}
	return &slabRec, nil
}

func (parser *MemoryEbpfParser) getStacks(slabName string,
	allocRec *memoryAlloc.AllocRecord, flameGraphData *utils.FlameGraphData) int64 {
	var bytesObserved int64 = 0

	for _, allocDetail := range allocRec.AllocDetails {
		stack := []string{}
		for _, ip := range allocDetail.CallchainIps {
			_symbol := fmt.Sprintf("0x%x", ip)
			if __symbol, err := parser.symbolizer.Symbolize(symbol.KernelMode, parser.kernelBuildID, ip); err == nil {
				_symbol = __symbol
			}
			stack = append(stack, _symbol)
		}
		stack = append(stack, allocRec.Comm)
		stack = append(stack, RecordedLabel)
		stack = append(stack, slabName)
		flameGraphData.Add(&stack, len(stack)-1, allocDetail.BytesAlloc)
		bytesObserved = bytesObserved + allocDetail.BytesAlloc
	}
	return bytesObserved
}

func (parser *MemoryEbpfParser) parseStacks(slabName string,
	bytes int64, rec *memoryAlloc.SlabRecord, flameGraphData *utils.FlameGraphData) {
	for _, allocRec := range *rec {
		bytesObserved := parser.getStacks(slabName, &allocRec, flameGraphData)
		bytes = bytes - bytesObserved
	}
	if bytes > 0 {
		stack := []string{UnrecordedLabel, slabName}
		flameGraphData.Add(&stack, len(stack)-1, bytes)
	}
}

func (parser *MemoryEbpfParser) writeStackCollapsedData(
	slabInfo *utils.SlabInfo, slabRec *map[string]memoryAlloc.SlabRecord, path string) error {
	fp, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	flameGraphData := utils.NewFlameGraphData()

	for slabName, rec := range *slabRec {
		bytes, isExist := (*slabInfo)[slabName]
		if !isExist {
			bytes = 0
		}
		parser.parseStacks(slabName, bytes, &rec, flameGraphData)
	}
	return flameGraphData.WriteToFile(path)
}

func (parser *MemoryEbpfParser) Parse(logPathManager log.LogPathManager, timestamp int64, logDataPostfix, outputDir string) error {
	matches, err := filepath.Glob(logPathManager.DataPath(logDataPostfix))
	if err != nil {
		return err
	}

	kernSymPath := ""
	for _, filePath := range matches {
		if strings.HasSuffix(filePath, ".kern.sym") {
			kernSymPath = filePath
			break
		}
	}

	buildID, err := symbol.KernelSymPrepare(parser.dbgDirPath, kernSymPath)
	if err != nil {
		return err
	}
	parser.kernelBuildID = buildID

	var slabInfo *utils.SlabInfo = nil
	var slabRec *map[string]memoryAlloc.SlabRecord = nil
	for _, filePath := range matches {
		if strings.HasSuffix(filePath, ".kern.sym") {
			continue
		}
		var err error
		if strings.Contains(filePath, memoryAlloc.SlabInfoFilePostfix) {
			slabInfo, err = parser.getSlabInfo(filePath)
		} else if strings.Contains(filePath, memoryAlloc.SlabRecFilePostfix) {
			slabRec, err = parser.getSlabRec(filePath)
		} else {
			err = fmt.Errorf("Unexpected file path [%s]", filePath)
		}

		if err != nil {
			return err
		}
	}

	outputPath := filepath.Join(outputDir, strconv.FormatInt(timestamp, 10), ParsedPostfix[MemleakProfileJob])
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return err
	}
	return parser.writeStackCollapsedData(slabInfo, slabRec, outputPath)
}
