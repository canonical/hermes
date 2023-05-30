package collector

import (
	"os"

	"hermes/common"
	"hermes/log"
)

type BinaryContext struct {
	Cmds []string
}

type TaskBinaryInstance struct{}

func NewTaskBinaryInstance(_ common.TaskType) (TaskInstance, error) {
	return &TaskBinaryInstance{}, nil
}

func (instance *TaskBinaryInstance) GetLogDataPathPostfix() string {
	return ".binary"
}

func (instance *TaskBinaryInstance) Process(instContext interface{}, logDataPathGenerator log.LogDataPathGenerator, result chan error) {
	binaryContext := instContext.(*BinaryContext)
	var err error
	defer func() {
		result <- err
	}()

	logDataPath := logDataPathGenerator(".binary")
	env := map[string]string{
		"OUTPUT_FILE": logDataPath,
	}
	cmd := PrepareCmd(binaryContext.Cmds, env)
	outfile, err := os.Create(logDataPath)
	if err != nil {
		return
	}
	defer outfile.Close()

	cmd.Stdout = outfile
	err = cmd.Run()
}
