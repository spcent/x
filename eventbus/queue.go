package eventbus

import "time"

type job struct {
	evt      Event
	priority int
	ts       time.Time
}

type prioItem struct {
	value job
	index int
}

type prioQueue []*prioItem

func (pq prioQueue) Len() int { return len(pq) }
func (pq prioQueue) Less(i, j int) bool {
	if pq[i].value.priority == pq[j].value.priority {
		return pq[i].value.ts.Before(pq[j].value.ts)
	}
	return pq[i].value.priority > pq[j].value.priority
}
func (pq prioQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i]; pq[i].index = i; pq[j].index = j }
func (pq *prioQueue) Push(x any)   { *pq = append(*pq, x.(*prioItem)) }
func (pq *prioQueue) Pop() any {
	old := *pq
	n := len(old)
	it := old[n-1]
	*pq = old[:n-1]
	return it
}
