go test -count=1 -run TestMemoryOneMillionKeys -v

go test -count=1 -run TestConcurrentReadWriteDelete -v 