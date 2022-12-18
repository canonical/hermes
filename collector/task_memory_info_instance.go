package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"hermes/backend/utils"
	"hermes/parser"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const MemTotal = "MemTotal"

type MemoryInfoContext struct {
	Thresholds map[string]uint64 `json:"thresholds"`
}

type TaskMemoryInfoInstance struct{}

func NewMemoryInfoInstance() (TaskInstance, error) {
	return &TaskMemoryInfoInstance{}, nil
}

func (instance *TaskMemoryInfoInstance) isBeyondExpectation(memInfo *utils.MemInfo, context *MemoryInfoContext) bool {
	memTotal, isExist := memInfo.Infos[MemTotal]
	if !isExist {
		return false
	}

	for entry, percent := range context.Thresholds {
		val, isExist := memInfo.Infos[entry]
		if !isExist {
			continue
		}
		if val <= memTotal*percent/100 {
			return true
		}
	}

	return false
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

	memInfo, err := utils.GetMemInfo()
	if err != nil {
		logrus.Errorf("Failed to get meminfo, err [%s]", err)
		return
	}

	err = memInfo.ToFile(outputPath)
	if err != nil {
		logrus.Errorf("Failed to write meminfo to file [%s], err [%s]", outputPath, err)
		return
	}
	taskResult.OutputFiles = append(taskResult.OutputFiles, filepath.Base(outputPath))

	if instance.isBeyondExpectation(memInfo, &memoryInfoContext) {
		err = nil
	} else {
		err = fmt.Errorf("MemInfo value does not exceed thresholds")
	}
}
