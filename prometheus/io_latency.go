package prometheus

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	iolat "hermes/backend/ebpf/io_latency"
	"hermes/parser"
	"sync"
)

type IoLatExporter struct {
	viewDir string
	dataDir string

	ioLat *prometheus.HistogramVec

	// per comm metrics
	totalIos   *prometheus.GaugeVec
	reads      *prometheus.GaugeVec
	syncReads  *prometheus.GaugeVec
	writes     *prometheus.GaugeVec
	syncWrites *prometheus.GaugeVec
	other      *prometheus.GaugeVec
	syncOther  *prometheus.GaugeVec
	latAvgUs   *prometheus.GaugeVec
	latHighUs  *prometheus.GaugeVec
	latLowUs   *prometheus.GaugeVec

	// per dev metrics
	deviceTotalIos   *prometheus.GaugeVec
	deviceReads      *prometheus.GaugeVec
	deviceSyncReads  *prometheus.GaugeVec
	deviceWrites     *prometheus.GaugeVec
	deviceSyncWrites *prometheus.GaugeVec
	deviceOther      *prometheus.GaugeVec
	deviceSyncOther  *prometheus.GaugeVec
	deviceLatAvgUs   *prometheus.GaugeVec
	deviceLatHighUs  *prometheus.GaugeVec
	deviceLatLowUs   *prometheus.GaugeVec
}

// sourceDir: the 'viewDir' or dir containing parsed outputs
func NewIoLatExporter(namespace string, viewDir string, dataDir string) *IoLatExporter {
	return &IoLatExporter{
		viewDir: viewDir,
		dataDir: dataDir,
		ioLat: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "latency_us",
			Help:      "the latency of a block device operation in microseconds (us)",
			Buckets:   []float64{1, 500, 1000, 5000, 10000, 50000, 100000, 200000, 500000},
		},
			[]string{"device"},
		),
		totalIos: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ios_total",
			Help:      "the total number of io operations",
		},
			[]string{"comm"},
		),
		reads: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "reads",
			Help:      "the total number of (async) reads",
		},
			[]string{"comm"},
		),
		syncReads: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sync_reads",
			Help:      "the total number of sync reads",
		},
			[]string{"comm"},
		),
		writes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "writes",
			Help:      "the total number of (async) writes",
		},
			[]string{"comm"},
		),
		syncWrites: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sync_writes",
			Help:      "the total number of sync writes",
		},
			[]string{"comm"},
		),
		other: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "other",
			Help:      "the total number of other ios",
		},
			[]string{"comm"},
		),
		syncOther: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sync_other",
			Help:      "the total number of sync other ios",
		},
			[]string{"comm"},
		),
		latAvgUs: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "lat_avg_us",
			Help:      "average io latency",
		},
			[]string{"comm"},
		),
		latHighUs: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "lat_high_us",
			Help:      "highest io latency for this timestamp",
		},
			[]string{"comm"},
		),
		latLowUs: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "lat_low_us",
			Help:      "lowest io latency for this timestamp",
		},
			[]string{"comm"},
		),
		deviceTotalIos: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "device_ios_total",
			Help:      "the total number of io operations on a device",
		},
			[]string{"device"},
		),
		deviceReads: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "device_reads",
			Help:      "the total number of (async) reads on a device",
		},
			[]string{"device"},
		),
		deviceSyncReads: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "device_sync_reads",
			Help:      "the total reads on a device",
		},
			[]string{"device"},
		),
		deviceWrites: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "device_writes",
			Help:      "the total writes on a device",
		},
			[]string{"device"},
		),
		deviceSyncWrites: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "device_sync_writes",
			Help:      "the total sync writes on a device",
		},
			[]string{"device"},
		),
		deviceOther: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "device_other",
			Help:      "the total (async) other ios on a device",
		},
			[]string{"device"},
		),
		deviceSyncOther: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "device_sync_other",
			Help:      "the total other sync ios on a device",
		},
			[]string{"device"},
		),
		deviceLatAvgUs: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "device_lat_avg_us",
			Help:      "the average latency from this timestamp",
		},
			[]string{"device"},
		),
		deviceLatHighUs: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "device_lat_high_us",
			Help:      "the highest latency op in this timestamp",
		},
			[]string{"device"},
		),
		deviceLatLowUs: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "device_lat_low_us",
			Help:      "the lowest latency op in this timestamp",
		},
			[]string{"device"},
		),
	}
}

