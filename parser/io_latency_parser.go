package parser

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	iolat "hermes/backend/ebpf/io_latency"
	"hermes/log"
)

type IoLatParser struct{}

func GetIoLatEbpfParser() (ParserInstance, error) {
	return &IoLatParser{}, nil
}

type RawBlkLatRec iolat.BlkLatRec

type BlkLatRecord struct {
	TotalIos   int    `json:"total_ios"`
	Reads      int    `json:"reads"`
	SyncReads  int    `json:"sync_reads"`
	Writes     int    `json:"writes"`
	SyncWrites int    `json:"sync_writes"`
	Other      int    `json:"other"`
	SyncOther  int    `json:"sync_other"`
	LatAvgUs   uint64 `json:"lat_avg_us"`
	LatHighUs  uint64 `json:"lat_high_us"`
	LatLowUs   uint64 `json:"lat_low_us"`
	LatSum     uint64 `json:"-"` //only used for calculating average
}

type PidBlkLatRecord struct {
	Comm   string       `json:"comm"`
	BlkLat BlkLatRecord `json:"blk_lat"`
}

// update BlkLatRecord based on a new raw record
func (b *BlkLatRecord) update(next RawBlkLatRec) {
	// update high
	if next.LatUs > b.LatHighUs || b.TotalIos == 0 {
		b.LatHighUs = next.LatUs
	}

	// update low
	if next.LatUs < b.LatLowUs || b.TotalIos == 0 {
		b.LatLowUs = next.LatUs
	}

	// update counts
	if next.OpInfo.Op == "read" {
		if next.OpInfo.Sync == true {
			b.SyncReads += 1
		} else {
			b.Reads += 1
		}
	}

	if next.OpInfo.Op == "write" {
		if next.OpInfo.Sync == true {
			b.SyncWrites += 1
		} else {
			b.Writes += 1
		}
	}

	if next.OpInfo.Op == "other" {
		if next.OpInfo.Sync == true {
			b.SyncOther += 1
		} else {
			b.Other += 1
		}
	}
	b.TotalIos += 1
	b.LatSum += next.LatUs // only used for calculating avg
}

func (b *BlkLatRecord) calculateAverage() {
	if b.TotalIos != 0 {
		b.LatAvgUs = b.LatSum / uint64(b.TotalIos)
	}
	return
}

func (p *IoLatParser) getParsedBlkData(rawRecs []RawBlkLatRec) OutputBlkData {
	all := BlkLatRecord{}
	perPid := map[uint32]PidBlkLatRecord{}
	perDev := map[string]BlkLatRecord{}
	perComm := map[string]BlkLatRecord{}

	for _, rec := range rawRecs {
		pid := rec.Pid
		pidRec := perPid[pid]
		pidBlkRec := pidRec.BlkLat

		dev := rec.Device
		devBlkRec := perDev[dev]

		comm := rec.Comm
		commBlkRec := perComm[comm]

		// set comm for per pid
		if pidBlkRec.TotalIos == 0 {
			pidRec.Comm = rec.Comm
		}

		//update values
		all.update(rec)
		pidBlkRec.update(rec)
		devBlkRec.update(rec)
		commBlkRec.update(rec)

		// update records
		pidRec.BlkLat = pidBlkRec
		perPid[pid] = pidRec
		perDev[dev] = devBlkRec
		perComm[comm] = commBlkRec

	}

	// calculate averages
	all.calculateAverage()

	for i, pidRec := range perPid {
		pidBlkRec := pidRec.BlkLat
		pidBlkRec.calculateAverage()
		pidRec.BlkLat = pidBlkRec
		perPid[i] = pidRec
	}
	for i, devBlkRec := range perDev {
		devBlkRec.calculateAverage()
		perDev[i] = devBlkRec
	}
	for i, commBlkRec := range perComm {
		commBlkRec.calculateAverage()
		perComm[i] = commBlkRec
	}

	return OutputBlkData{all, perDev, perComm, perPid}
}

type OutputBlkData struct {
	AllIo   BlkLatRecord               `json:"all"`
	PerDev  map[string]BlkLatRecord    `json:"per_dev"`
	PerComm map[string]BlkLatRecord    `json:"per_comm"`
	PerPid  map[uint32]PidBlkLatRecord `json:"per_pid"`
}

// get raw records created by collector
func (p *IoLatParser) getRawBlkRecord(timestamp int64, path string) ([]RawBlkLatRec, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rawRecords []RawBlkLatRec
	if err := json.Unmarshal(bytes, &rawRecords); err != nil {
		return nil, err
	}
	return rawRecords, nil
}

func (p *IoLatParser) writeJSONDataBlk(rec OutputBlkData, path string) error {
	bytes, err := json.Marshal(&rec)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bytes, 0644)
}

func (p *IoLatParser) Parse(logDataPathGenerator log.LogDataPathGenerator, timestamp int64, logDataPostfix, outputDir string) error {
	var rawRecs []RawBlkLatRec

	rawRecs, err := p.getRawBlkRecord(timestamp, logDataPathGenerator(logDataPostfix))
	if err != nil {
		return err
	}
	outputBlkData := p.getParsedBlkData(rawRecs)

	outputBlkPath := filepath.Join(outputDir, strconv.FormatInt(timestamp, 10), "blk_ios.json")
	if err := os.MkdirAll(filepath.Dir(outputBlkPath), os.ModePerm); err != nil {
		return err
	}
	err = p.writeJSONDataBlk(outputBlkData, outputBlkPath)
	if err != nil {
		return err
	}

	return nil
}
