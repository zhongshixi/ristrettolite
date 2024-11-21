package ristrettolite

import (
	"sync"
)

// shardedMap is a map that is sharded into multiple locked maps
//
// in high-concurrency scenarios, instead of using one map and one lock, by sharding the map into multiple locked maps,
// we can use lock for each sharded map, thus when multiple set/get/delete operations are happening on different keys, they are less likely
// to block each other, thus reducing contention
type shardedMap[V any] struct {
	sync.RWMutex
	shards []*lockedMap[V]

	numShards uint64
}

func newShardedMap[V any](numShards uint64) *shardedMap[V] {
	sm := &shardedMap[V]{
		shards:    make([]*lockedMap[V], numShards),
		numShards: numShards,
	}

	for i := range sm.shards {
		sm.shards[i] = newLockedMap[V]()
	}
	return sm
}

func (sm *shardedMap[V]) Get(key uint64) (V, bool) {
	return sm.shards[key%sm.numShards].get(key)
}

func (sm *shardedMap[V]) Put(i *Item[V]) {
	if i == nil {
		return
	}
	sm.shards[i.Key%sm.numShards].put(i)
}

func (sm *shardedMap[V]) Remove(key uint64) (V, bool) {
	return sm.shards[key%sm.numShards].remove(key)
}

func (sm *shardedMap[V]) Clear() {
	for _, shard := range sm.shards {
		shard.clear()
	}
}
