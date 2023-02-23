package collector

import (
	"encoding/json"
	"fmt"
	"hermes/backend/utils"
	"hermes/parser"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const CpuInfoTask = "cpu_info"

type CpuInfoContext struct {
	Threshold uint64
	Usage     uint64
}

type TaskCpuInfoInstance struct{}

func NewCpuInfoInstance() (TaskInstance, error) {
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

func (instance *TaskCpuInfoInstance) Process(instContext interface{}, outputPath string, result chan TaskResult) {
	cpuInfoContext := instContext.(*CpuInfoContext)
	taskResult := TaskResult{
		Err:         nil,
		ParserType:  parser.CpuInfo,
		OutputFiles: []string{},
	}
	var err error
	defer func() {
		taskResult.Err = err
		result <- taskResult
	}()

	cpuInfoContext.Usage, err = utils.GetCpuUsage()
	if err != nil {
		logrus.Errorf("Failed to get cpu usage, set zero as default value")
		cpuInfoContext.Usage = 0
	}
	if instance.isBeyondExpectation(cpuInfoContext) {
		err = nil
	} else {
		err = fmt.Errorf("CpuInfo value does not exceed threshold")
	}

	cpuCondFile := outputPath + ".cond"
	if err := instance.writeToFile(cpuInfoContext, cpuCondFile); err != nil {
		logrus.Errorf("Failed to write to file [%s], err [%s]", cpuCondFile, err)
	}
	taskResult.OutputFiles = append(taskResult.OutputFiles, filepath.Base(cpuCondFile))
}
