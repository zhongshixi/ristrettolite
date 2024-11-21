package ristrettolite

import (
	"container/heap"
	"time"
)

type evictionPolicy[V any] struct {
	maxCost int
	curCost int

	itemTracker map[uint64]*Item[V]
	costQueue   *CostPriorityQueue[V]
}

func newEvictionPolicy[V any](maxCost int) *evictionPolicy[V] {
	return &evictionPolicy[V]{
		maxCost:     maxCost,
		curCost:     0,
		itemTracker: make(map[uint64]*Item[V]),
		costQueue:   &CostPriorityQueue[V]{},
	}
}

// update updates the item in the eviction policy and returns the evicted items and whether the item is inserted or not
func (ep *evictionPolicy[V]) update(item *Item[V]) ([]*Item[V], bool) {
	prevItem, ok := ep.itemTracker[item.Key]
	if ok {

		// if the new cost is the same as the previous cost, we don't need to do any evict
		// if the new cost is the less than the previous cost, then we just need to change the cost of the item in the queue
		// if the new cost is greater than the previous cost, then we need to evict some items
		addedCost := item.Cost - prevItem.Cost
		prevItem.Cost = item.Cost
		prevItem.Value = item.Value
		prevItem.ExpireAt = item.ExpireAt

		if addedCost == 0 {
			return nil, true
		}

		ep.curCost += addedCost
		if ep.curCost < ep.maxCost {
			heap.Fix(ep.costQueue, prevItem.pqIndex)
			return nil, true

		}

		heap.Fix(ep.costQueue, prevItem.pqIndex)
		items := ep.evictUntilRoomLeft()

		// that means the item is not inserted due to cost is too low so it gets evicted immediately
		if _, ok := ep.itemTracker[item.Key]; !ok {
			return items, false
		}

		return items, true

	}

	return nil, false
}

// Insert insert the item into the eviction policy and return items evicted and whether the item is inserted or not
// if the item is not inserted, it will return false
// if the item tries to be inserted and get evicted immediately, it will return the item in the evicted items and false
func (ep *evictionPolicy[V]) Insert(item *Item[V]) ([]*Item[V], bool) {
	if item == nil {
		return nil, false
	}

	if item.Cost > ep.maxCost {
		return nil, false
	}

	_, ok := ep.itemTracker[item.Key]
	if ok {
		return ep.update(item)
	}

	evictedItems := make([]*Item[V], 0)

	heap.Push(ep.costQueue, item)
	ep.curCost += item.Cost
	ep.itemTracker[item.Key] = item

	if ep.curCost > ep.maxCost {
		evictedItems = append(evictedItems, ep.evictUntilRoomLeft()...)
	}

	// that means the item is not inserted due to cost is too low so it gets evicted immediately
	if _, ok := ep.itemTracker[item.Key]; !ok {
		return evictedItems, false
	}

	return evictedItems, true

}

// evictUntilRoomLeft evicts items until the current cost is less than the max cost
func (ep *evictionPolicy[V]) evictUntilRoomLeft() []*Item[V] {
	evicted := make([]*Item[V], 0)
	for ep.curCost > ep.maxCost {
		if ep.costQueue.Len() == 0 {
			break
		}

		item := heap.Pop(ep.costQueue).(*Item[V])
		delete(ep.itemTracker, item.Key)
		ep.curCost -= item.Cost
		evicted = append(evicted, item)
	}

	return evicted
}

// Remove removes the item from the eviction policy and returns the item and whether the item is removed or not
func (ep *evictionPolicy[V]) Remove(key uint64) (*Item[V], bool) {
	item, ok := ep.itemTracker[key]
	if !ok {
		return nil, false
	}
	heap.Remove(ep.costQueue, item.pqIndex)
	delete(ep.itemTracker, item.Key)
	ep.curCost -= item.Cost
	return item, true
}

// EvictExpiredItems evicts all the items that have expired before the given timestamp
func (ep *evictionPolicy[V]) EvictExpiredItems(ts time.Time) []*Item[V] {
	evicted := make([]*Item[V], 0)

	for _, item := range ep.itemTracker {
		if item.ExpireAt.After(ts) {
			continue
		}

		heap.Remove(ep.costQueue, item.pqIndex)
		delete(ep.itemTracker, item.Key)
		ep.curCost -= item.Cost
		evicted = append(evicted, item)
	}

	return evicted
}

// Clear resets the eviction policy by reinitializing the cost queue and item tracker and lets GC handle the rest
func (ep *evictionPolicy[V]) Clear() {
	// Note: GC may be slow depending on the size of the map
	ep.costQueue = &CostPriorityQueue[V]{}
	ep.itemTracker = make(map[uint64]*Item[V])
	ep.curCost = 0
}

// Size returns the number of items in the eviction policy
func (ep *evictionPolicy[V]) Size() int {
	return len(ep.itemTracker)
}

// Cost returns the current cost of the eviction policy
func (ep *evictionPolicy[V]) Cost() int {
	return ep.curCost
}
