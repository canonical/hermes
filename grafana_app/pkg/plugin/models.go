package plugin

import (
	"encoding/json"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type Query struct {
	Group   string `json:"group"`
	Routine string `json:"routine"`
}

func ParseQuery(query backend.DataQuery) (Query, error) {
	_query := Query{}

	if err := json.Unmarshal(query.JSON, &_query); err != nil {
		return _query, err
	}
	return _query, nil
}

type ProfileRecord struct {
	Timestamp int64  `json:"timestamp"`
	Threshold uint64 `json:"threshold"`
	Val       uint64 `json:"val"`
	Triggered bool   `json:"triggered"`
}
