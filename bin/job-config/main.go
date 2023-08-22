package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"hermes/collector"
	"hermes/common"

	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	metadataDir string
	configDir   string
)

func init() {
	flag.StringVar(&configDir, "config_dir", metadataDir+common.ConfigDirDefault, "The path of config directory")
	flag.Usage = Usage
}

func Usage() {
	fmt.Println("Usage: job-config [config_dir]")
	flag.PrintDefaults()
}

func TrimFileNameExt(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func getJobStatus(filePath string) (string, error) {
	job, err := collector.NewJob(filePath)
	if err != nil {
		return "", err
	}
	return job.Status, nil
}

func updateJobStatus(filePath, status string) error {
	job, err := collector.NewJob(filePath)
	if err != nil {
		return err
	}
	job.Status = status
	bytes, err := yaml.Marshal(&job)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, bytes, 0644)
}

func main() {
	app := tview.NewApplication()
	form := tview.NewForm()

	flag.Parse()

	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, file := range files {
		if file.Name() == collector.TasksDir {
			continue
		}

		filePath := filepath.Join(configDir, file.Name())
		status, err := getJobStatus(filePath)
		if err != nil {
			logrus.Errorf("Failed to get job status, config [%s], err [%s]", filePath, err)
			continue
		}

		curOption := 0
		if status == collector.Disabled {
			curOption = 1
		}

		form.AddDropDown(TrimFileNameExt(file.Name()), []string{"Enabled", "Disabled"}, curOption,
			func(opt string, optIdx int) {
				status := collector.Enabled
				if opt == "Disabled" {
					status = collector.Disabled
				}
				if err := updateJobStatus(filePath, status); err != nil {
					logrus.Errorf("Failed to update job status, path [%s], status [%s], err [%s]",
						filePath, status, err)
				}
			})
	}

	if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		logrus.Fatal(err)
	}
}
