package ristrettolite

// CostPriorityQueue implements a priority queue using a min-heap with heap.Interface
type CostPriorityQueue[T any] []*Item[T]

func (pq CostPriorityQueue[T]) Len() int { return len(pq) }

// Less is part of heap.Interface.
// Lower cost has higher priority; for equal cost, earlier ExpireAt has higher priority.
func (pq CostPriorityQueue[T]) Less(i, j int) bool {
	if pq[i].Cost == pq[j].Cost {
		return pq[i].ExpireAt.Before(pq[j].ExpireAt)
	}
	return pq[i].Cost < pq[j].Cost
}

func (pq CostPriorityQueue[T]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].pqIndex = i
	pq[j].pqIndex = j
}

func (pq *CostPriorityQueue[T]) Push(x any) {
	item := x.(*Item[T])
	item.pqIndex = len(*pq)
	*pq = append(*pq, item)
}

func (pq *CostPriorityQueue[T]) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.pqIndex = -1
	*pq = old[:n-1]
	return item
}

func (pq CostPriorityQueue[T]) Peek() *Item[T] {
	if len(pq) == 0 {
		return nil
	}
	return pq[0]
}
