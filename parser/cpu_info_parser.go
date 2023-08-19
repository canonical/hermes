package parser

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"hermes/collector"
	"hermes/log"
)

type CpuInfoRecord struct {
	Timestamp int64  `json:"timestamp"`
	Threshold uint64 `json:"threshold"`
	Val       uint64 `json:"val"`
	Triggered bool   `json:"triggered"`
}

type CpuInfoParser struct{}

func GetCpuInfoParser() (ParserInstance, error) {
	return &CpuInfoParser{}, nil
}

func (parser *CpuInfoParser) getCpuInfoRecord(timestamp int64, path string) (*CpuInfoRecord, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var context collector.CpuInfoContext
	if err := json.Unmarshal(bytes, &context); err != nil {
		return nil, err
	}
	return &CpuInfoRecord{
		Timestamp: timestamp,
		Threshold: context.Threshold,
		Val:       context.Usage,
		Triggered: context.Triggered,
	}, nil
}

func (parser *CpuInfoParser) writeJSONData(rec *CpuInfoRecord, path string) error {
	var recs []CpuInfoRecord
	if _, err := os.Stat(path); err == nil {
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(bytes, &recs); err != nil {
			return err
		}
	}

	recs = append(recs, *rec)
	bytes, err := json.Marshal(&recs)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bytes, 0644)
}

func (parser *CpuInfoParser) Parse(logDataPathGenerator log.LogDataPathGenerator, timestamp int64, logDataPostfix, outputDir string) error {
	rec, err := parser.getCpuInfoRecord(timestamp, logDataPathGenerator(logDataPostfix))
	if err != nil {
		return err
	}

	err = parser.writeJSONData(rec, outputDir+string("/overview"))
	if err != nil {
		return err
	}
	return nil
}
