package collector

import (
	"encoding/json"
	"os"

	"hermes/parser"

	"github.com/sirupsen/logrus"
)

type BinaryContext struct {
	Cmds []string `json:"cmds"`
}

type TaskBinaryInstance struct{}

func NewTaskBinaryInstance() (TaskInstance, error) {
	return &TaskBinaryInstance{}, nil
}

func (instance *TaskBinaryInstance) Process(param, paramOverride, outputPath string, result chan TaskResult) {
	context := BinaryContext{}
	taskResult := TaskResult{
		Err:         nil,
		ParserType:  parser.None,
		OutputFiles: []string{},
	}
	err := taskResult.Err
	defer func() { result <- taskResult }()

	err = json.Unmarshal([]byte(param), &context)
	if err != nil {
		logrus.Errorf("Failed to unmarshal json, param [%s]", param)
		return
	}
	if paramOverride != "" {
		err = json.Unmarshal([]byte(paramOverride), &context)
		if err != nil {
			logrus.Errorf("Failed to unmarshal json, paramOverride [%s]", paramOverride)
			return
		}
	}

	env := map[string]string{
		"OUTPUT_FILE": outputPath,
	}
	cmd := PrepareCmd(context.Cmds, env)
	outfile, err := os.Create(outputPath)
	if err != nil {
		return
	}
	defer outfile.Close()

	cmd.Stdout = outfile
	err = cmd.Run()
}
