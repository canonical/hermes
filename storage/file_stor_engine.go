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

func (engine *FileStorEngine) loadFile(file string) ([]log.LogMetadata, error) {
	var logMetas []log.LogMetadata
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(bytes, &logMetas); err != nil {
		return nil, err
	}
	return logMetas, nil
}

func (engine *FileStorEngine) Save(timestamp int64, logMeta log.LogMetadata) error {
	logMetaPath := filepath.Join(engine.logDir, LogMetaDirName, strconv.FormatInt(timestamp, 10))
	var metasToWrite []log.LogMetadata

	//collect any existing entries if metadata file already exists
	if _, err := os.Stat(logMetaPath); !os.IsNotExist(err) {
		metasToWrite, err = engine.loadFile(logMetaPath)
		if err != nil {
			return err
		}
	}

	//append new entry and write out
	metasToWrite = append(metasToWrite, logMeta)

	bytes, err := yaml.Marshal(metasToWrite)
	if err != nil {
		return err
	}

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

func (engine *FileStorEngine) Load() (map[int64][]log.LogMetadata, error) {
	logMetas := map[int64][]log.LogMetadata{}
	matches, err := filepath.Glob(filepath.Join(engine.logDir, LogMetaDirName, "*"))
	if err != nil {
		return nil, err
	}

	for _, file := range matches {
		logMeta, err := engine.loadFile(file)
		if err != nil {
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
