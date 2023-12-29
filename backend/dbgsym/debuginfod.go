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
	url     string
	dstPath string
}

func NewDebugInfod(dstPath string) *DebugInfod {
	return &DebugInfod{
		url:     DebugInfodURL,
		dstPath: dstPath,
	}
}

func (inst *DebugInfod) DownloadDebugInfo(buildID string) error {
	if _, err := os.Stat(inst.dstPath); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(inst.dstPath), os.ModePerm); err != nil {
		return err
	}
	url, err := url.JoinPath(inst.url, Buildid, buildID, DebugInfo)
	if err != nil {
		return err
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fp, err := os.Create(inst.dstPath)
	if err != nil {
		return err
	}
	defer fp.Close()
	fp.ReadFrom(resp.Body)
	return nil
}
