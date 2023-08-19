package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"hermes/collector"
	"hermes/log"
)

const MemTotal = "MemTotal"
const MemFree = "MemFree"

type MemoryInfoRecord struct {
	Timestamp int64 `json:"timestamp"`
	Threshold int64 `json:"threshold"`
	Val       int64 `json:"val"`
	Triggered bool  `json:"triggered"`
}

type MemoryInfoParser struct{}

func GetMemoryInfoParser() (ParserInstance, error) {
	return &MemoryInfoParser{}, nil
}

func (parser *MemoryInfoParser) getMemoryInfoRecord(timestamp int64, path string) (*MemoryInfoRecord, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var context collector.MemoryInfoContext
	if err := json.Unmarshal(bytes, &context); err != nil {
		return nil, err
	}

	memTotal, isExist := (*context.MemInfo)[collector.MemTotal]
	if !isExist {
		return nil, fmt.Errorf("Entry [MemTotal] does not exist")
	}

	memFree, isExist := (*context.MemInfo)[MemFree]
	if !isExist {
		return nil, fmt.Errorf("Entry [MemFree] does not exist")
	}

	percent, isExist := context.Thresholds[MemFree]
	if !isExist {
		return nil, fmt.Errorf("Threshold [MemFree] does not exist")
	}

	return &MemoryInfoRecord{
		Timestamp: timestamp,
		Threshold: memTotal * percent / 100,
		Val:       memFree,
		Triggered: context.Triggered,
	}, nil
}

func (parser *MemoryInfoParser) writeJSONData(rec *MemoryInfoRecord, path string) error {
	var recs []MemoryInfoRecord
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

func (parser *MemoryInfoParser) Parse(logDataPathGenerator log.LogDataPathGenerator, timestamp int64, logDataPostfix, outputDir string) error {
	rec, err := parser.getMemoryInfoRecord(timestamp, logDataPathGenerator(logDataPostfix))
	if err != nil {
		return err
	}

	err = parser.writeJSONData(rec, outputDir+string("/overview"))
	if err != nil {
		return err
	}
	return nil
}
