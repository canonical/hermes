package collector

import (
	"fmt"
	"os"

	"hermes/common"
	"hermes/log"
)

type BinaryContext struct {
	Cmds []string
}

func (context *BinaryContext) check() error {
	if len(context.Cmds) == 0 {
		return fmt.Errorf("The cmds cannot be empty")
	}
	return nil
}

func (context *BinaryContext) Fill(param, paramOverride *[]byte) error {
	if err := common.FillContext(param, paramOverride, context); err != nil {
		return err
	}
	return context.check()
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
