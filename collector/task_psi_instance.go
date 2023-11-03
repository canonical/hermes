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

type PSIThresholds utils.PSIResult

type PSIContext struct {
	Type        utils.PSIType    `json:"psi_type"`
	Thresholds  *PSIThresholds   `json:"thresholds"`
	Levels      *utils.PSIResult `json:"levels"`
	Triggered   bool             `json:"triggered"`
	TriggeredBy string           `json:"triggered_by"`
}

func (context *PSIContext) check() error {
	if !common.Contains([]utils.PSIType{utils.CpuPSI, utils.MemoryPSI, utils.IOPSI}, context.Type) {
		return fmt.Errorf("Unrecognized type [%s]", context.Type)
	}
	return nil
}

func (context *PSIContext) Fill(param, paramOverride *[]byte) error {
	if err := common.FillContext(param, paramOverride, context); err != nil {
		return err
	}
	return context.check()
}

type TaskPSIInstance struct{}

func NewTaskPSIInstance(_ common.TaskType) (TaskInstance, error) {
	return &TaskPSIInstance{}, nil
}

func (instance *TaskPSIInstance) isBeyondExpectation(psiAvgs *utils.PSIAvgs, expected *utils.PSIAvgs) string {
	if psiAvgs.Avg10 >= expected.Avg10 {
		return utils.PSIAvg10
	} else if psiAvgs.Avg60 >= expected.Avg10 {
		return utils.PSIAvg60
	} else if psiAvgs.Avg300 >= expected.Avg300 {
		return utils.PSIAvg300
	}
	return ""
}

func (instance *TaskPSIInstance) ToFile(psiContext *PSIContext, logDataPath string) error {
	bytes, err := json.Marshal(psiContext)
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

func (instance *TaskPSIInstance) GetLogDataPathPostfix(instContext interface{}) string {
	return ".psi"
}

func (instance *TaskPSIInstance) Process(instContext interface{}, logDataPathGenerator log.LogDataPathGenerator, result chan error) {
	psiContext := instContext.(*PSIContext)
	var err error
	defer func() {
		result <- err
	}()

	var psi utils.PSI
	psiContext.Levels, err = psi.GetSystemLevel(psiContext.Type)
	if err != nil {
		return
	}

	var interval string
	interval = instance.isBeyondExpectation(&psiContext.Levels.Some, &psiContext.Thresholds.Some)
	if interval != "" {
		psiContext.Triggered = true
		psiContext.TriggeredBy = utils.PSISome + "/" + interval
	} else {
		interval = instance.isBeyondExpectation(&psiContext.Levels.Full, &psiContext.Thresholds.Full)
		if interval != "" {
			psiContext.Triggered = true
			psiContext.TriggeredBy = utils.PSIFull + "/" + interval
		}
	}
	err = instance.ToFile(psiContext, logDataPathGenerator(".psi"))

	if err != nil {
		logrus.Errorf("Failed to write PSI result, err [%s]", err)
		return
	}

	if psiContext.Triggered {
		err = nil
	} else {
		err = fmt.Errorf("PSI value does not exceed threshold")
	}
}
