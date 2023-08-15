package plugin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

const (
	thresFieldName = "Threshold"
)

func handleCpuProfileResp(startTime, endTime int64, response []byte) ([]*data.Frame, error) {
	records := []CpuProfileRecord{}
	if err := json.Unmarshal(response, &records); err != nil {
		return nil, err
	}

	timeField := data.NewFieldFromFieldType(data.FieldTypeTime, 0)
	timeField.Name = data.TimeSeriesTimeFieldName
	thresField := data.NewFieldFromFieldType(data.FieldTypeUint64, 0)
	thresField.Name = data.TimeSeriesValueFieldName
	valField := data.NewFieldFromFieldType(data.FieldTypeUint64, 0)
	valField.Name = data.TimeSeriesValueFieldName

	thresFrames := data.NewFrame("threshold", timeField, thresField)
	valFrames := data.NewFrame("usage", timeField, valField)

	for _, rec := range records {
		if rec.Timestamp < startTime || rec.Timestamp > endTime {
			continue
		}
		timeField.Append(time.Unix(rec.Timestamp, 0))
		thresField.Append(rec.Threshold)
		valField.Append(rec.Val)
	}
	return []*data.Frame{thresFrames, valFrames}, nil
}

func HandleQueryResp(group, routine string, startTime, endTime int64, response []byte) ([]*data.Frame, error) {
	handlers := map[string]map[string]func(int64, int64, []byte) ([]*data.Frame, error){
		"cpu": {
			"cpu_profile": handleCpuProfileResp,
		},
	}
	if _, isExist := handlers[group]; !isExist {
		return nil, fmt.Errorf("Cannot find a handler for group %s, routine %s", group, routine)
	}
	handler, isExist := handlers[group][routine]
	if !isExist {
		return nil, fmt.Errorf("Cannot find a handler for group %s, routine %s", group, routine)
	}
	return handler(startTime, endTime, response)
}
