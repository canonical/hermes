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

type CpuInfoContext struct {
	Threshold uint64
	Usage     uint64
	Triggered bool
}

func (context *CpuInfoContext) Fill(param, paramOverride *[]byte) error {
	return common.FillContext(param, paramOverride, context)
}

type TaskCpuInfoInstance struct{}

func NewCpuInfoInstance(_ common.TaskType) (TaskInstance, error) {
	return &TaskCpuInfoInstance{}, nil
}

func (instance *TaskCpuInfoInstance) isBeyondExpectation(context *CpuInfoContext) bool {
	return context.Usage >= context.Threshold
}

func (instance *TaskCpuInfoInstance) writeToFile(context *CpuInfoContext, path string) error {
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

func (instance *TaskCpuInfoInstance) GetLogDataPathPostfix(instContext interface{}) string {
	return ".cpuinfo"
}

func (instance *TaskCpuInfoInstance) Process(instContext interface{}, logDataPathGenerator log.LogDataPathGenerator, result chan error) {
	cpuInfoContext := instContext.(*CpuInfoContext)
	var err error
	defer func() {
		result <- err
	}()

	cpuInfoContext.Usage, err = utils.GetCpuUsage()
	if err != nil {
		logrus.Errorf("Failed to get cpu usage, set zero as default value")
		cpuInfoContext.Usage = 0
	}

	cpuInfoContext.Triggered = instance.isBeyondExpectation(cpuInfoContext)
	if cpuInfoContext.Triggered {
		err = nil
	} else {
		err = fmt.Errorf("CpuInfo value does not exceed threshold")
	}

	logDataPath := logDataPathGenerator(".cpuinfo")
	if instance.writeToFile(cpuInfoContext, logDataPath) != nil {
		logrus.Errorf("Failed to write to file [%s], err [%s]", logDataPath, err)
	}
}
