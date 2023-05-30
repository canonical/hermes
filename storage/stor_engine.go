package storage

import (
	"fmt"

	"hermes/log"
)

type StorEngineType string

var (
	storEngineMap = map[string]func(string) (StorEngine, error){
		"file": GetFileStorEngine,
	}
)

const (
	File StorEngineType = "file"
)

type StorEngine interface {
	Save(timestamp int64, logMeta log.LogMetadata) error
	Load() (map[int64]log.LogMetadata, error)
}

func GetStorEngine(storEngine, logDir string) (StorEngine, error) {
	getStorEngineFunc, isExist := storEngineMap[storEngine]
	if isExist {
		return getStorEngineFunc(logDir)
	}
	return nil, fmt.Errorf("Unhandled storage engine [%s]", storEngine)
}
