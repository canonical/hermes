package parser

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"

	"hermes/log"
)

type CpuInfoRecord struct {
	Threshold uint64 `json:"threshold"`
	Usage     uint64 `json:"usage"`
}

type CpuInfoParser struct{}

func GetCpuInfoParser() (ParserInstance, error) {
	return &CpuInfoParser{}, nil
}

func (parser *CpuInfoParser) getCpuInfoRecord(path string) (*CpuInfoRecord, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rec CpuInfoRecord
	if err := json.Unmarshal(bytes, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (parser *CpuInfoParser) getCSVData(timestamp int64, rec *CpuInfoRecord) ([]string, error) {
	return []string{strconv.FormatInt(timestamp, 10), strconv.FormatUint(rec.Threshold, 10), strconv.FormatUint(rec.Usage, 10)}, nil
}

func (parser *CpuInfoParser) writeCSVData(csvData []string, path string) error {
	var needHeader = false

	if _, err := os.Stat(path); err != nil {
		needHeader = true
	}
	fp, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(fp)

	defer func() {
		writer.Flush()
		fp.Close()
	}()

	if needHeader {
		header := []string{"timestamp", "threshold", "val"}
		writer.Write(header)
	}
	writer.Write(csvData)
	return nil
}

func (parser *CpuInfoParser) Parse(logDataPathGenerator log.LogDataPathGenerator, timestamp int64, logDataPostfix, outputDir string) error {
	rec, err := parser.getCpuInfoRecord(logDataPathGenerator(logDataPostfix))
	if err != nil {
		return err
	}

	csvData, err := parser.getCSVData(timestamp, rec)
	if err != nil {
		return err
	}

	err = parser.writeCSVData(csvData, outputDir+string("/overview"))
	if err != nil {
		return err
	}
	return nil
}
