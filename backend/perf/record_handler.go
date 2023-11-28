package perf

import (
	"encoding/json"

	"hermes/backend/utils"
)

type RecordHandler struct {
	flameGraphData utils.FlameGraphData
}

func GetRecordHandler() *RecordHandler {
	return &RecordHandler{
		flameGraphData: *utils.GetFlameGraphData(),
	}
}

func (handler *RecordHandler) parseSampleRec(bytes []byte) error {
	var rec SampleRecord
	if err := json.Unmarshal(bytes, &rec); err != nil {
		return err
	}
	stack := []string{}
	for _, inst := range rec.CallchainInsts {
		if inst.Symbol == "" {
			continue
		}
		stack = append(stack, inst.Symbol)
	}
	handler.flameGraphData.Add(&stack, len(stack)-1, 1)
	return nil
}

func (handler *RecordHandler) Parse(bytes []byte) error {
	var header Header
	if err := json.Unmarshal(bytes, &header); err != nil {
		return err
	}

	switch header.Type {
	case SampleRec:
		return handler.parseSampleRec(bytes)
	}
	return nil
}

func (handler *RecordHandler) GetFlameGraphData() *utils.FlameGraphData {
	return &handler.flameGraphData
}
