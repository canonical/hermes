package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/yukariatlas/hermes/backend/utils"
)

type PSIContext struct {
	Type utils.PSIType `json:"type"`
	Some []float64     `json:"some"`
	Full []float64     `json:"full"`
}

type TaskPSIInstance struct{}

func NewTaskPSIInstance() (TaskInstance, error) {
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

func (instance *TaskPSIInstance) writeToFile(psiResult *utils.PSIResult, outputPath string) error {
	bytes, err := json.Marshal(psiResult)
	if err != nil {
		return err
	}
	fp, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString(string(bytes)); err != nil {
		return err
	}
	return nil
}

func (instance *TaskPSIInstance) Process(param, paramOverride, outputPath string, finish chan error) {
	psiContext := PSIContext{}
	err := errors.New("")
	defer func() { finish <- err }()

	err = json.Unmarshal([]byte(param), &psiContext)
	if err != nil {
		logrus.Errorf("Failed to unmarshal json, param [%s]", param)
		return
	}
	if paramOverride != "" {
		err = json.Unmarshal([]byte(paramOverride), &psiContext)
		if err != nil {
			logrus.Errorf("Failed to unmarshal json, paramOverride [%s]", paramOverride)
			return
		}
	}

	var psi utils.PSI
	psiResult, err := psi.GetSystemLevel(psiContext.Type)
	if err != nil {
		return
	}
	if err = instance.writeToFile(psiResult, outputPath+string(".psi")); err != nil {
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
