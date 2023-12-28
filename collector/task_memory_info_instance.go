package collector

import (
	"encoding/json"
	"fmt"
	"os"

	"hermes/backend/utils"
	"hermes/common"
	"hermes/log"

	"github.com/sirupsen/logrus"
)

const MemTotal = "MemTotal"

type MemoryInfoContext struct {
	Thresholds map[string]int64
	MemInfo    *utils.MemInfo
	Triggered  bool
}

func (context *MemoryInfoContext) Fill(param, paramOverride *[]byte) error {
	return common.FillContext(param, paramOverride, context)
}

type TaskMemoryInfoInstance struct{}

func NewMemoryInfoInstance(_ common.TaskType) (TaskInstance, error) {
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

func (instance *TaskMemoryInfoInstance) GetLogDataPathPostfix(instContext interface{}) string {
	return ".meminfo"
}

func (instance *TaskMemoryInfoInstance) Process(instContext interface{}, logPathManager log.LogPathManager, result chan error) {
	memoryInfoContext := instContext.(*MemoryInfoContext)
	var err error
	defer func() {
		result <- err
	}()

	memoryInfoContext.MemInfo, err = utils.GetMemInfo()
	if err != nil {
		logrus.Errorf("Failed to get meminfo, err [%s]", err)
		return
	}

	memoryInfoContext.Triggered = instance.isBeyondExpectation(memoryInfoContext)
	if memoryInfoContext.Triggered {
		err = nil
	} else {
		err = fmt.Errorf("MemInfo value does not exceed thresholds")
	}

	logDataPath := logPathManager.DataPath(".meminfo")
	if err := instance.writeToFile(memoryInfoContext, logDataPath); err != nil {
		logrus.Errorf("Failed to write to file [%s], err [%s]", logDataPath, err)
	}
}
