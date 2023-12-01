package utils

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const (
	DebugInfodURL = "https://debuginfod.ubuntu.com"
	InfoDir       = "/tmp/.hermes.debuginfo/"
	Buildid       = "buildid"
	DebugInfo     = "debuginfo"
)

type DebugInfod struct {
	url     string
	infoDir string
}

func NewDebugInfod() *DebugInfod {
	return &DebugInfod{
		url:     DebugInfodURL,
		infoDir: InfoDir,
	}
}

func (debugInfod *DebugInfod) DownloadDebugInfo(file string) error {
	buildID, err := NewBuildID(file).Get()
	if err != nil {
		return err
	}
	debugInfoPath := filepath.Join(debugInfod.infoDir, buildID, DebugInfo)
	if _, err := os.Stat(debugInfoPath); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(debugInfoPath), os.ModePerm); err != nil {
		return err
	}
	url, err := url.JoinPath(debugInfod.url, Buildid, buildID, DebugInfo)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fp, err := os.Create(debugInfoPath)
	if err != nil {
		return err
	}
	defer fp.Close()
	fp.ReadFrom(resp.Body)
	return nil
}
