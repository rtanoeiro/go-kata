package shard

import (
	"encoding/binary"
	"hash/fnv"
	"log/slog"
	"sync"
)

type ShardedMap[K comparable, V any] struct {
	Shards  []map[K]V
	Mutexes []sync.RWMutex
	Size    uint64
}

func CreateShardedMap[K comparable, V any](numberShards int) *ShardedMap[K, V] {
	if numberShards <= 0 {
		return nil
	}

	shards := make([]map[K]V, numberShards)
	for index := range shards {
		shards[index] = make(map[K]V)
	}

	mutexes := make([]sync.RWMutex, numberShards)

	shardedMap := ShardedMap[K, V]{
		Shards:  shards,
		Mutexes: mutexes,
		Size:    uint64(len(shards)),
	}
	return &shardedMap
}

func (sm *ShardedMap[K, V]) Get(key K) (V, bool) {
	index := getShardIndex(key, sm.Size)
	sm.Mutexes[index].RLock()
	defer sm.Mutexes[index].RUnlock()
	value, exists := sm.Shards[index][key]
	return value, exists
}

func (sm *ShardedMap[K, V]) Set(key K, value V) {
	index := getShardIndex(key, sm.Size)
	sm.Mutexes[index].Lock()
	defer sm.Mutexes[index].Unlock()
	sm.Shards[index][key] = value
}

func (sm *ShardedMap[K, V]) Delete(key K) {
	index := getShardIndex(key, sm.Size)
	sm.Mutexes[index].Lock()
	defer sm.Mutexes[index].Unlock()
	delete(sm.Shards[index], key)
}

func (sm *ShardedMap[K, V]) Keys() []K {
	var keys []K
	for i := 0; i < int(sm.Size); i++ {
		sm.Mutexes[i].RLock()
		for key := range sm.Shards[i] {
			keys = append(keys, key)
		}
		sm.Mutexes[i].RUnlock()
	}
	return keys
}

func getShardIndex[K comparable](key K, shardSize uint64) uint64 {
	var bytesRpr uint64
	switch v := any(key).(type) {
	case string:
		bytesRpr = getIntegerRepresentationFromString(v)
	case int:
		bytesRpr = getIntegerRepresentationFromint(v)
	default:
		slog.Info("Unsupported type", "type", v)
	}
	return divideIntbyShardsNumber(bytesRpr, shardSize)
}

func divideIntbyShardsNumber(value uint64, n uint64) uint64 {
	return value % n
}

func getIntegerRepresentationFromString(text string) uint64 {
	h := fnv.New64a()
	for i := 0; i < len(text); i++ {
		cByte := byte(text[i])
		h.Write([]byte{cByte})
	}
	return h.Sum64()
}

// had to get AI help to write this, as the concept is extremely blur on this one
func getIntegerRepresentationFromint(value int) uint64 {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(value))

	h := uint64(14695981039346656037) // offset64
	for i := 0; i < len(buf); i++ {
		h ^= uint64(buf[i])
		h *= 1099511628211 // prime64
	}
	return h
}
