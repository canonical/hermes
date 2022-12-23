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

type MemoryCollection struct {
	Thresholds map[string]uint64 `json:"thresholds"`
	MemInfo    *utils.MemInfo    `json:"memInfo"`
	Triggered  bool              `json:"triggered"`
}

type MemoryInfoParser struct{}

func GetMemoryInfoParser() (ParserInstance, error) {
	return &MemoryInfoParser{}, nil
}

func (parser *MemoryInfoParser) getMemoryCollection(path string) (*MemoryCollection, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var collection MemoryCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, err
	}
	return &collection, nil
}

func (parser *MemoryInfoParser) getCSVData(timestamp string, collection *MemoryCollection) ([]string, error) {
	memTotal, isExist := (*collection.MemInfo)[MemTotal]
	if !isExist {
		return []string{}, fmt.Errorf("Entry [MemTotal] does not exist")
	}

	memFree, isExist := (*collection.MemInfo)[MemFree]
	if !isExist {
		return []string{}, fmt.Errorf("Entry [MemFree] does not exist")
	}

	percent, isExist := collection.Thresholds[MemFree]
	if !isExist {
		return []string{}, fmt.Errorf("Threshold [MemFree] does not exist")
	}

	return []string{timestamp, strconv.FormatUint(memTotal*percent/100, 10), strconv.FormatUint(memFree, 10)}, nil
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

	collection, err := parser.getMemoryCollection(logDir + "/" + logs[0])
	if err != nil {
		return err
	}

	csvData, err := parser.getCSVData(filepath.Base(outputDir), collection)
	if err != nil {
		return err
	}

	err = parser.writeCSVData(csvData, filepath.Dir(outputDir)+string("/overview"))
	if err != nil {
		return err
	}
	return nil
}
