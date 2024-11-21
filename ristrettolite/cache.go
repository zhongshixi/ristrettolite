package ristrettolite

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cespare/xxhash/v2"
)

type Config struct {

	// max cost describes the maximum cost the cache can hold, it can be any arbitrary number
	// the cost can be your estimation of the memory usage of the cached item
	// the cost can be the weight or deemed value of the cached item
	// it is the only way you manage the size of the cache
	MaxCost int

	// num shards describes the number of shards for the cache, more shards means less contention in setting and getting the items
	// but it can also introduce more memory overhead
	// suggest using the default value which is also the default of the original ristretto cache
	NumShards uint64

	// max set buffer size describes the maximum number of items can live in the buffer at once waiting to be added or removed
	// if the buffer is full, the Put operation will be contested and blocked, a large buffer size can reduce contention
	// but it can also introduce more memory overhead
	// suggest using the default value  which is also the default of the original ristretto cache
	MaxSetBufferSize int

	// cleanup interval milliseconds describes the interval in milliseconds to run the cleanup operation to clean up items that are expired
	CleanupIntervalMilli int
}

func DefaultConfig() Config {
	return Config{
		MaxCost:              1 << 30,   // 1GB
		NumShards:            256,       // - the number of shards for the cache
		MaxSetBufferSize:     32 * 1024, // 32768, the number of items can live in the buffer at once waiting to be added or removed
		CleanupIntervalMilli: 60000,     // 60 seconds
	}
}

func (c *Config) Validate() error {
	if c.MaxCost <= 0 {
		return fmt.Errorf("MaxCost must be greater than 0")
	}

	if c.NumShards <= 0 {
		return fmt.Errorf("NumShards must be greater than 0")
	}

	if c.CleanupIntervalMilli <= 0 {
		return fmt.Errorf("CleanupIntervalMilli must be greater than 0")
	}

	if c.MaxSetBufferSize <= 0 {
		return fmt.Errorf("MaxSetBufferSize must be greater than 0")
	}

	return nil
}

// Cache is a concurrent safe cache that supports Put, Get, Remove, and Wait operations
type Cache[V any] struct {
	conf Config

	shardedMap *shardedMap[V]

	evictionPolicy *evictionPolicy[V]

	cleanupTicker *time.Ticker

	setBuf chan *Item[V]

	stopSig chan struct{}

	isClosed atomic.Bool
}

func NewCache[V any](conf Config) (*Cache[V], error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	c := &Cache[V]{
		conf:           conf,
		shardedMap:     newShardedMap[V](conf.NumShards),
		cleanupTicker:  time.NewTicker(time.Duration(conf.CleanupIntervalMilli) * time.Millisecond),
		setBuf:         make(chan *Item[V], conf.MaxSetBufferSize),
		evictionPolicy: newEvictionPolicy[V](conf.MaxCost),
		stopSig:        make(chan struct{}),
	}

	go c.processItems()

	return c, nil
}

// Put returns false if the cost is 0 or ttl is 0 or cache is closed or buffer is full(indicating contention and failed operation)
// it does not immediately adds a new item to the cache, instead it adds the item to the buffer for background processing, eventually the item will be added to the cache or evicted by the policy
// it runs in O(1) time complexity
func (c *Cache[V]) Put(key string, value V, cost int, ttlMillis int) bool {
	if c == nil || c.isClosed.Load() {
		return false
	}

	if cost <= 0 {
		return false
	}

	if ttlMillis <= 0 {
		return false
	}

	item := &Item[V]{
		// this hash algorithm is fast and efficient - see https://github.com/cespare/xxhash
		// the hash algorithm is using XXH64 which is very fast https://xxhash.com/
		Key:      xxhash.Sum64String(key),
		Value:    value,
		Cost:     cost,
		ExpireAt: time.Now().Add(time.Duration(ttlMillis) * time.Millisecond),
		Action:   ActionPut,
	}

	select {
	case c.setBuf <- item:
		return true
	default:
		return false
	}

}

