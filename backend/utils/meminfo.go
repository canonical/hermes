package utils

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

const MemInfoEntry = "/proc/meminfo"

type MemInfo struct {
	Infos map[string]uint64
}

func GetMemInfo() (*MemInfo, error) {
	file, err := os.Open(MemInfoEntry)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	memInfo := MemInfo{
		Infos: make(map[string]uint64),
	}
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
		memInfo.Infos[tokens[0]] = percent
	}

	return &memInfo, nil
}

func (memInfo *MemInfo) ToFile(path string) error {
	bytes, err := json.Marshal(*memInfo)
	if err != nil {
		return err
	}
	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString(string(bytes)); err != nil {
		return err
	}
	return nil
}
