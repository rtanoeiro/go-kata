go test -count=1 -run TestMemoryOneMillionKeys -v

go test -count=1 -run TestConcurrentReadWriteDelete -v 

go test -bench=Benchmark_ContentionOneShardedMap -benchmem -cpuprofile=cpu_contention_1shard.prof

go test -bench=Benchmark_ContentionSixtyFourShardedMap -benchmem -cpuprofile=cpu_contention_64shard.prof

go test -bench=Benchmark_OneMillionIntShardedMap -benchmem -cpuprofile=cpu_MillionMemory.prof

go tool pprof -top cpu_contention_1shard.prof

go tool pprof -top cpu_MillionMemory.prof