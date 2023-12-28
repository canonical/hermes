package log

import (
	"os"
	"path/filepath"
)

type LogPathManager struct {
	logDir       string
	logDataLabel string
}

const (
	LogDataDirName   = "data"
	LogMetaDirName   = "metadata"
	LogDbgsymDirName = "dbgsym"
)

func NewLogPathManager(logDir string) *LogPathManager {
	return &LogPathManager{
		logDir: logDir,
	}
}

func (inst *LogPathManager) Prepare() error {
	logDataDir := filepath.Join(inst.logDir, LogDataDirName)
	if err := os.MkdirAll(logDataDir, os.ModePerm); err != nil {
		return err
	}
	logMetaDir := filepath.Join(inst.logDir, LogMetaDirName)
	if err := os.MkdirAll(logMetaDir, os.ModePerm); err != nil {
		return err
	}
	logDbgsymDir := filepath.Join(inst.logDir, LogDbgsymDirName)
	if err := os.MkdirAll(logDbgsymDir, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (inst *LogPathManager) SetDataLabel(logDataLabel string) *LogPathManager {
	inst.logDataLabel = logDataLabel
	return inst
}

func (inst *LogPathManager) DataPath(postfix string) string {
	return filepath.Join(inst.logDir, LogDataDirName, inst.logDataLabel+postfix)
}

func (inst *LogPathManager) MetadataPath() string {
	return filepath.Join(inst.logDir, LogMetaDirName)
}

func (inst *LogPathManager) DbgsymPath() string {
	return filepath.Join(inst.logDir, LogDbgsymDirName)
}
