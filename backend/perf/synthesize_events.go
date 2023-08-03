package perf

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

type SynthesizeEvents struct {
	outputPath string
}

func NewSynthesizeEvents(outputPath string) (*SynthesizeEvents, error) {
	return &SynthesizeEvents{
		outputPath: outputPath,
	}, nil
}

func (events *SynthesizeEvents) synthesizeCommEvents(pid, tid int) error {
	path := filepath.Join("/proc", strconv.Itoa(tid), "status")
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	rec := CommRecord{
		Header: Header{
			Type: CommRec,
		},
		Pid: uint32(pid),
		Tid: uint32(tid),
	}
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		data := string(scanner.Bytes())
		token, val, found := strings.Cut(data, ":")
		if !found {
			logrus.Errorf("Unexpected data format [%s]", data)
			continue
		}
		if token == "Name" {
			rec.Comm = strings.TrimSpace(val)
			break
		}
	}

	if err := AppendToFile(&rec, events.outputPath); err != nil {
		return err
	}
	return nil
}

func (events *SynthesizeEvents) parseProcMapsLine(data string, rec *Mmap2Record) error {
	tokens := strings.Fields(data)
	if len(tokens) < 6 {
		return fmt.Errorf("Unexpected data format [%s]", data)
	}

	addr := strings.Split(tokens[0], "-")
	if len(addr) != 2 {
		return fmt.Errorf("Unexpected addr format [%s]", tokens[0])
	}
	start, err := strconv.ParseInt(addr[0], 16, 64)
	if err != nil {
		return err
	}
	end, err := strconv.ParseInt(addr[1], 16, 64)
	if err != nil {
		return err
	}
	rec.Addr = uint64(start)
	rec.Len = uint64(end)

	for _, c := range tokens[1] {
		if c == '-' {
			continue
		}
		if c == 'r' {
			rec.Prot |= syscall.PROT_READ
		} else if c == 'w' {
			rec.Prot |= syscall.PROT_WRITE
		} else if c == 'x' {
			rec.Prot |= syscall.PROT_EXEC
		} else if c == 's' {
			rec.Flags = syscall.MAP_SHARED
		} else if c == 'p' {
			rec.Flags = syscall.MAP_PRIVATE
		}
	}

	offset, err := strconv.ParseInt(tokens[2], 16, 64)
	if err != nil {
		return err
	}
	rec.Pgoff = uint64(offset)

	dev := strings.Split(tokens[3], ":")
	if len(dev) != 2 {
		return fmt.Errorf("Unexpected dev format [%s]", tokens[3])
	}
	maj, err := strconv.ParseInt(dev[0], 16, 64)
	if err != nil {
		return err
	}
	min, err := strconv.ParseInt(dev[1], 16, 64)
	if err != nil {
		return err
	}
	rec.MajorID = uint32(maj)
	rec.MinorID = uint32(min)

	ino, err := strconv.ParseInt(tokens[4], 10, 64)
	if err != nil {
		return err
	}
	rec.Ino = uint64(ino)

	var filename string
	for i := 5; i < len(tokens); i++ {
		if len(filename) != 0 {
			filename += " "
		}
		filename += tokens[i]
	}
	rec.Filename = filename
	return nil
}

func (events *SynthesizeEvents) synthesizeMmap2Events(pid, tid int) error {
	if pid != tid {
		return nil
	}
	path := filepath.Join("/proc", strconv.Itoa(tid), "maps")
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()

	rec := Mmap2Record{
		Header: Header{
			Type: MmapRec,
		},
		Pid: uint32(pid),
		Tid: uint32(tid),
	}
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		if err := events.parseProcMapsLine(string(scanner.Bytes()), &rec); err != nil {
			continue
		}
		if err := AppendToFile(&rec, events.outputPath); err != nil {
			logrus.Errorf("Failed to append record to file [%s], err [%s]",
				events.outputPath, err)
		}
	}
	return nil
}

func (events *SynthesizeEvents) Synthesize() error {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		path := filepath.Join("/proc", entry.Name(), "task")
		_entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, _entry := range _entries {
			tid, err := strconv.Atoi(_entry.Name())
			if err != nil {
				return err
			}
			if err := events.synthesizeCommEvents(pid, tid); err != nil {
				logrus.Errorf("Failed to synthesize comm events, err [%s]", err)
			}
			if err := events.synthesizeMmap2Events(pid, tid); err != nil {
				logrus.Errorf("Failed to synthesize mmap events, err [%s]", err)
			}
		}
	}
	return nil
}
