package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"hermes/backend/utils"
	"hermes/parser"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const MemTotal = "MemTotal"

type MemoryInfoContext struct {
	Thresholds map[string]uint64 `json:"thresholds"`
	MemInfo    *utils.MemInfo    `json:"memInfo"`
	Triggered  bool              `json:"triggered"`
}

type TaskMemoryInfoInstance struct{}

func NewMemoryInfoInstance() (TaskInstance, error) {
	return &TaskMemoryInfoInstance{}, nil
}

func (instance *TaskMemoryInfoInstance) isBeyondExpectation(context *MemoryInfoContext) bool {
	memTotal, isExist := (*context.MemInfo)[MemTotal]
	if !isExist {
		return false
	}

	for entry, percent := range context.Thresholds {
		val, isExist := (*context.MemInfo)[entry]
		if !isExist {
			continue
		}
		if val <= memTotal*percent/100 {
			return true
		}
	}

	return false
}

func (instance *TaskMemoryInfoInstance) writeToFile(context *MemoryInfoContext, path string) error {
	bytes, err := json.Marshal(*context)
	if err != nil {
		return err
	}
	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString(string(bytes)); err != nil {
		return err
	}
	return nil
}

func (instance *TaskMemoryInfoInstance) Process(param, paramOverride, outputPath string, result chan TaskResult) {
	memoryInfoContext := MemoryInfoContext{}
	taskResult := TaskResult{
		Err:         nil,
		ParserType:  parser.MemoryInfo,
		OutputFiles: []string{},
	}
	err := errors.New("")
	defer func() {
		taskResult.Err = err
		result <- taskResult
	}()

	err = json.Unmarshal([]byte(param), &memoryInfoContext)
	if err != nil {
		logrus.Errorf("Failed to unmarshal json, param [%s]", param)
		return
	}
	if paramOverride != "" {
		err = json.Unmarshal([]byte(paramOverride), &memoryInfoContext)
		if err != nil {
			logrus.Errorf("Failed to unmarshal json, paramOverride [%s]", paramOverride)
			return
		}
	}

	memoryInfoContext.MemInfo, err = utils.GetMemInfo()
	if err != nil {
		logrus.Errorf("Failed to get meminfo, err [%s]", err)
		return
	}

	memoryInfoContext.Triggered = instance.isBeyondExpectation(&memoryInfoContext)
	if memoryInfoContext.Triggered {
		err = nil
	} else {
		err = fmt.Errorf("MemInfo value does not exceed thresholds")
	}

	if err := instance.writeToFile(&memoryInfoContext, outputPath); err != nil {
		logrus.Errorf("Failed to write to file [%s], err [%s]", outputPath, err)
	}
	taskResult.OutputFiles = append(taskResult.OutputFiles, filepath.Base(outputPath))
}
