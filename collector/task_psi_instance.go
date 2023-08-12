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

type PSIContext struct {
	Type utils.PSIType
	Some []float64
	Full []float64
}

func (context *PSIContext) Check() error {
	if !common.Contains([]utils.PSIType{utils.CpuPSI, utils.MemoryPSI, utils.IOPSI}, context.Type) {
		return fmt.Errorf("Unrecognized type [%d]", context.Type)
	}
	if len(context.Some) != 3 {
		return fmt.Errorf("The length of some entry is not three")
	}
	if len(context.Full) != 3 {
		return fmt.Errorf("The length of full entry is not three")
	}
	return nil
}

type TaskPSIInstance struct{}

func NewTaskPSIInstance(_ common.TaskType) (TaskInstance, error) {
	return &TaskPSIInstance{}, nil
}

func (instance *TaskPSIInstance) isBeyondExpectation(psiAvgs *utils.PSIAvgs, expected []float64) bool {
	avgs := [...]float64{psiAvgs.Avg10, psiAvgs.Avg60, psiAvgs.Avg300}
	for idx, val := range expected {
		if val == -1 {
			continue
		}
		if idx >= len(avgs) {
			break
		}
		if avgs[idx] >= val {
			return true
		}
	}

	return false
}

func (instance *TaskPSIInstance) ToFile(psiResult *utils.PSIResult, logDataPath string) error {
	bytes, err := json.Marshal(psiResult)
	if err != nil {
		return err
	}
	fp, err := os.OpenFile(logDataPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString(string(bytes)); err != nil {
		return err
	}
	return nil
}

func (instance *TaskPSIInstance) GetLogDataPathPostfix() string {
	return ".psi"
}

func (instance *TaskPSIInstance) Process(instContext interface{}, logDataPathGenerator log.LogDataPathGenerator, result chan error) {
	psiContext := instContext.(*PSIContext)
	var err error
	defer func() {
		result <- err
	}()

	var psi utils.PSI
	psiResult, err := psi.GetSystemLevel(psiContext.Type)
	if err != nil {
		return
	}

	err = instance.ToFile(psiResult, logDataPathGenerator(".psi"))
	if err != nil {
		logrus.Errorf("Failed to write PSI result, err [%s]", err)
		return
	}

	if instance.isBeyondExpectation(&psiResult.Some, psiContext.Some) ||
		instance.isBeyondExpectation(&psiResult.Full, psiContext.Full) {
		err = nil
	} else {
		err = fmt.Errorf("PSI value does not exceed threshold")
	}
}
