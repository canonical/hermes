package collector

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/sirupsen/logrus"
)

type BinaryContext struct {
	Cmds []string `json:"cmds"`
}

type TaskBinaryInstance struct{}

func NewTaskBinaryInstance() (TaskInstance, error) {
	return &TaskBinaryInstance{}, nil
}

func (instance *TaskBinaryInstance) Execute(content string, outputPath string, finish chan error) {
	context := BinaryContext{}
	err := errors.New("")
	defer func() { finish <- err }()

	err = json.Unmarshal([]byte(content), &context)
	if err != nil {
		logrus.Errorf("Failed to unmarshal json, content [%s]", content)
		return
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
