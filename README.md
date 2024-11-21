# Ristretto Lite 
A light-weight version of ristretto cache originated from https://github.com/dgraph-io/ristretto

Ristretto Lite is a fast, concurrent cache library built with focus on performance and correctness, but with simpler implementation.

it is highly recommended to use the cache in large scale scenario while get operations are significantly more frequent than set/delete operation


# Features
* **Fast Throughset** - The implementation uses the similar trick in the original Ristretto to manage contention in large-scale concurrency scenario.

* **Cost-Based Eviction** - The implementation also uses the cost-based eviction, any large new items can evict multiple smaller items. And the user defines what cost means. The difference is the light version uses priority queue approach.

* **Fully Concurrent** - you can use as many goroutine as you want with little throughset degradation.

* **Simple API** -  it has standard cache interfaces like `Set`, `Get`, `Delete`, `Clear`.it offers control workflow interface like `Wait` and `Close`. It also provides stats interface like `Size` and `Cost`.


## Config File
```go

type Config struct {

	// MaxCost the maximum cost the cache can hold, it can be any arbitrary number
	// the cost can be your estimation of the memory usage of the cached item
	// the cost can be the weight or deemed value of the cached item
	// it is the only way you manage the size of the cache
	MaxCost int

	// NumShards describes the number of shards for the cache, more shards means less contention in setting and getting the items
	// but it can also introduce more overhead
	// suggestion - use number that is a power of 2
	NumShards uint64

	// MaxSetBufferSize describes the maximum number of items can live in the buffer at once waiting to be added or removed
	// if the buffer is full, the set operation will be contested and will fail, a large buffer size can reduce contention but can also add more memory overhead
	MaxSetBufferSize int

	// CleanupIntervalMilli describes the interval in milliseconds to run the cleanup operation to clean up items that are expired
	// The lower the value, the more frequent the cleanup operation will run while it can also introduce delay in the processing of setting and removing items.
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
	// you can also use ristrettolite.DefaultConfig()
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
	cache.Set("key", "value", 1, 10000)

	// wait for value to pass through buffers
	cache.Wait()

	// get value from cache
	value, found := cache.Get("key")
	if !found {
		panic("missing value")
	}
	fmt.Println(value)

	// delete value from cache
	cache.Delete("key")
}
```

## Main.go

`main.go` has some play-around on using the cache associated with measurement API.
