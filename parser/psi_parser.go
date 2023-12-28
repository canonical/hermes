package parser

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"hermes/backend/utils"
	"hermes/collector"
	"hermes/log"
)

type PSIRecord struct {
	Timestamp   int64         `json:"timestamp"`
	Type        utils.PSIType `json:"psi_type"`
	TriggeredBy string        `json:"triggered_by"`
	Threshold   float64       `json:"threshold"`
	Val         float64       `json:"val"`
	Triggered   bool          `json:"triggered"`
}

type PSIParser struct{}

func GetPSIParser() (ParserInstance, error) {
	return &PSIParser{}, nil
}

func (parser *PSIParser) getPSIRecord(timestamp int64, path string) (*PSIRecord, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var context collector.PSIContext
	if err := json.Unmarshal(bytes, &context); err != nil {
		return nil, err
	}

	ret := &PSIRecord{
		Timestamp:   timestamp,
		Type:        context.Type,
		TriggeredBy: context.TriggeredBy,
		Triggered:   context.Triggered,
	}

	if context.Triggered {
		triggerMetric := strings.Split(context.TriggeredBy, "/")[0]
		triggerInterval := strings.Split(context.TriggeredBy, "/")[1]
		if triggerMetric == utils.PSISome {
			ret.Threshold = context.Thresholds.Some.GetInterval(triggerInterval)
			ret.Val = context.Levels.Some.GetInterval(triggerInterval)
		} else { // triggerMetric == utils.PSIFull
			ret.Threshold = context.Thresholds.Full.GetInterval(triggerInterval)
			ret.Val = context.Levels.Full.GetInterval(triggerInterval)
		}
	}
	return ret, nil
}

func (parser *PSIParser) writeJSONData(rec *PSIRecord, path string) error {
	var recs []PSIRecord
	if _, err := os.Stat(path); err == nil {
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(bytes, &recs); err != nil {
			return err
		}
	}

	recs = append(recs, *rec)
	bytes, err := json.Marshal(&recs)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bytes, 0644)
}

func (parser *PSIParser) Parse(logPathManager log.LogPathManager, timestamp int64, logDataPostfix, outputDir string) error {
	rec, err := parser.getPSIRecord(timestamp, logPathManager.DataPath(logDataPostfix))
	if err != nil {
		return err
	}

	err = parser.writeJSONData(rec, outputDir+string("/overview"))
	if err != nil {
		return err
	}
	return nil
}
