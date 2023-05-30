package parser

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"hermes/backend/perf"
	"hermes/backend/utils"
	"hermes/log"
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

func (parser *CpuProfileParser) Parse(logDataPathGenerator log.LogDataPathGenerator, timestamp int64, logDataPostfix, outputDir string) error {
	flameGraphData := utils.GetFlameGraphData()
	matches, err := filepath.Glob(logDataPathGenerator(logDataPostfix))
	if err != nil {
		return err
	}

	for _, filePath := range matches {
		if err := parser.parseStackCollapsedData(filePath, flameGraphData); err != nil {
			return err
		}
	}

	outputPath := filepath.Join(outputDir, strconv.FormatInt(timestamp, 10), "overall_cpu.stack.json")
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return err
	}
	return flameGraphData.WriteToFile(outputPath)
}
