package parser

import (
	"bufio"
	"encoding/json"
	"hermes/backend/perf"
	"hermes/backend/utils"
	"os"
)

type CpuProfileParser struct{}

func GetCpuProfileParser() (ParserInstance, error) {
	return &CpuProfileParser{}, nil
}

func (parser *CpuProfileParser) getStacks(rec *perf.SampleRecord, flameGraphData *utils.FlameGraphData) {
	stack := []string{}
	for _, inst := range rec.CallchainInsts {
		if inst.Symbol == "" {
			continue
		}
		stack = append(stack, inst.Symbol)
	}
	flameGraphData.Add(&stack, len(stack)-1, 1)
}

func (parser *CpuProfileParser) parseStackCollapsedData(logPath string, flameGraphData *utils.FlameGraphData) error {
	fp, err := os.Open(logPath)
	if err != nil {
		return err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		var header perf.Header
		bytes := scanner.Bytes()
		if err := json.Unmarshal(bytes, &header); err != nil {
			return err
		}
		if header.Type != perf.SampleRec {
			continue
		}

		var rec perf.SampleRecord
		if err := json.Unmarshal(bytes, &rec); err != nil {
			return err
		}
		parser.getStacks(&rec, flameGraphData)
	}
	return nil
}

func (parser *CpuProfileParser) Parse(logDir string, logs []string, outputDir string) error {
	flameGraphData := utils.GetFlameGraphData()
	for _, log := range logs {
		logPath := logDir + "/" + log
		if err := parser.parseStackCollapsedData(logPath, flameGraphData); err != nil {
			return err
		}
	}

	outputPath := outputDir + "/overall_cpu.stack.json"
	return flameGraphData.WriteToFile(outputPath)
}
