package ristrettolite

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLockedMap_Put(t *testing.T) {
	tests := []struct {
		name        string
		putItem     *Item[string]
		getKey      uint64
		expectedVal string
		expectedOk  bool
	}{
		{
			name: "Put valid item",
			putItem: &Item[string]{
				Key:      1,
				Value:    "value1",
				Cost:     10,
				ExpireAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			getKey:      1,
			expectedVal: "value1",
			expectedOk:  true,
		},
		{
			name:        "Put nil item",
			putItem:     nil,
			getKey:      1,
			expectedVal: "",
			expectedOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newLockedMap[string]()

			// Perform Put operation
			m.put(tt.putItem)

			// Test Get
			val, ok := m.get(tt.getKey)

			// Assertions
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedVal, val)
		})
	}
}

func TestLockedMap_Remove(t *testing.T) {
	tests := []struct {
		name        string
		initialData []*Item[string]
		removeKey   uint64
		expectedVal string
		expectedOk  bool
	}{
		{
			name: "Remove existing key",
			initialData: []*Item[string]{
				{
					Key:      1,
					Value:    "value1",
					Cost:     10,
					ExpireAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			removeKey:   1,
			expectedVal: "value1",
			expectedOk:  true,
		},
		{
			name:        "Remove non-existing key",
			initialData: nil,
			removeKey:   2,
			expectedVal: "",
			expectedOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newLockedMap[string]()

			// Populate initial data
			for _, item := range tt.initialData {
				m.put(item)
			}

			// Perform Remove operation
			val, ok := m.remove(tt.removeKey)

			// Assertions
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedVal, val)

			// Ensure the key is no longer in the map
			_, exists := m.get(tt.removeKey)
			assert.False(t, exists, "Key should not exist after removal")
		})
	}
}

func TestLockedMap_Concurrency(t *testing.T) {
	m := newLockedMap[int]()

	// Prepare a wait group to synchronize goroutines
	var wg sync.WaitGroup

	// Number of concurrent operations
	const numGoroutines = 100
	const numItems = 50

	// Perform concurrent writes
	for i := 1; i < numGoroutines+1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.put(&Item[int]{
				Key:      uint64(i),
				Value:    i * 10,
				Cost:     i * 2,
				ExpireAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			})
		}(i)
	}

	// Perform concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			val, ok := m.get(uint64(i))
			if ok {
				assert.NotEmpty(t, val, "Value should not be empty for existing keys")
			}
		}(i)
	}

	// Perform concurrent removals
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.remove(uint64(i))
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Verify map state
	for i := 0; i < numItems; i++ {
		val, ok := m.get(uint64(i))
		if ok {
			assert.Equal(t, i*10, val, "Value mismatch after concurrent updates")
		}
	}
}
