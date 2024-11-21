# Ristretto Lite 
A light-weight version of ristretto cache originated from https://github.com/dgraph-io/ristretto

Ristretto Lite is a fast, concurrent cache library built with focus on performance and correctness, but with simpler implementation.

it is highly recommended to use the cache in large scale scenario while get operations are significantly more frequent than set/remove operation


# Features
* **Fast Throughput** - The implementation uses the similar trick in the original Ristretto to manage contention in large-scale concurrency scenario.

* **Cost-Based Eviction** - The implementation also uses the cost-based eviction, any large new items can evict multiple smaller items. And the user defines what cost means. The difference is the light version uses priority queue approach.

* **Fully Concurrent** - you can use as many goroutine as you want with little throughput degradation.

* **Simple API** -  it has standard cache interfaces like `Put`, `Get`, `Remove`, `Clear`.it offers control workflow interface like `Wait` and `Close`. It also provides stats interface like `Size` and `Cost`.


## Config File
```go

type Config struct {

	// MaxCost describes the maximum cost the cache can hold, it can be any arbitrary number
	// the cost can be your estimation of the memory usage of the cached item
	// the cost can be the weight or deemed value of the cached item
	// it is the only way you manage the size of the cache
	MaxCost int

	// NumShards describes the number of shards for the cache, more shards means less contention in setting and getting the items in high concurrency scenario, but it can also introduce overhead
	// suggestion: using the default value
	NumShards uint64

	// MaxSetBufferSize buffer size describes the maximum number of items can live in the buffer at once waiting to be added or removed
	// if the buffer is full, the Put operation will be contested and blocked, a large buffer size can reduce contention
	// but it can also introduce overhead in memory footprint
	// suggestion: using the default value
	MaxSetBufferSize int

	// CleanupIntervalMilli describes the interval in milliseconds to run the cleanup operation to clean up items that are expired
	// lower interval means higher rate of inspecting and evicting expired item and potentially more, also potentially delaying the process of putting and removing operation.
	// suggestion: do not set it too
	CleanupIntervalMilli int
}

func DefaultConfig() Config {
	return Config{
		MaxCost:              1 << 30,   // 1GB
		NumShards:            256,       // the number of shards for the cache
		MaxSetBufferSize:     32 * 1024, // 32768, the number of items can live in the buffer at once waiting to be added or removed
		CleanupIntervalMilli: 10000,      // 10 seconds
	}
}


```

## Sample
```go
func main() {
	cache, err := ristrettolite.NewCache(Config{
		MaxCost:              1 << 30,   // 1GB
		NumShards:            256,       // the number of shards for the cache
		MaxSetBufferSize:     32 * 1024, // 32768, the number of items can live in the buffer at once waiting to be added or removed
		CleanupIntervalMilli: 10000,      // 10 seconds
	})
	if err != nil {
		panic(err)
	}
	defer cache.Close()

	// set a value with a cost of 1
	cache.Set("key", "value", 1)

	// wait for value to pass through buffers
	cache.Wait()

	// get value from cache
	value, found := cache.Get("key")
	if !found {
		panic("missing value")
	}
	fmt.Println(value)

	// del value from cache
	cache.Del("key")
}

## Main.go

`main.go` has some experimentation on how to use this cache associated with emission API
