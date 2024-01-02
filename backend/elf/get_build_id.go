package elf

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"hermes/common"
	"os"
	"strings"
)

const (
	GNUBuildID     = 3
	BuildIDSize    = 20
	SysKernelNotes = "/sys/kernel/notes"
)

type GetBuildID struct{}

func NewGetBuildID() *GetBuildID {
	return &GetBuildID{}
}

func (inst *GetBuildID) noteAlign(size int) int {
	const align = 4
	return (size + align - 1) & ^(align - 1)
}

func (inst *GetBuildID) parseBuildID(data []byte) (string, error) {
	var hdr struct {
		NameSize uint32
		DescSize uint32
		Type     uint32
	}

	for offset := 0; offset < len(data); {
		if err := binary.Read(bytes.NewReader(data[offset:]), common.NativeEndian(), &hdr); err != nil {
			return "", err
		}

		const gnu = "GNU"
		nameSizeAligned, descSizeAligned := inst.noteAlign(int(hdr.NameSize)), inst.noteAlign(int(hdr.DescSize))
		offset += binary.Size(hdr)
		name := string(data[offset : offset+nameSizeAligned])
		offset += nameSizeAligned
		if (hdr.Type == GNUBuildID) && (nameSizeAligned == len(gnu)+1) {
			if strings.HasPrefix(name, gnu) {
				size := BuildIDSize
				if size > descSizeAligned {
					size = descSizeAligned
				}
				return fmt.Sprintf("%x", data[offset:offset+size]), nil
			}
		}
		offset += descSizeAligned
	}
	return "", nil
}

func (inst *GetBuildID) File(filePath string) (string, error) {
	fp, err := elf.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	sections := []string{".note.gnu.build-id", ".notes", ".note"}
	for _, section := range sections {
		sec := fp.Section(section)
		if sec == nil {
			continue
		}
		data, err := sec.Data()
		if err != nil {
			return "", err
		}
		if buildID, err := inst.parseBuildID(data); err != nil {
			return "", err
		} else if buildID != "" {
			return buildID, nil
		}
	}
	return "", nil
}

func (inst *GetBuildID) Kernel() (string, error) {
	fp, err := os.Open(SysKernelNotes)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	fileInfo, err := fp.Stat()
	if err != nil {
		return "", err
	}

	buf := make([]byte, fileInfo.Size())
	if _, err := fp.Read(buf); err != nil {
		return "", err
	}
	return inst.parseBuildID(buf)
}
