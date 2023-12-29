package ebpf

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"hermes/common"
	"hermes/log"
	"io"
	"os"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -type blk_req_event -target $BPF_ARCH -cc $BPF_CLANG -cflags $BPF_CFLAGS bpf io_latency.c -- -I$BPF_VMLINUX_HEADER

const BlkRecFilePostfix = ".blk.rec"

type IoLatLoader struct {
	objs *bpfObjects
}

func GetLoader() (*IoLatLoader, error) {
	// Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		return nil, err
	}

	return &IoLatLoader{
		objs: &objs,
	}, nil
}

func (loader *IoLatLoader) GetLogDataPathPostfix() string {
	return BlkRecFilePostfix
}

func (loader *IoLatLoader) Prepare(logPathManager log.LogPathManager) error {
	return nil
}

func (loader *IoLatLoader) Load(ctx context.Context) error {
	kpBlkIoStart, err := link.Kprobe("blk_account_io_start", loader.objs.KprobeBlkAccountIoStart, nil)
	if err != nil {
		logrus.Errorf("Failed to open blk_account_io_start kprobe, err [%s]", err)
		return err
	}
	defer kpBlkIoStart.Close()

	kpBlkIoDone, err := link.Kprobe("blk_account_io_done", loader.objs.KprobeBlkAccountIoDone, nil)
	if err != nil {
		logrus.Errorf("Failed to open blk_account_io_start kprobe, err [%s]", err)
		return err
	}
	defer kpBlkIoDone.Close()

	<-ctx.Done()
	return nil
}

// request ops
// copied from vmlinux.h
const (
	REQ_OP_READ  = 0
	REQ_OP_WRITE = 1
	REQ_OP_FLUSH = 2

// REQ_OP_DISCARD = 3,
// REQ_OP_SECURE_ERASE = 5,
// REQ_OP_WRITE_SAME = 7,
// REQ_OP_WRITE_ZEROES = 9,
// REQ_OP_ZONE_OPEN = 10,
// REQ_OP_ZONE_CLOSE = 11,
// REQ_OP_ZONE_FINISH = 12,
// REQ_OP_ZONE_APPEND = 13,
// REQ_OP_ZONE_RESET = 15,
// REQ_OP_ZONE_RESET_ALL = 17,
// REQ_OP_DRV_IN = 34,
// REQ_OP_DRV_OUT = 35,
// REQ_OP_LAST = 36,
// __REQ_FAILFAST_DEV = 8,
// __REQ_FAILFAST_TRANSPORT = 9,
// __REQ_FAILFAST_DRIVER = 10,
)

// request flag bitshifts
// copied from vmlinux.h
const (
	__REQ_SYNC = 11

// __REQ_META = 12,
// __REQ_PRIO = 13,
// __REQ_NOMERGE = 14,
// __REQ_IDLE = 15,
// __REQ_INTEGRITY = 16,
// __REQ_FUA = 17,
// __REQ_PREFLUSH = 18,
// __REQ_RAHEAD = 19,
// __REQ_BACKGROUND = 20,
// __REQ_NOWAIT = 21,
// __REQ_CGROUP_PUNT = 22,
// __REQ_NOUNMAP = 23,
// __REQ_HIPRI = 24,
// __REQ_DRV = 25,
// __REQ_SWAP = 26,
// __REQ_NR_BITS = 27,
)

const (
	REQ_SYNC = 1 << __REQ_SYNC
)

func getOpInfo(b bpfBlkReqEvent) BlkOp {
	var ret BlkOp

	if b.CmdFlags&REQ_OP_WRITE > 0 {
		ret.Op = `write`
	} else if b.CmdFlags&REQ_OP_FLUSH > 0 {
		ret.Op = `flush`
	} else if b.CmdFlags&REQ_OP_READ == 0 {
		ret.Op = `read`
	} else {
		ret.Op = `other`
	}

	if b.CmdFlags&REQ_SYNC > 0 {
		ret.Sync = true
	} else {
		ret.Sync = false
	}

	return ret
}

type BlkOp struct {
	Op   string `json:op`
	Sync bool   `json:sync`
}

type BlkLatRec struct {
	Pid    uint32 `json:pid`
	LatUs  uint64 `json:lat_us`
	Device string `json:device`
	Comm   string `json:comm`
	OpInfo BlkOp
}

// Collect records from ringbuff, assumes Load() has already been called
func (loader *IoLatLoader) getBlkLatRecs() (blkRecs []BlkLatRec) {
	rd, err := ringbuf.NewReader(loader.objs.BlkReqEvents)
	rd.SetDeadline(time.Now().Add(1 * time.Second)) // don't wait longer than 1s for samples
	if err != nil {
		logrus.Errorf("Failed to open ringbuf reader, err [%s]", err)
	}
	defer rd.Close()

	for { // may need to be bounded
		var blkEvent bpfBlkReqEvent
		record, err := rd.Read()
		if err != nil {
			if err == os.ErrDeadlineExceeded || err == ringbuf.ErrClosed {
				break
			}
		}
		if err := binary.Read(bytes.NewBuffer(record.RawSample), common.NativeEndian(), &blkEvent); err != nil {
			if err == io.EOF {
				break
			}
			// there may be some other unchecked error here, we just try reading again
			continue
		} else {
			blkOp := getOpInfo(blkEvent)
			blkRec := BlkLatRec{
				Pid:    blkEvent.Pid,
				LatUs:  blkEvent.DeltaUs,
				Device: unix.ByteSliceToString(blkEvent.DiskName[:]),
				Comm:   unix.ByteSliceToString(blkEvent.Comm[:]),
				OpInfo: blkOp,
			}
			blkRecs = append(blkRecs, blkRec)
		}
	}
	return
}

func (loader *IoLatLoader) writeToFile(outputPath string, bytes *[]byte) error {
	fp, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString(string(*bytes)); err != nil {
		return err
	}
	return nil
}
func (loader *IoLatLoader) writeBlkRec(outputPath string, recs *[]BlkLatRec) error {
	bytes, err := json.Marshal(recs)
	if err != nil {
		return err
	}

	return loader.writeToFile(outputPath, &bytes)
}

func (loader *IoLatLoader) StoreData(logPathManager log.LogPathManager) error {
	blkRecs := loader.getBlkLatRecs()
	if len(blkRecs) > 0 {
		if err := loader.writeBlkRec(logPathManager.DataPath(BlkRecFilePostfix), &blkRecs); err != nil {
			logrus.Errorf("Failed to write blk records to file, err [%s]", err)
			return err
		}
	}

	return nil
}

func (loader *IoLatLoader) Close() {
	loader.objs.Close()
}
