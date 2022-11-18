package perf

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const (
	CallingThread = 0
	AllThreads    = -1
)

const AllCPUs = -1

type PerfEvent struct {
	fd             int
	attr           *Attr
	ringBufHandler *RingBufHandler
}

func NewPerfEvent(attr *Attr, pid, cpu int) (*PerfEvent, error) {
	var event PerfEvent

	if err := event.Open(attr, pid, cpu); err != nil {
		return nil, err
	}
	return &event, nil
}

func (event *PerfEvent) open(attr *Attr, pid, cpu, flags int) error {
	if event.fd > 0 {
		return nil
	}

	flags |= unix.PERF_FLAG_FD_CLOEXEC
	fd, err := unix.PerfEventOpen(attr.ToUnixPerfEventAttr(), pid, cpu, -1, flags)
	if err != nil {
		return os.NewSyscallError("perf_event_open", err)
	}

	if err := unix.SetNonblock(fd, true); err != nil {
		unix.Close(fd)
		return os.NewSyscallError("set_non_block", err)
	}

	event.fd = fd
	event.attr = new(Attr)
	*(event.attr) = *attr
	return nil
}

func (event *PerfEvent) Open(attr *Attr, pid, cpu int) error {
	return event.open(attr, pid, cpu, 0)
}

func (event *PerfEvent) MapRingBuf() error {
	ringBufHandler, err := NewRingBufHandler(event.fd, event.attr)
	if err != nil {
		return err
	}

	event.ringBufHandler = ringBufHandler
	return nil
}

func (event *PerfEvent) isValid() bool {
	return event.fd > 0
}

func (event *PerfEvent) ioctl(req int, arg uintptr) error {
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(event.fd), uintptr(req), arg); errno != 0 {
		return errno
	}
	return nil
}

func (event *PerfEvent) Enable() error {
	if !event.isValid() {
		return fmt.Errorf("Failed to enable perf because event isn't valid")
	}

	if err := event.ioctl(unix.PERF_EVENT_IOC_ENABLE, 0); err != nil {
		return err
	}
	return nil
}

func (event *PerfEvent) Disable() error {
	if !event.isValid() {
		return fmt.Errorf("Failed to disable perf event because event isn't valid")
	}

	if err := event.ioctl(unix.PERF_EVENT_IOC_DISABLE, 0); err != nil {
		return err
	}
	return nil
}

func (event *PerfEvent) Reset() error {
	if !event.isValid() {
		return fmt.Errorf("Failed to reset perf event")
	}

	if err := event.ioctl(unix.PERF_EVENT_IOC_RESET, 0); err != nil {
		return err
	}
	return nil
}

func (event *PerfEvent) sendTermToRingBuf() {
	val := uint64(1)
	buf := (*[8]byte)(unsafe.Pointer(&val))[:]
	unix.Write(event.ringBufHandler.termFd, buf)
}

func (event *PerfEvent) HandleReadContent() error {
	var readContent ReadContent
	if !event.isValid() {
		return fmt.Errorf("Failed to handle read content")
	}

	buf := make([]byte, event.attr.ReadFormat.CalcRequiredSize())
	if _, err := unix.Read(event.fd, buf); err != nil {
		return os.NewSyscallError("read", err)
	}

	parser := FieldParser(buf)
	parser.ParseReadContent(event.attr.ReadFormat, &readContent)
	return nil
}

func (event *PerfEvent) Profile(timeout chan bool) error {
	if err := event.Disable(); err != nil {
		return err
	}
	if err := event.Reset(); err != nil {
		return err
	}
	if err := event.Enable(); err != nil {
		return err
	}

	if event.ringBufHandler != nil {
		event.ringBufHandler.HandleRecords()
	}

	<-timeout
	if err := event.Disable(); err != nil {
		logrus.Errorf("Failed to disable perf event [%s]", err.Error())
	}
	if event.ringBufHandler != nil {
		event.sendTermToRingBuf()
	}

	event.HandleReadContent()
	return nil
}
