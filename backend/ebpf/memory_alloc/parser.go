package ebpf

import (
	"encoding/json"
	"fmt"
	"hermes/backend/utils"
	"io/ioutil"
	"os"
	"strings"
)

type MemoryEbpfParser struct{}

const UnrecordedLabel = "Unrecorded"
const RecordedLabel = "Recorded"

func GetMemoryEbpfParser() (*MemoryEbpfParser, error) {
	return &MemoryEbpfParser{}, nil
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

func (parser *MemoryEbpfParser) getSlabRec(path string) (*map[string]SlabRecord, error) {
	slabRec := make(map[string]SlabRecord)

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &slabRec); err != nil {
		return nil, err
	}
	return &slabRec, nil
}

func (parser *MemoryEbpfParser) getStacks(slabName string, pid uint32,
	allocRec *AllocRecord, flameGraphData *utils.FlameGraphData) uint64 {
	var bytesObserved uint64 = 0

	for _, allocDetail := range allocRec.AllocDetails {
		stack := []string{}
		for _, inst := range allocDetail.CallchainInsts {
			stack = append(stack, inst)
		}
		stack = append(stack, fmt.Sprintf("%s (%d)", allocRec.Comm, pid))
		stack = append(stack, RecordedLabel)
		stack = append(stack, slabName)
		flameGraphData.Add(&stack, len(stack)-1, int(allocDetail.BytesAlloc))
		bytesObserved = bytesObserved + allocDetail.BytesAlloc
	}
	return bytesObserved
}

func (parser *MemoryEbpfParser) parseStacks(slabName string,
	bytes uint64, rec *SlabRecord, flameGraphData *utils.FlameGraphData) {
	for tgidPid, allocRec := range *rec {
		bytesObserved := parser.getStacks(slabName, uint32(tgidPid), &allocRec, flameGraphData)
		bytes = bytes - bytesObserved
	}
	if bytes > 0 {
		stack := []string{UnrecordedLabel, slabName}
		flameGraphData.Add(&stack, len(stack)-1, int(bytes))
	}
}

func (parser *MemoryEbpfParser) writeStackCollapsedData(
	slabInfo *utils.SlabInfo, slabRec *map[string]SlabRecord, path string) error {
	fp, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	flameGraphData := utils.GetFlameGraphData()

	for slabName, rec := range *slabRec {
		bytes, isExist := (*slabInfo)[slabName]
		if !isExist {
			bytes = 0
		}
		parser.parseStacks(slabName, bytes, &rec, flameGraphData)
	}
	return flameGraphData.WriteToFile(path)
}

func (parser *MemoryEbpfParser) Parse(logDir string, logs []string, outputDir string) error {
	if len(logs) != 2 {
		return fmt.Errorf("Unexpected number of logs, count [%d]", len(logs))
	}

	var slabInfo *utils.SlabInfo = nil
	var slabRec *map[string]SlabRecord = nil
	for _, log := range logs {
		var err error
		path := logDir + "/" + log
		if strings.Contains(log, SlabInfoFilePostfix) {
			slabInfo, err = parser.getSlabInfo(path)
		} else if strings.Contains(log, SlabRecFilePostfix) {
			slabRec, err = parser.getSlabRec(path)
		} else {
			err = fmt.Errorf("Unexpected log name [%s]", log)
		}

		if err != nil {
			return err
		}
	}

	outputPath := outputDir + "/slab.stack.json"
	if err := parser.writeStackCollapsedData(slabInfo, slabRec, outputPath); err != nil {
		return err
	}
	return nil
}
