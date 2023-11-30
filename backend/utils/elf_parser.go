package utils

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
	GNUBuildID  = 3
	BuildIDSize = 20
)

type BuildID struct {
	FilePath string
}

func NewBuildID(filepath string) *BuildID {
	return &BuildID{
		FilePath: filepath,
	}
}

func (buildID *BuildID) noteAlign(size int) int {
	const align = 4
	return (size + align - 1) & ^(align - 1)
}

func (buildID *BuildID) parseBuildID(buf []byte) (string, error) {
	for offset := 0; offset < len(buf); {
		var hdr struct {
			NameSize uint32
			DescSize uint32
			Type     uint32
		}

		if err := binary.Read(bytes.NewReader(buf[offset:]), common.NativeEndian(), &hdr); err != nil {
			return "", err
		}

		const gnu = "GNU"
		nameSizeAligned := buildID.noteAlign(int(hdr.NameSize))
		descSizeAligned := buildID.noteAlign(int(hdr.DescSize))
		offset += binary.Size(hdr)
		name := string(buf[offset : offset+nameSizeAligned])
		offset += nameSizeAligned

		if (hdr.Type == GNUBuildID) && (nameSizeAligned == len(gnu)+1) {
			if strings.HasPrefix(name, gnu) {
				size := BuildIDSize
				if size > descSizeAligned {
					size = descSizeAligned
				}
				return fmt.Sprintf("%x", buf[offset:offset+size]), nil
			}
		}
		offset += descSizeAligned
	}

	return "", nil
}

func (buildID *BuildID) getBuildID32(fp *os.File, fileEndian binary.ByteOrder) (string, error) {
	var header elf.Header32
	if err := binary.Read(fp, fileEndian, &header); err != nil {
		return "", err
	}

	buf := make([]byte, header.Phentsize*header.Phnum)
	if _, err := fp.Seek(int64(header.Phoff), os.SEEK_SET); err != nil {
		return "", err
	}
	if err := binary.Read(fp, fileEndian, &buf); err != nil {
		return "", err
	}

	var prog elf.Prog32
	progSize := binary.Size(prog)
	for i := 0; i < int(header.Phnum); i = i + 1 {
		progData := buf[i*progSize : (i+1)*progSize]
		if err := binary.Read(bytes.NewReader(progData), common.NativeEndian(), &prog); err != nil {
			return "", err
		}

		if prog.Type != uint32(elf.PT_NOTE) {
			continue
		}

		if _, err := fp.Seek(int64(prog.Off), os.SEEK_SET); err != nil {
			return "", err
		}
		_buf := make([]byte, prog.Filesz)
		if err := binary.Read(fp, fileEndian, &_buf); err != nil {
			return "", err
		}
		_buildID, err := buildID.parseBuildID(_buf)
		if err != nil {
			return "", err
		}
		if _buildID != "" {
			return _buildID, nil
		}
	}
	return "", fmt.Errorf("Build ID not found")
}

func (buildID *BuildID) getBuildID64(fp *os.File, fileEndian binary.ByteOrder) (string, error) {
	var header elf.Header64
	if err := binary.Read(fp, fileEndian, &header); err != nil {
		return "", err
	}

	buf := make([]byte, header.Phentsize*header.Phnum)
	if _, err := fp.Seek(int64(header.Phoff), os.SEEK_SET); err != nil {
		return "", err
	}
	if err := binary.Read(fp, fileEndian, &buf); err != nil {
		return "", err
	}

	var prog elf.Prog64
	progSize := binary.Size(prog)
	for i := 0; i < int(header.Phnum); i += 1 {
		progData := buf[i*progSize : (i+1)*progSize]
		if err := binary.Read(bytes.NewReader(progData), common.NativeEndian(), &prog); err != nil {
			return "", err
		}

		if prog.Type != uint32(elf.PT_NOTE) {
			continue
		}

		if _, err := fp.Seek(int64(prog.Off), os.SEEK_SET); err != nil {
			return "", err
		}
		_buf := make([]byte, prog.Filesz)
		if err := binary.Read(fp, fileEndian, &_buf); err != nil {
			return "", err
		}
		_buildID, err := buildID.parseBuildID(_buf)
		if err != nil {
			return "", err
		}
		if _buildID != "" {
			return _buildID, nil
		}
	}
	return "", fmt.Errorf("Build ID not found")
}

func (buildID *BuildID) getIdent(fp *os.File) ([]byte, error) {
	ident := make([]byte, elf.EI_NIDENT)
	if _, err := fp.Read(ident); err != nil {
		return nil, err
	}

	if elfMagic := []byte(elf.ELFMAG); !bytes.Equal(ident[:len(elfMagic)], elfMagic) {
		return nil, fmt.Errorf("The magic number doesn't match")
	}
	if ident[elf.EI_VERSION] != byte(elf.EV_CURRENT) {
		return nil, fmt.Errorf("The version doesn't match")
	}

	if _, err := fp.Seek(0, os.SEEK_SET); err != nil {
		return nil, err
	}
	return ident, nil
}

func (buildID *BuildID) getEndian(ident []byte) binary.ByteOrder {
	if ident[elf.EI_DATA] == byte(elf.ELFDATA2LSB) {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func (buildID *BuildID) Get() (string, error) {
	if buildID.FilePath == "" {
		return "", fmt.Errorf("The file path is empty")
	}

	fp, err := os.Open(buildID.FilePath)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	ident, err := buildID.getIdent(fp)
	if err != nil {
		return "", err
	}

	fileEndian := buildID.getEndian(ident)
	if ident[elf.EI_CLASS] == byte(elf.ELFCLASS32) {
		return buildID.getBuildID32(fp, fileEndian)
	}
	return buildID.getBuildID64(fp, fileEndian)
}
