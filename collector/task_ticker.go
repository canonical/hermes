package collector

import (
	"context"
	"time"
)

type TaskTicker struct {
	ReadyTasks chan string
	quit       chan bool
	cancel     map[string]chan bool
}

func NewTaskTicker() (*TaskTicker, error) {
	return &TaskTicker{
		ReadyTasks: make(chan string, 16),
		quit:       make(chan bool),
		cancel:     make(map[string]chan bool)}, nil
}

func (taskTicker *TaskTicker) AddTask(task Task) {
	taskTicker.cancel[task.Name] = make(chan bool)

	switch task.Class {
	case Disposable:
		go func() {
			timer := time.NewTimer(time.Microsecond)
			defer timer.Stop()
			select {
			case <-taskTicker.quit:
			case <-taskTicker.cancel[task.Name]:
				return
			case <-timer.C:
				taskTicker.ReadyTasks <- task.Name
			}
		}()
	case Periodic:
		go func() {
			ticker := time.NewTicker(time.Duration(task.Interval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-taskTicker.quit:
				case <-taskTicker.cancel[task.Name]:
					return
				case <-ticker.C:
					taskTicker.ReadyTasks <- task.Name
				}
			}
		}()
	}
}

func (taskTicker *TaskTicker) RemoveTask(taskName string) {
	if _, isExist := taskTicker.cancel[taskName]; isExist {
		taskTicker.cancel[taskName] <- true
	}
}

func (taskTicker *TaskTicker) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			taskTicker.quit <- true
		}
	}
}

func (taskTicker *TaskTicker) Run(ctx context.Context) {
	go taskTicker.run(ctx)
}
