package utils

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

const MemInfoEntry = "/proc/meminfo"
const SlabInfoEntry = "/proc/slabinfo"

type MemInfo map[string]int64
type SlabInfo map[string]int64

func GetMemInfo() (*MemInfo, error) {
	file, err := os.Open(MemInfoEntry)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	memInfo := make(MemInfo)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		text := scanner.Text()
		tokens := strings.Split(text, ":")
		if len(tokens) != 2 {
			continue
		}
		val := strings.TrimSpace(strings.TrimRight(tokens[1], "kB"))
		percent, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			continue
		}
		memInfo[tokens[0]] = percent
	}

	return &memInfo, nil
}

func GetSlabInfo() (*SlabInfo, error) {
	file, err := os.Open(SlabInfoEntry)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	slabInfo := make(SlabInfo)
	scanner := bufio.NewScanner(file)

	/* Skip header */
	scanner.Scan()

	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Fields(text)
		if len(fields) < 4 {
			continue
		}

		numObjs, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			continue
		}

		objSize, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			continue
		}
		slabInfo[fields[0]] = numObjs * objSize
	}

	return &slabInfo, nil
}
