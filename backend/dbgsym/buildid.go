package dbgsym

import (
	"fmt"
	"os"
	"path/filepath"

	"hermes/backend/elf"
	"hermes/common"
)

type CpuMode int

const (
	KernelMode CpuMode = iota
	UserMode
)

type BuildID struct {
	mode       CpuMode
	filePath   string
	outputDir  string
	getBuildID elf.GetBuildID
}

func NewBuildID(mode CpuMode, filePath, outputDir string) *BuildID {
	return &BuildID{
		mode:       mode,
		filePath:   filePath,
		outputDir:  outputDir,
		getBuildID: *elf.NewGetBuildID(),
	}
}

func (inst *BuildID) composePath(buildID, fileName string) string {
	return filepath.Join(inst.outputDir, buildID, fileName)
}

func (inst *BuildID) GetKernelPath(buildID string) string {
	return inst.composePath(buildID, "kallsyms")
}

func (inst *BuildID) GetUserPath(buildID string) string {
	return inst.composePath(buildID, "debuginfo")
}

func (inst *BuildID) buildKernel() (string, error) {
	buildID, err := inst.getBuildID.Kernel()
	if err != nil {
		return "", err
	}

	srcPath := "/proc/kallsyms"
	dstPath := inst.GetKernelPath(buildID)
	if _, err := os.Stat(dstPath); err == nil {
		return dstPath, err
	}

	return buildID, common.CopyFile(srcPath, dstPath)
}

func (inst *BuildID) buildUser() (string, error) {
	buildID, err := inst.getBuildID.File(inst.filePath)
	if err != nil {
		return "", err
	}
	dstPath := inst.GetUserPath(buildID)
	debugInfod := NewDebugInfod(dstPath)
	return buildID, debugInfod.DownloadDebugInfo(buildID)
}

func (inst *BuildID) Build() (string, error) {
	switch inst.mode {
	case KernelMode:
		return inst.buildKernel()
	case UserMode:
		return inst.buildUser()
	}
	return "", fmt.Errorf("Unhandled mode: %d", inst.mode)
}

func GetBuildIDByPath(path string) string {
	return filepath.Base(filepath.Dir(path))
}
