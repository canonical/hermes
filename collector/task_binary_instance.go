package collector

import (
	"os"

	"hermes/parser"
)

const BinaryTask = "binary"

type BinaryContext struct {
	Cmds []string
}

type TaskBinaryInstance struct{}

func NewTaskBinaryInstance() (TaskInstance, error) {
	return &TaskBinaryInstance{}, nil
}

func (instance *TaskBinaryInstance) GetParserType(instContext interface{}) parser.ParserType {
	return parser.None
}

func (instance *TaskBinaryInstance) Process(instContext interface{}, outputPath string, result chan TaskResult) {
	binaryContext := instContext.(*BinaryContext)
	taskResult := TaskResult{}
	var err error
	defer func() {
		taskResult.Err = err
		result <- taskResult
	}()

	env := map[string]string{
		"OUTPUT_FILE": outputPath,
	}
	cmd := PrepareCmd(binaryContext.Cmds, env)
	outfile, err := os.Create(outputPath)
	if err != nil {
		return
	}
	defer outfile.Close()

	cmd.Stdout = outfile
	err = cmd.Run()
}
