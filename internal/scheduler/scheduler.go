package scheduler

import (
	"container/heap"
	"context"
	"fmt"
	"time"

	"github.com/hyssedev/steady/internal/config"
)

type Scheduler struct {
	ctx context.Context
	cfg config.Config
}

func NewScheduler(ctx context.Context, cfg config.Config) Scheduler {
	return Scheduler{
		ctx: ctx,
		cfg: cfg,
	}
}

type scheduledMonitor struct {
	*config.Monitor
	nextCheck time.Time
	lastCheck time.Time
	index     int
}

func (s Scheduler) Run() {
	mh := make(MonitorHeap, len(s.cfg.Monitors))

	for i, m := range s.cfg.Monitors {
		mh[i] = &scheduledMonitor{
			Monitor:   &m,
			nextCheck: time.Now().Add(s.cfg.Interval),
			index:     i,
		}
	}

	heap.Init(&mh)

	for {
		monitor := heap.Pop(&mh).(*scheduledMonitor)

		wait := time.Until(monitor.nextCheck)
		fmt.Printf("wait %v\n", wait)

		timer := time.NewTimer(wait)

		select {
		case <-timer.C:
			monitor.nextCheck = time.Now().Add(s.cfg.Interval)
			monitor.lastCheck = time.Now()

			heap.Push(&mh, monitor)
		case <-s.ctx.Done():
			// TODO: clean-up
			return
		}
	}
}
