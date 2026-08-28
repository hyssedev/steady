package scheduler

import (
	"container/heap"
	"time"
)

type MonitorHeap []*scheduledMonitor

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
	monitor := x.(*scheduledMonitor)
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

func (mh *MonitorHeap) update(monitor *scheduledMonitor, nextCheck time.Time) {
	monitor.nextCheck = nextCheck
	heap.Fix(mh, monitor.index)
}
