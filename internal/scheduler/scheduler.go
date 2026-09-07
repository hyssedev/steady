package scheduler

import (
	"container/heap"
	"context"
	"fmt"
	"time"

	"github.com/hyssedev/steady/internal/monitor"
)

type Scheduler struct {
	ctx        context.Context
	interval   time.Duration
	monitors   []monitor.Monitor
	workerChan chan monitor.Monitor
}

func NewScheduler(
	ctx context.Context,
	interval time.Duration,
	monitors []monitor.Monitor,
	workerChan chan monitor.Monitor,
) Scheduler {
	return Scheduler{
		ctx:        ctx,
		interval:   interval,
		monitors:   monitors,
		workerChan: workerChan,
	}
}

type scheduledMonitor struct {
	monitor   monitor.Monitor
	nextCheck time.Time
}

func (s Scheduler) Run() {
	mh := make(MonitorHeap, len(s.monitors))

	for i := range s.monitors {
		mh[i] = &scheduledMonitor{
			monitor:   s.monitors[i],
			nextCheck: time.Now().Add(s.interval),
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
			select {
			case s.workerChan <- scheduled.monitor:
				scheduled.nextCheck = time.Now().Add(s.interval)
				heap.Push(&mh, scheduled)

			case <-s.ctx.Done():
				return
			}
		case <-s.ctx.Done():
			timer.Stop()
			return
		}
	}
}
