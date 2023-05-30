package common

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const EnvFile = "hermes.env"

func LoadEnv(metaDir string) error {
	envPath := filepath.Join(metaDir, EnvFile)

	err := godotenv.Load(envPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
