package shard

import (
	"runtime"
	"sync"
	"testing"
)

const (
	NUM_ITERATIONS = 1000000
	NUM_GOROUTINES = 8
)

func benchmarhIntShardedSet(b *testing.B, numShards int) {
	shardedMap := CreateShardedMap[int, int](numShards)
	numGoroutines := runtime.GOMAXPROCS(NUM_GOROUTINES)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			shardedMap.Set(i%10000, i)
			i++
		}
	})
	b.ReportMetric(float64(numGoroutines), "goroutines")
}

func Benchmark_ContentionOneShardedMap(b *testing.B) {
	benchmarhIntShardedSet(b, 1)
}

func Benchmark_ContentionSixtyFourShardedMap(b *testing.B) {
	benchmarhIntShardedSet(b, 64)
}

func Benchmark_OneMillionIntShardedMap(b *testing.B) {
	benchmarhIntShardedSet(b, 64)
}

func TestConcurrentReadWriteDelete(t *testing.T) {
	m := CreateShardedMap[int, int](64)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				key := (i * 8) + g
				m.Set(key, i)
				_, _ = m.Get(key)
				if i%2 == 0 {
					m.Delete(key)
				}
				if i%100 == 0 {
					_ = m.Keys()
				}
			}
		}(g)
	}
	wg.Wait()
}

func TestMemoryOneMillionKeys(t *testing.T) {
	const n = 1_000_000
	const mb = 1024 * 1024

	baseline := heapAllocDelta(func() any {
		m := make(map[int]any, n)
		for i := 0; i < n; i++ {
			m[i] = i
		}
		return m
	})

	sharded := heapAllocDelta(func() any {
		m := CreateShardedMap[int, any](64)
		for i := 0; i < n; i++ {
			m.Set(i, i)
		}
		return m
	})

	const maxExtraBytes = 50 * 1024 * 1024
	extra := int64(sharded) - int64(baseline)
	t.Logf("baseline=%d MB, sharded=%d MB, extra=%d MB", baseline/mb, sharded/mb, extra/mb)
	if extra > maxExtraBytes {
		t.Fatalf("sharded map used %d bytes more than the baseline map (limit %d)", extra, maxExtraBytes)
	}
}

func heapAllocDelta(fn func() any) uint64 {
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	obj := fn()
	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(obj)
	return after.HeapAlloc - before.HeapAlloc
}
