package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zhongshixi/ristrettolite/emission"
	"github.com/zhongshixi/ristrettolite/ristrettolite"
)

const oneDay = 24 * time.Hour
const numOfRequests = 100
const numOfRowsPerRequests = 2

type Stats struct {
	CacheHit  atomic.Int64
	CacheMiss atomic.Int64

	CachePutSuccess atomic.Int64
	CachePutFailed  atomic.Int64

	RequestSuccess atomic.Int64
	RequestFailed  atomic.Int64
}

func (c *Stats) String() string {
	return fmt.Sprintf("cache hit: %d, cache miss: %d, cache put success: %d, cache put failed: %d, request success: %d, request failed: %d", c.CacheHit.Load(), c.CacheMiss.Load(), c.CachePutSuccess.Load(), c.CachePutFailed.Load(), c.RequestSuccess.Load(), c.RequestFailed.Load())
}

func main() {

	config := ristrettolite.Config{
		MaxCost:              150,
		NumShards:            256,
		MaxSetBufferSize:     32 * 1024,
		CleanupIntervalMilli: 10000,
	}

	cache, err := ristrettolite.NewCache[emission.EmissionDataRow](config)
	if err != nil {
		panic(err)
	}

	defer cache.Close()

	api := emission.NewMeasurementAPI("<your_api_key>")

	stats := &Stats{}

	// warm up the cache
	var wg sync.WaitGroup
	wg.Add(numOfRequests)

	payloads := make([]emission.EmissionRequestPayload, numOfRequests)
	for i := 0; i < numOfRequests; i++ {
		payloads[i] = emission.GenerateEmissionRequestPayload(numOfRowsPerRequests)
	}

	for i := 0; i < numOfRequests; i++ {
		go func() {
			defer wg.Done()

			payload := payloads[i]
			ctx := context.Background()
			cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			response, err := api.GetEmissionMeasurement(cctx, payload)
			if err != nil {
				fmt.Printf("request failed: %v\n", err)
				stats.RequestFailed.Add(1)
				return
			}

			stats.RequestSuccess.Add(1)

			reqRowToRespRow := map[string]emission.EmissionDataRow{}
			for _, row := range response.Rows {
				reqRowToRespRow[row.RowIdentifier] = row
			}

			for _, row := range payload.Rows {
				if respRow, ok := reqRowToRespRow[row.RowIdentifier]; ok {
					success := cache.Put(row.String(), respRow, row.Cost, int(oneDay.Milliseconds()))
					if success {
						stats.CachePutSuccess.Add(1)
					} else {
						stats.CachePutFailed.Add(1)
					}
				}
			}
		}()
	}
	wg.Wait()

	// send requests to see if cache is hit or miss
	var wg2 sync.WaitGroup
	wg2.Add(numOfRequests)
	for i := 0; i < numOfRequests; i++ {
		go func() {
			defer wg2.Done()
			payload := payloads[i]

			for _, row := range payload.Rows {
				_, ok := cache.Get(row.String())
				if ok {
					stats.CacheHit.Add(1)
				} else {
					stats.CacheMiss.Add(1)
				}
			}

		}()
	}

	wg2.Wait()

	fmt.Printf("stats: %+v , cache cost: %d, cache size: %d \n", stats.String(), cache.Cost(), cache.Size())

}
