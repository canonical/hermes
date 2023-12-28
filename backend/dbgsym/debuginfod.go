package dbgsym

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const (
	DebugInfodURL = "https://debuginfod.ubuntu.com"
	Buildid       = "buildid"
	DebugInfo     = "debuginfo"
)

type DebugInfod struct {
	url       string
	outputDir string
}

func NewDebugInfod(outputDir string) *DebugInfod {
	return &DebugInfod{
		url:       DebugInfodURL,
		outputDir: outputDir,
	}
}

func (debugInfod *DebugInfod) DownloadDebugInfo(buildID string) error {
	debugInfoPath := filepath.Join(debugInfod.outputDir, buildID, DebugInfo)
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
