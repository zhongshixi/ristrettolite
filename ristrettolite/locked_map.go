package ristrettolite

import (
	"sync"
)

type lockedMap[V any] struct {
	sync.RWMutex
	data map[uint64]*Item[V]
}

func newLockedMap[V any]() *lockedMap[V] {
	return &lockedMap[V]{
		data: make(map[uint64]*Item[V]),
	}
}

func (m *lockedMap[V]) get(key uint64) (V, bool) {
	m.RLock()
	defer m.RUnlock()
	v, ok := m.data[key]
	if !ok {
		return zeroValue[V](), false
	}

	// return value that could be already expired, let the consumer handle the expired value
	return v.Value, ok
}

func (m *lockedMap[V]) put(item *Item[V]) {
	if item == nil {
		return
	}
	m.Lock()
	defer m.Unlock()
	m.data[item.Key] = item
}

func (m *lockedMap[V]) remove(key uint64) (V, bool) {
	m.Lock()
	defer m.Unlock()

	v, ok := m.data[key]
	if !ok {
		return zeroValue[V](), false
	}

	delete(m.data, key)

	return v.Value, true
}

func (m *lockedMap[V]) clear() {
	m.Lock()
	defer m.Unlock()

	// Note:
	// re-init the map, the old map will be garbage collected by GC
	// this is O(1) operation, but if the map was very large, GC might take some time to clean up the old map
	// may cause the program to pause for a while
	m.data = make(map[uint64]*Item[V])
}