func (i *IoLatExporter) Describe(ch chan<- *prometheus.Desc) {
	i.ioLat.Describe(ch)

	i.totalIos.Describe(ch)
	i.reads.Describe(ch)
	i.syncReads.Describe(ch)
	i.writes.Describe(ch)
	i.syncWrites.Describe(ch)
	i.other.Describe(ch)
	i.syncOther.Describe(ch)
	i.latAvgUs.Describe(ch)
	i.latHighUs.Describe(ch)
	i.latLowUs.Describe(ch)

	i.deviceTotalIos.Describe(ch)
	i.deviceReads.Describe(ch)
	i.deviceSyncReads.Describe(ch)
	i.deviceWrites.Describe(ch)
	i.deviceSyncWrites.Describe(ch)
	i.deviceOther.Describe(ch)
	i.deviceSyncOther.Describe(ch)
	i.deviceLatAvgUs.Describe(ch)
	i.deviceLatHighUs.Describe(ch)
	i.deviceLatLowUs.Describe(ch)
}

func (i *IoLatExporter) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	wg.Add(2)
	go i.collectRaw(ch, &wg)
	go i.collectParsed(ch, &wg)
	wg.Wait()
}

// collect data from pre-parsed hermes raw data files
func (i *IoLatExporter) collectRaw(ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	bytes, err := GetLatestRawDataAsBytes(i.dataDir, iolat.BlkRecFilePostfix)

	if err != nil {
		logrus.Errorf("collectRaw: error getting latest data from dir [%s] postfix [%s]: [%s]", i.viewDir, iolat.BlkRecFilePostfix, err)
	}
	var data []iolat.BlkLatRec
	if err := json.Unmarshal(bytes, &data); err != nil {
		logrus.Errorf("collectRaw: error unmarshalling data [%s]", err)
	}
	//process per comm data
	for _, rec := range data {
		labels := prometheus.Labels{"device": rec.Device}
		i.ioLat.With(labels).Observe(float64(rec.LatUs))
	}

	// send per comm
	i.ioLat.Collect(ch)
	wg.Done()
}

// collect data from post-parsed hermes data files
func (i *IoLatExporter) collectParsed(ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	bytes, err := GetLatestParsedDataAsBytes(i.viewDir, parser.IoLatencyJob)
	if err != nil {
		logrus.Errorf("collectParsed: error getting latest data [%s]", err)
	}
	var data parser.OutputBlkData
	if err := json.Unmarshal(bytes, &data); err != nil {
		logrus.Errorf("collectParsed: error unmarshalling data [%s]", err)
	}
	//process per comm data
	for comm, rec := range data.PerComm {
		labels := prometheus.Labels{"comm": comm}
		i.totalIos.With(labels).Set(float64(rec.TotalIos))
		i.reads.With(labels).Set(float64(rec.Reads))
		i.syncReads.With(labels).Set(float64(rec.SyncReads))
		i.writes.With(labels).Set(float64(rec.Writes))
		i.syncWrites.With(labels).Set(float64(rec.SyncWrites))
		i.other.With(labels).Set(float64(rec.Other))
		i.syncOther.With(labels).Set(float64(rec.SyncOther))
		i.latAvgUs.With(labels).Set(float64(rec.LatAvgUs))
		i.latHighUs.With(labels).Set(float64(rec.LatHighUs))
		i.latLowUs.With(labels).Set(float64(rec.LatLowUs))
	}
	//process per dev data
	for dev, rec := range data.PerDev {
		labels := prometheus.Labels{"device": dev}
		i.deviceTotalIos.With(labels).Set(float64(rec.TotalIos))
		i.deviceReads.With(labels).Set(float64(rec.Reads))
		i.deviceSyncReads.With(labels).Set(float64(rec.SyncReads))
		i.deviceWrites.With(labels).Set(float64(rec.Writes))
		i.deviceSyncWrites.With(labels).Set(float64(rec.SyncWrites))
		i.deviceOther.With(labels).Set(float64(rec.Other))
		i.deviceSyncOther.With(labels).Set(float64(rec.SyncOther))
		i.deviceLatAvgUs.With(labels).Set(float64(rec.LatAvgUs))
		i.deviceLatHighUs.With(labels).Set(float64(rec.LatHighUs))
		i.deviceLatLowUs.With(labels).Set(float64(rec.LatLowUs))
	}

	// send per comm
	i.totalIos.Collect(ch)
	i.reads.Collect(ch)
	i.syncReads.Collect(ch)
	i.writes.Collect(ch)
	i.syncWrites.Collect(ch)
	i.other.Collect(ch)
	i.syncOther.Collect(ch)
	i.latAvgUs.Collect(ch)
	i.latHighUs.Collect(ch)
	i.latLowUs.Collect(ch)

	i.deviceTotalIos.Collect(ch)
	i.deviceReads.Collect(ch)
	i.deviceSyncReads.Collect(ch)
	i.deviceWrites.Collect(ch)
	i.deviceSyncWrites.Collect(ch)
	i.deviceOther.Collect(ch)
	i.deviceSyncOther.Collect(ch)
	i.deviceLatAvgUs.Collect(ch)
	i.deviceLatHighUs.Collect(ch)
	i.deviceLatLowUs.Collect(ch)
	wg.Done()
}
