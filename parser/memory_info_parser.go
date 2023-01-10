package parser

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"hermes/backend/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const MemTotal = "MemTotal"
const MemFree = "MemFree"

type MemoryInfoRecord struct {
	Thresholds map[string]int64 `json:"thresholds"`
	MemInfo    *utils.MemInfo   `json:"memInfo"`
}

type MemoryInfoParser struct{}

func GetMemoryInfoParser() (ParserInstance, error) {
	return &MemoryInfoParser{}, nil
}

func (parser *MemoryInfoParser) getMemoryInfoRecord(path string) (*MemoryInfoRecord, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rec MemoryInfoRecord
	if err := json.Unmarshal(bytes, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (parser *MemoryInfoParser) getCSVData(timestamp string, rec *MemoryInfoRecord) ([]string, error) {
	memTotal, isExist := (*rec.MemInfo)[MemTotal]
	if !isExist {
		return []string{}, fmt.Errorf("Entry [MemTotal] does not exist")
	}

	memFree, isExist := (*rec.MemInfo)[MemFree]
	if !isExist {
		return []string{}, fmt.Errorf("Entry [MemFree] does not exist")
	}

	percent, isExist := rec.Thresholds[MemFree]
	if !isExist {
		return []string{}, fmt.Errorf("Threshold [MemFree] does not exist")
	}

	return []string{timestamp, strconv.FormatInt(memTotal*percent/100, 10), strconv.FormatInt(memFree, 10)}, nil
}

func (parser *MemoryInfoParser) writeCSVData(csvData []string, path string) error {
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

func (parser *MemoryInfoParser) Parse(logDir string, logs []string, outputDir string) error {
	if len(logs) != 1 {
		return fmt.Errorf("Some logs may not be handled")
	}

	rec, err := parser.getMemoryInfoRecord(logDir + "/" + logs[0])
	if err != nil {
		return err
	}

	csvData, err := parser.getCSVData(filepath.Base(outputDir), rec)
	if err != nil {
		return err
	}

	err = parser.writeCSVData(csvData, filepath.Dir(outputDir)+string("/overview"))
	if err != nil {
		return err
	}
	return nil
}
