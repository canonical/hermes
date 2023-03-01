package collector

import (
	"encoding/json"
	"fmt"
	"os"

	"hermes/backend/utils"
	"hermes/parser"

	"github.com/sirupsen/logrus"
)

const PSITask = "psi"

type PSIContext struct {
	Type utils.PSIType
	Some []float64
	Full []float64
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

func (instance *TaskPSIInstance) ToFile(psiResult *utils.PSIResult, outputPath string) error {
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

func (instance *TaskPSIInstance) GetParserType(instContext interface{}) parser.ParserType {
	return parser.None
}

func (instance *TaskPSIInstance) Process(instContext interface{}, outputPath string, result chan TaskResult) {
	psiContext := instContext.(*PSIContext)
	taskResult := TaskResult{}
	var err error
	defer func() {
		taskResult.Err = err
		result <- taskResult
	}()

	var psi utils.PSI
	psiResult, err := psi.GetSystemLevel(psiContext.Type)
	if err != nil {
		return
	}
	err = instance.ToFile(psiResult, outputPath+string(".psi"))
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
