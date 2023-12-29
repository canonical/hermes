package symbol

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unsafe"

	"hermes/backend/dbgsym"
	"hermes/common"

	"github.com/sirupsen/logrus"
)

type KsymParser struct {
	dbgDirPath string
}

func NewKsymParser(dbgDirPath string) *KsymParser {
	return &KsymParser{
		dbgDirPath: dbgDirPath,
	}
}

func unsafeString(bytes []byte) string {
	return *((*string)(unsafe.Pointer(&bytes)))
}

type IteratorCallback func(uint64, string, string) error

var ErrFound = errors.New("Found")

func (inst *KsymParser) iterator(filePath string, callback IteratorCallback) error {
	fp, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		addr, err := strconv.ParseUint(unsafeString(bytes[:16]), 16, 64)
		if err != nil {
			logrus.Errorf("Failed to parse kallsym bytes, err [%s]", err)
			continue
		}

		symbolType := string(bytes[17:18])
		symbolEndIdx := len(bytes)
		for i := 19; i < len(bytes); i++ {
			if bytes[i] == ' ' {
				symbolEndIdx = i
				break
			}
		}
		symbol := string(bytes[19:symbolEndIdx])
		if err = callback(addr, symbolType, symbol); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (inst *KsymParser) getFunctionAddr(filePath, sym string) (uint64, error) {
	var addr uint64 = 0
	err := inst.iterator(filePath, func(_addr uint64, symbolType, symbol string) error {
		_symbolType := strings.ToUpper(symbolType)
		if (_symbolType == "T" || _symbolType == "W" || symbolType == "A") && sym == symbol {
			addr = _addr
			return ErrFound
		}
		return nil
	})
	if err == ErrFound {
		err = nil
	}
	return addr, err
}

func (inst *KsymParser) getSymbolAddr(filePath, sym string) (uint64, error) {
	var addr uint64 = 0
	err := inst.iterator(filePath, func(_addr uint64, symbolType, symbol string) error {
		if sym == symbol {
			addr = _addr
			return ErrFound
		}
		return nil
	})
	if err == ErrFound {
		err = nil
	}
	return addr, err
}

func (inst *KsymParser) GetMapRange(_buildID string) (uint64, uint64, error) {
	var start, end, addr uint64
	var err error
	filePath := dbgsym.NewBuildID(dbgsym.KernelMode, "", inst.dbgDirPath).GetKernelPath(_buildID)
	startSyms := []string{"_text", "_stext"}

	for _, startSym := range startSyms {
		addr, err = inst.getFunctionAddr(filePath, startSym)
		if err == nil {
			break
		}
	}
	if err != nil {
		return 0, 0, err
	}
	start = addr

	addr, err = inst.getSymbolAddr(filePath, "_edata")
	if err != nil {
		addr, err = inst.getFunctionAddr(filePath, "_etext")
	}
	if err != nil {
		return 0, 0, err
	}
	end = addr
	return start, end, err
}

func (inst *KsymParser) Resolve(buildID string, addr uint64) string {
	filePath := dbgsym.NewBuildID(dbgsym.KernelMode, "", inst.dbgDirPath).GetKernelPath(buildID)
	sym := ""
	inst.iterator(filePath, func(_addr uint64, symbolType, symbol string) error {
		if _addr > addr {
			return ErrFound
		}
		sym = symbol
		return nil
	})
	return sym
}

func KernelSymPrepare(dbgDirPath, kernSymPath string) (string, error) {
	buildID := dbgsym.NewBuildID(dbgsym.KernelMode, "", dbgDirPath)
	_buildID := ""
	if _, err := os.Stat(kernSymPath); err == nil {
		_kernSymPath, err := filepath.EvalSymlinks(kernSymPath)
		if err != nil {
			return "", err
		}

		_buildID = dbgsym.GetBuildIDByPath(_kernSymPath)
		dbgFile := buildID.GetKernelPath(_buildID)
		if err := common.CopyFile(_kernSymPath, dbgFile); err != nil {
			return "", err
		}
	} else {
		logrus.Warn("The kernel symbol doesn't exist. Use /proc/kallsyms instead")
		__buildID, err := buildID.Build()
		if err != nil {
			return "", err
		}
		_buildID = __buildID
	}
	return _buildID, nil
}
