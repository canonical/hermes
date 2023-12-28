package dbgsym

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"hermes/backend/elf"
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

func (inst *BuildID) getKernel() (string, error) {
	buildID, err := inst.getBuildID.Kernel()
	if err != nil {
		return "", err
	}

	fpSrc, err := os.Open("/proc/kallsyms")
	if err != nil {
		return "", err
	}
	defer fpSrc.Close()

	dstPath := filepath.Join(inst.outputDir, buildID, "kallsyms")
	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		return "", err
	}
	fpDst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer fpDst.Close()

	_, err = io.Copy(fpDst, fpSrc)
	return dstPath, err
}

func (inst *BuildID) getUser() (string, error) {
	buildID, err := inst.getBuildID.File(inst.filePath)
	if err != nil {
		return "", err
	}
	return NewDebugInfod(inst.outputDir).DownloadDebugInfo(buildID)
}

func (inst *BuildID) Get() (string, error) {
	switch inst.mode {
	case KernelMode:
		return inst.getKernel()
	case UserMode:
		return inst.getUser()
	}
	return "", fmt.Errorf("Unhandled mode: %d", inst.mode)
}
