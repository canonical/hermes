package prometheus

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"hermes/parser"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	PREFIX = "hermes"
)

func HermesPrometheusHandler(reg prometheus.Gatherer) gin.HandlerFunc {
	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// Overarching exporter, with 'sub exporters'
type HermesExporter struct {
	Io *IoLatExporter
}

func NewHermesExporter(viewDir string, dataDir string) *HermesExporter {
	return &HermesExporter{
		Io: NewIoLatExporter(PREFIX+"_io", viewDir, dataDir),
	}
}

func (h *HermesExporter) Collect(ch chan<- prometheus.Metric) {
	h.Io.Collect(ch)
}

func (h *HermesExporter) Describe(ch chan<- *prometheus.Desc) {
	h.Io.Describe(ch)
}

// Return latest data as bytes based on parsed dirs (timestamp named)
// TODO? parsed data could be obtained from the hermes web server instead of filesytsem
func GetLatestParsedDataAsBytes(path string, kind string) ([]byte, error) {
	sourceDir := filepath.Join(path, kind)
	latestDir, err := GetLatestParsedDataDir(sourceDir)
	if err != nil {
		return nil, err
	}
	latestFile := filepath.Join(latestDir, parser.ParsedPostfix[kind])
	return ioutil.ReadFile(latestFile)
}

// Return the most up to date timestamp (dir) containing parsed metrics
func GetLatestParsedDataDir(path string) (ret string, err error) {
	dirItems, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}
	if len(dirItems) > 1 {
		latest := dirItems[0]
		for _, item := range dirItems {
			// check is dir TODO: also validate dirname is timestamp string
			if item.IsDir() {
				if latest.Name() <= item.Name() &&
					latest.Name() != "overview" {
					latest = item
				}
			}
		}
		ret = filepath.Join(path, latest.Name())
	}
	return
}

// Return latest raw data (based on file modtime and postfix name)
func GetLatestRawDataAsBytes(datadir string, postfix string) ([]byte, error) {
	dirItems, err := ioutil.ReadDir(datadir)
	if err != nil {
		return nil, err
	}
	var latest fs.FileInfo
	var fullpath string
	if len(dirItems) > 1 {
		for _, item := range dirItems {
			if item.Mode().IsRegular() {
				if strings.Contains(item.Name(), postfix) {
					if latest == nil {
						latest = item
					} else if item.ModTime().After(latest.ModTime()) {
						latest = item
					}
				}
			}
		}
	}
	if latest == nil {
		return nil, nil
	} else {
		fullpath = filepath.Join(datadir, latest.Name())
		logrus.Debugf("Reading from data file [%s]", fullpath)
	}
	return ioutil.ReadFile(fullpath)
}

// Return byte data from a file
func GetBytesFromFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
