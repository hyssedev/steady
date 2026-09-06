package scheduler

type MonitorHeap []*scheduledMonitor

func (mh MonitorHeap) Len() int {
	return len(mh)
}

func (mh MonitorHeap) Less(i, j int) bool {
	return mh[i].nextCheck.Before(mh[j].nextCheck)
}

func (mh MonitorHeap) Swap(i, j int) {
	mh[i], mh[j] = mh[j], mh[i]
}

func (mh *MonitorHeap) Push(x any) {
	monitor := x.(*scheduledMonitor)
	*mh = append(*mh, monitor)
}

func (mh *MonitorHeap) Pop() any {
	old := *mh
	n := len(old)
	monitor := old[n-1]
	old[n-1] = nil
	*mh = old[0 : n-1]
	return monitor
}
