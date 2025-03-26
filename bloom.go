package imagecapture

import (
	"hash/fnv"
	"math"
	"sync"
)

type BloomFilter struct {
	bitset    []bool
	size      uint
	hashFuncs uint
	mutex     sync.RWMutex
}

func NewBloomFilter(expectedItems uint, falsePositiveRate float64) *BloomFilter {
	size := optimalSize(expectedItems, falsePositiveRate)
	hashFuncs := optimalHashFuncs(size, expectedItems)

	return &BloomFilter{
		bitset:    make([]bool, size),
		size:      size,
		hashFuncs: hashFuncs,
	}
}

func (bf *BloomFilter) Add(item string) {
	bf.mutex.Lock()
	defer bf.mutex.Unlock()

	hashes := bf.getHashes(item)
	for i := uint(0); i < bf.hashFuncs; i++ {
		index := hashes[i] % bf.size
		bf.bitset[index] = true
	}
}

func (bf *BloomFilter) Contains(item string) bool {
	bf.mutex.RLock()
	defer bf.mutex.RUnlock()

	hashes := bf.getHashes(item)
	for i := uint(0); i < bf.hashFuncs; i++ {
		index := hashes[i] % bf.size
		if !bf.bitset[index] {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) getHashes(item string) []uint {
	hashes := make([]uint, bf.hashFuncs)
	h := fnv.New64()
	h.Write([]byte(item))
	hash1 := uint(h.Sum64())
	h.Reset()
	h.Write([]byte(item + "imagecapture"))
	hash2 := uint(h.Sum64())

	for i := uint(0); i < bf.hashFuncs; i++ {
		hashes[i] = hash1 + i*hash2
	}
	return hashes
}

func optimalSize(n uint, p float64) uint {
	return uint(math.Ceil(-float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
}

func optimalHashFuncs(size uint, n uint) uint {
	return uint(math.Ceil(float64(size) / float64(n) * math.Log(2)))
}
