package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	systemLevelPath = "/proc/pressure/"
)

type PSIType uint32

const (
	CpuPSI PSIType = iota
	MemoryPSI
	IOPSI
)

type PSIAvgs struct {
	Avg10  float64
	Avg60  float64
	Avg300 float64
}

type PSIResult struct {
	Some PSIAvgs
	Full PSIAvgs
}

type PSI struct{}

func (psi *PSI) getEntry(psiType PSIType) (string, error) {
	switch psiType {
	case CpuPSI:
		return "cpu", nil
	case MemoryPSI:
		return "memory", nil
	case IOPSI:
		return "io", nil
	default:
		return "", fmt.Errorf("Unhandled PSI type [%d]", psiType)
	}
}

func (psi *PSI) parseAvgs(tokens []string, avgs *PSIAvgs) {
	for _, token := range tokens {
		vals := strings.Split(token, "=")
		val, err := strconv.ParseFloat(vals[1], 64)
		if err != nil {
			logrus.Errorf("Failed to parse avgs [%s], err [%s]", token, err)
			continue
		}
		if vals[0] == "avg10" {
			avgs.Avg10 = val
		} else if vals[0] == "avg60" {
			avgs.Avg60 = val
		} else if vals[0] == "avg300" {
			avgs.Avg300 = val
		}
	}
}

func (psi *PSI) getResult(path string) (*PSIResult, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	var psiResult PSIResult
	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		data := string(scanner.Bytes())
		tokens := strings.Split(data, " ")
		if len(tokens) != 5 {
			logrus.Errorf("Unexpected data format [%s]", data)
			continue
		}

		var avgs *PSIAvgs
		if tokens[0] == "some" {
			avgs = &psiResult.Some
		} else if tokens[0] == "full" {
			avgs = &psiResult.Full
		} else {
			logrus.Errorf("Unexpected token [%s]", tokens[0])
		}

		psi.parseAvgs(tokens[1:4], avgs)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &psiResult, nil
}

func (psi *PSI) GetSystemLevel(psiType PSIType) (*PSIResult, error) {
	entry, err := psi.getEntry(psiType)
	if err != nil {
		return nil, err
	}

	return psi.getResult(systemLevelPath + entry)
}
