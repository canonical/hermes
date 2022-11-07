package collector

import (
	"encoding/json"
	"errors"
	"os"
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
