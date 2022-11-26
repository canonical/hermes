package perf

import (
	"context"
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
	ID             uint64
	Fd             int
	attr           *Attr
	ringBufHandler *RingBufHandler
	groups         []*PerfEvent
	idGroups       map[uint64]*PerfEvent
}

func NewPerfEvent(attr *Attr, pid, cpu int) (*PerfEvent, error) {
	var event PerfEvent

	if err := event.Open(attr, pid, cpu, nil); err != nil {
		return nil, err
	}
	if event.attr.Sample != 0 {
		if err := event.MapRingBuf(); err != nil {
			event.Release()
			return nil, err
		}
	}

	return &event, nil
}

func NewPerfGroupEvent(group *Group, pid, cpu int) (*PerfEvent, error) {
	var leader PerfEvent
	leaderAttr := group.GetLeaderAttr()
	if leaderAttr == nil {
		return nil, fmt.Errorf("Failed to get leader attr")
	}
	if err := leader.Open(leaderAttr, pid, cpu, nil); err != nil {
		return nil, err
	}
	if group.needRingBuf {
		if err := leader.MapRingBuf(); err != nil {
			leader.Release()
			return nil, err
		}
	}

	for _, attr := range group.GetFollowerAttrs() {
		var follower PerfEvent
		if err := follower.Open(attr, pid, cpu, &leader); err != nil {
			leader.Release()
			return nil, err
		}
		if attr.Sample != 0 {
			if err := follower.redirectReadRecord(&leader); err != nil {
				leader.Release()
				return nil, err
			}
		}
	}

	return &leader, nil
}

func (event *PerfEvent) Open(attr *Attr, pid, cpu int, groupEvent *PerfEvent) error {
	if event.IsValid() {
		return nil
	}

	groupPerfFd := -1
	if groupEvent != nil {
		if !groupEvent.IsValid() {
			return fmt.Errorf("Group event is invalid")
		}
		groupPerfFd = groupEvent.Fd
	}
	flags := unix.PERF_FLAG_FD_CLOEXEC
	fd, err := unix.PerfEventOpen(attr.ToUnixPerfEventAttr(), pid, cpu, groupPerfFd, flags)
	if err != nil {
		return os.NewSyscallError("perf_event_open", err)
	}
	event.Fd = fd

	if err := unix.SetNonblock(fd, true); err != nil {
		unix.Close(fd)
		return os.NewSyscallError("set_non_block", err)
	}

	id, err := event.GetID()
	if err != nil {
		unix.Close(fd)
		return err
	}
	event.ID = id

	event.attr = new(Attr)
	*(event.attr) = *attr
	if groupEvent != nil {
		if groupEvent.idGroups == nil {
			groupEvent.idGroups = map[uint64]*PerfEvent{}
		}
		groupEvent.groups = append(groupEvent.groups, event)
		groupEvent.idGroups[id] = event
	}
	return nil
}

func (event *PerfEvent) MapRingBuf() error {
	ringBufHandler, err := NewRingBufHandler(event.Fd, event.attr)
	if err != nil {
		return err
	}
	event.ringBufHandler = ringBufHandler
	return nil
}

func (event *PerfEvent) IsValid() bool {
	return event.Fd > 0
}

func (event *PerfEvent) ioctlPointer(req int, arg unsafe.Pointer) error {
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(event.Fd), uintptr(req), uintptr(arg)); errno != 0 {
		return errno
	}
	return nil
}

func (event *PerfEvent) ioctl(req int, arg uintptr) error {
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(event.Fd), uintptr(req), arg); errno != 0 {
		return errno
	}
	return nil
}

func (event *PerfEvent) GetID() (uint64, error) {
	if !event.IsValid() {
		return 0, fmt.Errorf("Event is invalid")
	}
	var val uint64
	err := event.ioctlPointer(unix.PERF_EVENT_IOC_ID, unsafe.Pointer(&val))
	return val, err
}

func (event *PerfEvent) redirectReadRecord(targetEvent *PerfEvent) error {
	if !event.IsValid() {
		return fmt.Errorf("Failed to redirect read record because follower is invalid")
	}
	var targetFd int
	if targetEvent == nil {
		targetFd = -1
	} else {
		if !targetEvent.IsValid() {
			return fmt.Errorf("Failed to redirect read record because target is invalid")
		}
		targetFd = targetEvent.Fd
	}
	return event.ioctl(unix.PERF_EVENT_IOC_SET_OUTPUT, uintptr(targetFd))
}

func (event *PerfEvent) Enable() error {
	if !event.IsValid() {
		return fmt.Errorf("Failed to enable perf because event isn't valid")
	}
	return event.ioctl(unix.PERF_EVENT_IOC_ENABLE, 0)
}

func (event *PerfEvent) Disable() error {
	if !event.IsValid() {
		return fmt.Errorf("Failed to disable perf event because event isn't valid")
	}
	return event.ioctl(unix.PERF_EVENT_IOC_DISABLE, 0)
}

func (event *PerfEvent) Reset() error {
	if !event.IsValid() {
		return fmt.Errorf("Failed to reset perf event")
	}
	return event.ioctl(unix.PERF_EVENT_IOC_RESET, 0)
}

func (event *PerfEvent) sendTermToRingBuf() {
	val := uint64(1)
	buf := (*[8]byte)(unsafe.Pointer(&val))[:]
	unix.Write(event.ringBufHandler.termFd, buf)
}

func (event *PerfEvent) handleSingleReadContent() error {
	var readContent ReadContent
	if !event.IsValid() {
		return fmt.Errorf("Failed to handle read content")
	}

	buf := make([]byte, event.attr.ReadFormat.CalcRequiredSize())
	if _, err := unix.Read(event.Fd, buf); err != nil {
		return os.NewSyscallError("read", err)
	}

	parser := FieldParser(buf)
	parser.ParseReadContent(event.attr.ReadFormat, &readContent)
	return nil
}

func (event *PerfEvent) handleGroupReadContent() error {
	var groupReadContent GroupReadContent
	size := event.attr.ReadFormat.CalcGroupRequiredSize(1 + len(event.groups))
	buf := make([]byte, size)
	if _, err := unix.Read(event.Fd, buf); err != nil {
		return os.NewSyscallError("read", err)
	}

	parser := FieldParser(buf)
	parser.ParseGroupReadContent(event.attr.ReadFormat, &groupReadContent)
	return nil
}

func (event *PerfEvent) handleReadContent() error {
	if !event.IsValid() {
		return fmt.Errorf("Failed to handle read content")
	}
	if len(event.groups) == 0 {
		return event.handleSingleReadContent()
	}
	return event.handleGroupReadContent()
}

func (event *PerfEvent) Profile(ctx context.Context, outputPath string) error {
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
		event.ringBufHandler.HandleRecords(outputPath)
	}

	<-ctx.Done()
	if err := event.Disable(); err != nil {
		logrus.Errorf("Failed to disable perf event [%s]", err.Error())
	}
	if event.ringBufHandler != nil {
		event.sendTermToRingBuf()
	}

	event.handleReadContent()
	return nil
}

func (event *PerfEvent) Release() {
	if event.ringBufHandler != nil {
		event.ringBufHandler.Release()
	}
	for _, _event := range event.groups {
		_event.Release()
	}
	unix.Close(event.Fd)
}
