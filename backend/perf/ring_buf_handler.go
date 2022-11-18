package perf

import (
	"os"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const DefaultPageNum = 128

type PollResp struct {
	Ready bool
	Term  bool
	Err   error
}

type RingBuf struct {
	Ring     []byte
	RingData []byte
	MetaPage *unix.PerfEventMmapPage
}

type RingBufHandler struct {
	ringBuf RingBuf
	perfFd  int
	attr    *Attr
	termFd  int
	timeout chan time.Duration
}

func NewRingBufHandler(perfFd int, attr *Attr) (*RingBufHandler, error) {
	pageSize := unix.Getpagesize()
	size := (DefaultPageNum + 1) * pageSize
	ring, err := unix.Mmap(perfFd, 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return nil, os.NewSyscallError("mmap", err)
	}

	metaPage := (*unix.PerfEventMmapPage)(unsafe.Pointer(&ring[0]))
	if metaPage.Data_offset == 0 && metaPage.Data_size == 0 {
		atomic.StoreUint64(&metaPage.Data_offset, uint64(pageSize))
		atomic.StoreUint64(&metaPage.Data_size, uint64(pageSize*DefaultPageNum))
	}

	termFd, err := unix.Eventfd(0, unix.EFD_CLOEXEC|unix.EFD_NONBLOCK)
	if err != nil {
		return nil, os.NewSyscallError("eventfd", err)
	}

	return &RingBufHandler{
		ringBuf: RingBuf{
			Ring:     ring,
			RingData: ring[metaPage.Data_offset:],
			MetaPage: metaPage,
		},
		perfFd:  perfFd,
		attr:    attr,
		termFd:  termFd,
		timeout: make(chan time.Duration),
	}, nil
}

func (handler *RingBufHandler) poll() PollResp {
	pollFds := []unix.PollFd{
		{Fd: int32(handler.perfFd), Events: unix.POLLIN},
		{Fd: int32(handler.termFd), Events: unix.POLLIN},
	}

	for {
		_, err := unix.Poll(pollFds, -1)
		if err == unix.EINTR {
			continue
		}
		if (pollFds[1].Revents & unix.POLLIN) != 0 {
			var buf [8]byte
			unix.Read(handler.termFd, buf[:])
		}
		return PollResp{
			Ready: (pollFds[0].Revents & unix.POLLIN) != 0,
			Term:  (pollFds[1].Revents & unix.POLLIN) != 0,
			Err:   err,
		}
	}
}

func (handler *RingBufHandler) parseRecords() error {
	parser := NewRecordParser(&handler.ringBuf, handler.attr)

	for {
		rec, err := parser.GetRecord()
		if err != nil {
			logrus.Errorf("Failed to get record [%s]", err.Error())
			return err
		}
		if rec == nil {
			break
		}
		rec.Process()
	}
	return nil
}

func (handler *RingBufHandler) handleRecords() {
	for {
		pollResp := handler.poll()
		if pollResp.Term {
			break
		} else if pollResp.Err != nil {
			logrus.Errorf("Failed to poll ring buffer [%s]", pollResp.Err.Error())
		}

		if err := handler.parseRecords(); err != nil {
			logrus.Errorf("Failed to get records from ring buffer [%s]", err.Error())
		}
	}

	if err := handler.parseRecords(); err != nil {
		logrus.Errorf("Failed to get records from ring buffer [%s]", err.Error())
	}
}

func (handler *RingBufHandler) HandleRecords() {
	go handler.handleRecords()
}
