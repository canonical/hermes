package utils

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

const MemInfoEntry = "/proc/meminfo"

type MemInfo map[string]uint64

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
		percent, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			continue
		}
		memInfo[tokens[0]] = percent
	}

	return &memInfo, nil
}
