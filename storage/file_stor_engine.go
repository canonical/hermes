package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"hermes/log"

	"gopkg.in/yaml.v2"
)

const LogMetaDirName = "metadata"

type FileStorEngine struct {
	logDir string
}

func GetFileStorEngine(logDir string) (StorEngine, error) {
	logMetaDir := filepath.Join(logDir, LogMetaDirName)
	if err := os.MkdirAll(logMetaDir, os.ModePerm); err != nil {
		return nil, err
	}
	return &FileStorEngine{
		logDir: logDir,
	}, nil
}

func (engine *FileStorEngine) Save(timestamp int64, logMeta log.LogMetadata) error {
	bytes, err := yaml.Marshal(logMeta)
	if err != nil {
		return err
	}

	logMetaPath := filepath.Join(engine.logDir, LogMetaDirName, strconv.FormatInt(timestamp, 10))
	fp, err := os.OpenFile(logMetaPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString(string(bytes)); err != nil {
		return err
	}
	return nil
}

func (engine *FileStorEngine) Load() (map[int64]log.LogMetadata, error) {
	logMetas := map[int64]log.LogMetadata{}
	matches, err := filepath.Glob(filepath.Join(engine.logDir, LogMetaDirName, "*"))
	if err != nil {
		return nil, err
	}

	for _, file := range matches {
		var logMeta log.LogMetadata
		bytes, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(bytes, &logMeta); err != nil {
			return nil, err
		}
		timestamp, err := strconv.ParseInt(filepath.Base(file), 10, 64)
		if err != nil {
			return nil, err
		}
		logMetas[timestamp] = logMeta
	}
	return logMetas, nil
}
