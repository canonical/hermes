package collector

import (
	"context"
	"time"
)

type JobTicker struct {
	ReadyJobs chan string
	quit      chan bool
	cancel    map[string]chan bool
}

func NewJobTicker() (*JobTicker, error) {
	return &JobTicker{
		ReadyJobs: make(chan string, 16),
		quit:      make(chan bool),
		cancel:    make(map[string]chan bool)}, nil
}

func (jobTicker *JobTicker) AddJob(job Job) {
	jobTicker.cancel[job.Name] = make(chan bool)

	switch job.Class {
	case Disposable:
		go func() {
			timer := time.NewTimer(time.Microsecond)
			defer timer.Stop()
			select {
			case <-jobTicker.quit:
			case <-jobTicker.cancel[job.Name]:
				return
			case <-timer.C:
				jobTicker.ReadyJobs <- job.Name
			}
		}()
	case Periodic:
		go func() {
			ticker := time.NewTicker(time.Duration(job.Interval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-jobTicker.quit:
				case <-jobTicker.cancel[job.Name]:
					return
				case <-ticker.C:
					jobTicker.ReadyJobs <- job.Name
				}
			}
		}()
	}
}

func (jobTicker *JobTicker) RemoveJob(jobName string) {
	if _, isExist := jobTicker.cancel[jobName]; isExist {
		jobTicker.cancel[jobName] <- true
	}
}

func (jobTicker *JobTicker) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			jobTicker.quit <- true
		}
	}
}

func (jobTicker *JobTicker) Run(ctx context.Context) {
	go jobTicker.run(ctx)
}
