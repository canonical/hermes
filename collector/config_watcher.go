package collector

import (
	"context"
	"io/ioutil"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

type ConfigWatcher struct {
	jobQueueComm JobQueueComm
	watcher      *fsnotify.Watcher
}

func NewConfigWatcher(jobQueueComm JobQueueComm) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &ConfigWatcher{
		jobQueueComm: jobQueueComm,
		watcher:      watcher}, nil
}

func (watcher *ConfigWatcher) initConfigs(configDir string) error {
	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}

		if err := watcher.handleConfig(fsnotify.Create, filepath.Join(configDir, file.Name())); err != nil {
			return err
		}
	}

	return nil
}

func (watcher *ConfigWatcher) handleConfig(op fsnotify.Op, configPath string) error {
	job, err := NewJob(configPath)
	if err != nil {
		return err
	}

	switch op {
	case fsnotify.Create:
		watcher.jobQueueComm.AddJob <- *job
	case fsnotify.Remove:
		watcher.jobQueueComm.RemoveJob <- job.Name
	}
	return <-watcher.jobQueueComm.Ack
}

func (watcher *ConfigWatcher) Run(ctx context.Context, configDir string) error {
	go func() {
		if err := watcher.initConfigs(configDir); err != nil {
			logrus.Error(err)
		}

		monitorOps := map[fsnotify.Op]bool{
			fsnotify.Create: true,
			fsnotify.Remove: true,
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.watcher.Events:
				if !ok {
					continue
				}
				if filepath.Ext(event.Name) != ".yaml" {
					continue
				}
				if _, isExist := monitorOps[event.Op]; !isExist {
					continue
				}
				err := watcher.handleConfig(event.Op, event.Name)
				if err != nil {
					logrus.Errorf("Failed to handle config path [%s], err [%s].", event.Name, err)
				}
			case err, ok := <-watcher.watcher.Errors:
				if !ok {
					continue
				}
				logrus.Errorf("Failed to watch events, err [%s].", err)
			}
		}
	}()

	err := watcher.watcher.Add(configDir)
	if err != nil {
		return err
	}

	return nil
}

func (watcher *ConfigWatcher) Release() {
	watcher.watcher.Close()
}
