package ristrettolite

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_setAndGet(t *testing.T) {
	conf := DefaultConfig()
	conf.MaxCost = 1000
	conf.NumShards = 8
	conf.CleanupIntervalMilli = 100
	cache, err := NewCache[string](conf)
	assert.NoError(t, err, "Cache initialization should not fail")

	defer cache.Close()

	t.Run("Set and Get valid item", func(t *testing.T) {
		success := cache.Set("key1", "value1", 10, 1000) // TTL = 1 second
		assert.True(t, success, "Set should succeed")

		// wait for the item to store in the cache
		cache.Wait()

		val, ok := cache.Get("key1")
		assert.True(t, ok, "Get should succeed for an existing key")
		assert.Equal(t, "value1", val, "Get should return the correct value")
	})

	t.Run("set and Get an expired item, then after being evicted get should receive no item", func(t *testing.T) {
		success := cache.Set("key1", "value1", 10, 1) // TTL = 1 milliseconds
		assert.True(t, success, "set should succeed")

		time.Sleep(10 * time.Millisecond)
		cache.Wait()

		val, ok := cache.Get("key1")
		assert.True(t, ok, "Get should succeed for an existing key")
		assert.Equal(t, "value1", val, "Get should return the correct value")

		// ensure the eviction kicks and the item is removed
		time.Sleep(200 * time.Millisecond)

		val, ok = cache.Get("key1")
		assert.False(t, ok, "Get should failed since the item is expired")
		assert.Equal(t, "", val, "Get should return the correct empty value")
	})

	t.Run("set and clear", func(t *testing.T) {
		success := cache.Set("key1", "value1", 10, 1) // TTL = 1 milliseconds
		assert.True(t, success, "set should succeed")

		cache.Wait()
		val, ok := cache.Get("key1")
		assert.True(t, ok, "Get should succeed for an existing key")
		assert.Equal(t, "value1", val, "Get should return the correct value")

		cache.Clear()

		val, ok = cache.Get("key1")
		assert.False(t, ok, "Get should failed since the item is expired")
		assert.Equal(t, "", val, "Get should return the correct empty value")

		success = cache.Set("key2", "value2", 10, 1)
		assert.True(t, success, "set should succeed")

		cache.Wait()

		val, ok = cache.Get("key2")
		assert.True(t, ok, "Get should succeed for an existing key")
		assert.Equal(t, "value2", val, "Get should return the correct value")

	})

	t.Run("set and close", func(t *testing.T) {
		success := cache.Set("key1", "value1", 10, 1) // TTL = 1 milliseconds
		assert.True(t, success, "set should succeed")

		cache.Wait()
		val, ok := cache.Get("key1")
		assert.True(t, ok, "Get should succeed for an existing key")
		assert.Equal(t, "value1", val, "Get should return the correct value")

		cache.Close()

		val, ok = cache.Get("key1")
		assert.False(t, ok, "Get should failed since the item is expired")
		assert.Equal(t, "", val, "Get should return the correct empty value")
	})

	t.Run("set and delete", func(t *testing.T) {
		success := cache.Set("key1", "value1", 10, 1) // TTL = 1 milliseconds
		assert.True(t, success, "set should succeed")

		cache.Wait()
		val, ok := cache.Get("key1")
		assert.True(t, ok, "Get should succeed for an existing key")
		assert.Equal(t, "value1", val, "Get should return the correct value")

		cache.Delete("key1")

		cache.Wait()

		val, ok = cache.Get("key1")
		assert.False(t, ok, "Get should failed since the item is expired")
		assert.Equal(t, "", val, "Get should return the correct empty value")
	})
}