// Get returns the value of the key if it exists in the cache immediately without blocking, it runs in O(1) time complexity
// it returns the value and true if the key exists, otherwise it returns nil and false
// it will return the value even if the value is expired, let the consumer decide what to do with the expired value
func (c *Cache[V]) Get(key string) (V, bool) {
	if c == nil || c.isClosed.Load() {
		return zeroValue[V](), false
	}

	return c.shardedMap.Get(xxhash.Sum64String(key))
}

// Remove marks the item to be removed from the cache, it does not immediately remove the item from the cache
// instead it adds the item to the buffer for background processing, eventually the item will be removed from the cache
func (c *Cache[V]) Remove(key string) {
	if c == nil || c.isClosed.Load() {
		return
	}

	item := &Item[V]{
		Key:    xxhash.Sum64String(key),
		Action: ActionRemove,
	}

	c.setBuf <- item
}

// Wait blocks until the all the items in the set buffer added before Wait() is invoked are processed
func (c *Cache[V]) Wait() {
	if c == nil || c.isClosed.Load() {
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	item := &Item[V]{
		Action:    ActionWait,
		WaitGroup: wg,
	}

	go func() {
		c.setBuf <- item
	}()

	wg.Wait()

}

// processItems is a background goroutine that processes the items in the set buffer
func (c *Cache[V]) processItems() {
	for {
		select {
		case item := <-c.setBuf:
			switch item.Action {
			// if the action is wait, then signal the corresponding wait group it is done, so Wait() can return
			case ActionWait:
				if item.WaitGroup != nil {
					item.WaitGroup.Done()
				}

			// if the action is put, then insert the item to the cache and evict items if necessary
			case ActionPut:
				// figure out if the item can be inserted to the cache or not
				// the evictedItem could contain the item that was just inserted
				evictedItems, ok := c.evictionPolicy.Insert(item)
				if ok {
					c.shardedMap.Put(item)
				}
				for _, evictedItem := range evictedItems {
					// if the evicted item is not the same as the item that was just inserted, then remove the evicted item from the cache
					if evictedItem.Key != item.Key {
						c.shardedMap.Remove(evictedItem.Key)
					}
				}
			case ActionRemove:
				_, ok := c.evictionPolicy.Remove(item.Key)
				if ok {
					c.shardedMap.Remove(item.Key)
				}
			}

		case <-c.cleanupTicker.C:
			evictedItems := c.evictionPolicy.EvictExpiredItems(time.Now())
			for _, evictedItem := range evictedItems {
				c.shardedMap.Remove(evictedItem.Key)
			}
		case <-c.stopSig:
			return
		}
	}
}

// Clear clears the cache,  it stops the background processing
// during the clearance it is suggested that user should not call Put, Get, Remove, Wait operations to avoid delay in the clearance
func (c *Cache[V]) Clear() {
	if c == nil || c.isClosed.Load() {
		return
	}

	// signal the stop signal to stop the processItems goroutine
	// block until the processItems goroutine is stopped
	c.stopSig <- struct{}{}

	// clear the rest of set buffer items
loop:
	for {
		select {
		case item := <-c.setBuf:
			if item.WaitGroup != nil {
				item.WaitGroup.Done()
				continue
			}

		default:
			break loop
		}
	}

	// clear the eviction policy and sharded map
	c.evictionPolicy.Clear()
	c.shardedMap.Clear()

	// restart the go routine to process items again
	go c.processItems()
}

func (c *Cache[V]) Close() {
	if c == nil || c.isClosed.Load() {
		return
	}
	c.Clear()

	// Block until processItems goroutine is returned.
	c.stopSig <- struct{}{}
	close(c.stopSig)
	close(c.setBuf)
	c.cleanupTicker.Stop()
	c.isClosed.Store(true)
}

// Size returns the number of items presented in the cache
func (c *Cache[V]) Size() int {
	return c.evictionPolicy.Size()
}

// Cost returns current collected cost of the cache
func (c *Cache[V]) Cost() int {
	return c.evictionPolicy.Cost()
}
