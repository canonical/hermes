package log

import (
	"os"
	"path/filepath"
)

type LogDataPathGenerator func(string) string

const LogDataDirName = "data"

func PrepareLogDataDir(logDir string) error {
	logDataDir := filepath.Join(logDir, LogDataDirName)
	if err := os.MkdirAll(logDataDir, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func GetLogDataPathGenerator(logDir, logDataLabel string) LogDataPathGenerator {
	return func(postfix string) string {
		return filepath.Join(logDir, LogDataDirName, logDataLabel+postfix)
	}
}
