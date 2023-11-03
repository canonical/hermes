package utils

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	systemLevelPath = "/proc/pressure/"
	PSISome         = "some"
	PSIFull         = "full"
	CpuPSI          = "cpu"
	MemoryPSI       = "memory"
	IOPSI           = "io"
	PSIAvg10        = "avg10"
	PSIAvg60        = "avg60"
	PSIAvg300       = "avg300"
)

type PSIType string

type PSIAvgs struct {
	Avg10  float64 `json:"avg10"`
	Avg60  float64 `json:"avg60"`
	Avg300 float64 `json:"avg300"`
}

func (psiAvgs *PSIAvgs) GetInterval(interval string) float64 {
	if interval == PSIAvg10 {
		return psiAvgs.Avg10
	} else if interval == PSIAvg60 {
		return psiAvgs.Avg60
	} else if interval == PSIAvg300 {
		return psiAvgs.Avg300
	}
	return 0 // unknown psi interval
}

type PSIResult struct {
	Some PSIAvgs `json:"some"`
	Full PSIAvgs `json:"full"`
}

type PSI struct{}

func (psi *PSI) parseAvgs(tokens []string, avgs *PSIAvgs) {
	for _, token := range tokens {
		vals := strings.Split(token, "=")
		val, err := strconv.ParseFloat(vals[1], 64)
		if err != nil {
			logrus.Errorf("Failed to parse avgs [%s], err [%s]", token, err)
			continue
		}
		if vals[0] == PSIAvg10 {
			avgs.Avg10 = val
		} else if vals[0] == PSIAvg60 {
			avgs.Avg60 = val
		} else if vals[0] == PSIAvg300 {
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
		if tokens[0] == PSISome {
			avgs = &psiResult.Some
		} else if tokens[0] == PSIFull {
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
	return psi.getResult(systemLevelPath + string(psiType))
}
