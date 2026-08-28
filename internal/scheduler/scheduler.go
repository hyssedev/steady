package scheduler

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/hyssedev/steady/internal/config"
)

type Scheduler struct{}

type Monitor struct {
	config.Monitor
	nextCheck time.Time
	index     int
}

type MonitorHeap []*Monitor

func (mh MonitorHeap) Len() int {
	return len(mh)
}

func (mh MonitorHeap) Less(i, j int) bool {
	return mh[i].nextCheck.Before(mh[j].nextCheck)
}

func (mh MonitorHeap) Swap(i, j int) {
	mh[i], mh[j] = mh[j], mh[i]
	mh[i].index = i
	mh[j].index = j
}

func (mh *MonitorHeap) Push(x any) {
	n := len(*mh)
	monitor := x.(*Monitor)
	monitor.index = n
	*mh = append(*mh, monitor)
}

func (mh *MonitorHeap) Pop() any {
	old := *mh
	n := len(old)
	monitor := old[n-1]
	monitor.index = -1 // for safety
	*mh = old[0 : n-1]
	return monitor
}

func (mh *MonitorHeap) update(monitor *Monitor, nextCheck time.Time) {
	monitor.nextCheck = nextCheck
	heap.Fix(mh, monitor.index)
}

func Test(cfg config.Config) {
	mh := make(MonitorHeap, len(cfg.Monitors))

	for i, m := range cfg.Monitors {
		mh[i] = &Monitor{
			Monitor:   m,
			nextCheck: time.Now().Add(cfg.Interval),
			index:     i,
		}
	}

	heap.Init(&mh)

	for {
		monitor := heap.Pop(&mh).(*Monitor)

		wait := time.Until(monitor.nextCheck)
		fmt.Printf("wait %v\n", wait)

		timer := time.NewTimer(wait)

		select {
		case <-timer.C:
			// fmt.Printf("%v\n", monitor.nextCheck)
			monitor.nextCheck = time.Now().Add(cfg.Interval)
			heap.Push(&mh, monitor)
		}
	}
}
