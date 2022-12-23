package ebpf

import (
	"encoding/json"
	"fmt"
	"hermes/backend/utils"
	"io/ioutil"
	"os"
)

type MemoryEbpfParser struct {
	Recs map[uint64]DataRecord
}

func GetMemoryEbpfParser() (*MemoryEbpfParser, error) {
	return &MemoryEbpfParser{
		Recs: make(map[uint64]DataRecord),
	}, nil
}

func (parser *MemoryEbpfParser) getStack(pid uint32, comm string, allocRec *AllocRecord) []string {
	stack := []string{}
	for _, inst := range allocRec.CallchainInsts {
		stack = append(stack, inst)
	}
	stack = append(stack, fmt.Sprintf("%s (%d)", comm, pid))
	return stack
}

func (parser *MemoryEbpfParser) writeStackCollapsedData(outputPath string) error {
	fp, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	flameGraphData := utils.GetFlameGraphData()

	for tgidPid, rec := range parser.Recs {
		for _, allocRec := range rec.AllocRecs {
			stack := parser.getStack(uint32(tgidPid), rec.Comm, &allocRec)
			flameGraphData.Add(&stack, len(stack)-1, 1)
		}
	}
	return flameGraphData.WriteToFile(outputPath)
}

func (parser *MemoryEbpfParser) Parse(logDir string, logs []string, outputDir string) error {
	for _, log := range logs {
		logPath := logDir + string("/") + log
		data, err := ioutil.ReadFile(logPath)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(data, &parser.Recs); err != nil {
			return err
		}

		outputPath := outputDir + string("/") + log + string(".stack.json")
		if err := parser.writeStackCollapsedData(outputPath); err != nil {
			return err
		}
	}

	return nil
}
