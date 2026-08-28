package scheduler

import (
	"container/heap"
	"context"
	"fmt"
	"time"

	"github.com/hyssedev/steady/internal/config"
	"github.com/hyssedev/steady/internal/monitor"
)

type Scheduler struct {
	ctx        context.Context
	cfg        config.Config
	workerChan chan *monitor.Monitor
}

func NewScheduler(ctx context.Context, cfg config.Config, workerChan chan *monitor.Monitor) Scheduler {
	return Scheduler{
		ctx:        ctx,
		cfg:        cfg,
		workerChan: workerChan,
	}
}

type scheduledMonitor struct {
	monitor   *monitor.Monitor
	nextCheck time.Time
	lastCheck time.Time
	index     int
}

func (s Scheduler) Run() {
	mh := make(MonitorHeap, len(s.cfg.Monitors))

	for i, m := range s.cfg.Monitors {
		mh[i] = &scheduledMonitor{
			monitor:   &m,
			nextCheck: time.Now().Add(s.cfg.Interval),
			index:     i,
		}
	}

	heap.Init(&mh)

	for {
		scheduled := heap.Pop(&mh).(*scheduledMonitor)

		wait := time.Until(scheduled.nextCheck)

		fmt.Printf("wait %v\n", wait)

		timer := time.NewTimer(wait)

		select {
		case <-timer.C:
			s.workerChan <- scheduled.monitor

			scheduled.nextCheck = time.Now().Add(s.cfg.Interval)
			scheduled.lastCheck = time.Now()

			heap.Push(&mh, scheduled)
		case <-s.ctx.Done():
			// TODO: clean-up
			return
		}
	}
}
