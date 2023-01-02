package utils

import (
	"math"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

func GetCpuUsage() (uint64, error) {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	return uint64(math.Ceil(percent[0])), nil
}
