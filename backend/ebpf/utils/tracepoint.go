package ebpf

import (
	"os"
	"path/filepath"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/sirupsen/logrus"
)

const (
	TraceFsDir = "/sys/kernel/tracing/events"
)

func isTracepointExist(group, name string) bool {
	dirPath := filepath.Join(TraceFsDir, group, name)
	if _, err := os.Stat(dirPath); err != nil {
		if !os.IsNotExist(err) {
			logrus.Errorf("Failed to check existence, tracepoint [%s/%s], err [%s]", group, name, err)
		}
		return false
	}
	return true
}

func Tracepoint(group, name string, prog *ebpf.Program) (*link.Link, error) {
	if !isTracepointExist(group, name) {
		return nil, nil
	}
	tp, err := link.Tracepoint(group, name, prog, nil)
	return &tp, err
}

func Close(link *link.Link) {
	if link != nil {
		(*link).Close()
	}
}
