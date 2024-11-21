package ristrettolite

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEvictionPolicy_Insert(t *testing.T) {
	tests := []struct {
		name                    string
		maxCost                 int
		InsertItems             []*Item[string]
		expectedEvictedLen      int
		expectedExistingItemLen int
		ExpectedSuccess         []bool
		expectedCost            int
	}{
		{
			name:    "Insert a single item below max cost",
			maxCost: 100,
			InsertItems: []*Item[string]{
				{
					Key:   1,
					Value: "Item1",
					Cost:  10,
				},
			},
			ExpectedSuccess:         []bool{true},
			expectedExistingItemLen: 1,
			expectedEvictedLen:      0,
			expectedCost:            10,
		},

		{
			name:    "Insert a single item above max cost",
			maxCost: 100,
			InsertItems: []*Item[string]{
				{
					Key:   1,
					Value: "Item1",
					Cost:  500,
				},
			},
			ExpectedSuccess:         []bool{false},
			expectedEvictedLen:      0,
			expectedCost:            0,
			expectedExistingItemLen: 0,
		},

		{
			name:    "Insert two items to evict the first one",
			maxCost: 100,
			InsertItems: []*Item[string]{
				{
					Key:   1,
					Value: "Item1",
					Cost:  10,
				},
				{
					Key:   2,
					Value: "Item2",
					Cost:  100,
				},
			},
			ExpectedSuccess:         []bool{true, true},
			expectedEvictedLen:      1,
			expectedCost:            100,
			expectedExistingItemLen: 1,
		},

		{
			name:    "Insert two items to update the first one",
			maxCost: 100,
			InsertItems: []*Item[string]{
				{
					Key:   1,
					Value: "Item1",
					Cost:  10,
				},
				{
					Key:   1,
					Value: "Item2",
					Cost:  50,
				},
			},
			ExpectedSuccess:         []bool{true, true},
			expectedEvictedLen:      0,
			expectedCost:            50,
			expectedExistingItemLen: 1,
		},
		{
			name:    "Insert three items, but the 3rd insertion will update the 2nd item and evict itself",
			maxCost: 20,
			InsertItems: []*Item[string]{
				{
					Key:   1,
					Value: "Item1",
					Cost:  15,
				},
				{
					Key:   2,
					Value: "Item2",
					Cost:  5,
				},
				{
					Key:   2,
					Value: "Item3",
					Cost:  6,
				},
			},
			ExpectedSuccess:         []bool{true, true, false},
			expectedCost:            15,
			expectedEvictedLen:      1,
			expectedExistingItemLen: 1,
		},

		{
			name:    "Insert two items to update the first one, but the cost is greater than max cost",
			maxCost: 100,
			InsertItems: []*Item[string]{
				{
					Key:   1,
					Value: "Item1",
					Cost:  10,
				},
				{
					Key:   1,
					Value: "Item2",
					Cost:  110,
				},
			},
			ExpectedSuccess:         []bool{true, false},
			expectedEvictedLen:      0,
			expectedCost:            10,
			expectedExistingItemLen: 1,
		},
		{
			name:    "Insert three items, the 3rd insertion will evict itself",
			maxCost: 100,
			InsertItems: []*Item[string]{
				{
					Key:   1,
					Value: "Item1",
					Cost:  10,
				},
				{
					Key:   2,
					Value: "Item2",
					Cost:  85,
				},
				{
					Key:   3,
					Value: "Item3",
					Cost:  6,
				},
			},
			ExpectedSuccess:         []bool{true, true, false},
			expectedEvictedLen:      1,
			expectedCost:            95,
			expectedExistingItemLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize eviction policy
			ep := newEvictionPolicy[string](tt.maxCost)

			evictedItems := make([]*Item[string], 0)

			// Insert the new item
			for i, item := range tt.InsertItems {
				evicted, success := ep.Insert(item)
				evictedItems = append(evictedItems, evicted...)
				assert.Equal(t, tt.ExpectedSuccess[i], success, "insert success signal mismatch")
			}

			// Check the cost
			assert.Equal(t, tt.expectedCost, ep.curCost, "cost mismatch")

			// check the evicted items length
			assert.Equal(t, tt.expectedEvictedLen, len(evictedItems), "evicted items count mismatch")

			// check the existing item length
			assert.Equal(t, tt.expectedExistingItemLen, len(ep.itemTracker), "existing item count mismatch in tracker")
			assert.Equal(t, tt.expectedExistingItemLen, ep.costQueue.Len(), "existing item count mismatch in cost queue")

		})
	}
}

func TestEvictionPolicy_Remove(t *testing.T) {
	tests := []struct {
		name                    string
		maxCost                 int
		InsertItems             []*Item[string]
		RemoveKeys              []uint64
		expectedExistingItemLen int
		expectedCost            int
	}{
		{
			name:    "Insert two items and delete one",
			maxCost: 100,
			InsertItems: []*Item[string]{
				{
					Key:   1,
					Value: "Item1",
					Cost:  10,
				},
				{
					Key:   2,
					Value: "Item1",
					Cost:  20,
				},
			},
			RemoveKeys:              []uint64{1},
			expectedExistingItemLen: 1,
			expectedCost:            20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize eviction policy
			ep := newEvictionPolicy[string](tt.maxCost)

			// Insert the new item
			for _, item := range tt.InsertItems {
				ep.Insert(item)

			}

			for _, key := range tt.RemoveKeys {
				ep.Delete(key)
			}

			// Check the cost
			assert.Equal(t, tt.expectedCost, ep.curCost, "cost mismatch")

			// check the existing item length
			assert.Equal(t, tt.expectedExistingItemLen, len(ep.itemTracker), "existing item count mismatch in tracker")
			assert.Equal(t, tt.expectedExistingItemLen, ep.costQueue.Len(), "existing item count mismatch in cost queue")

		})
	}
}

func TestEvictionPolicy_EvictExpiredItem(t *testing.T) {
	tests := []struct {
		name                    string
		maxCost                 int
		InsertItems             []*Item[string]
		ExpireTime              time.Time
		expectedEvictedLen      int
		expectedExistingItemLen int
		expectedCost            int
	}{
		{
			name:       "Insert two items and delete one",
			maxCost:    100,
			ExpireTime: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			InsertItems: []*Item[string]{
				{
					Key:      1,
					Value:    "Item1",
					Cost:     10,
					ExpireAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					Key:      2,
					Value:    "Item2",
					Cost:     20,
					ExpireAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			expectedEvictedLen:      1,
			expectedExistingItemLen: 1,
			expectedCost:            20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize eviction policy
			ep := newEvictionPolicy[string](tt.maxCost)

			// Insert the new item
			for _, item := range tt.InsertItems {
				ep.Insert(item)
			}

			evictedItems := ep.EvictExpiredItems(tt.ExpireTime)

			// Check the evicted items length
			assert.Equal(t, tt.expectedEvictedLen, len(evictedItems), "evicted items count mismatch")

			// Check the cost
			assert.Equal(t, tt.expectedCost, ep.curCost, "cost mismatch")

			// check the existing item length
			assert.Equal(t, tt.expectedExistingItemLen, len(ep.itemTracker), "existing item count mismatch in tracker")
			assert.Equal(t, tt.expectedExistingItemLen, ep.costQueue.Len(), "existing item count mismatch in cost queue")

		})
	}
}
