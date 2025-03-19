package elastichash

import (
	"errors"
	"fmt"
)

type FunnelHashTable struct {
	levels    []Level            // slice of levels 0..B-1
	special   []int              // special overflow array
	b         int                // bucket size (slots per bucket)
	size      int
	capacity  int
}

// Each level has an array of buckets. We store as a flat slice and compute bucket indices.
type Level struct {
	slots []int  // length = number of buckets * b
	numBuckets int
}

// NewFunnelHashTable creates a FunnelHashTable with given total size N, bucket size b, and empty fraction delta.
func NewFunnelHashTable(N int, b int, delta float64) *FunnelHashTable {
	if delta < 0 || delta >= 1 {
		panic("delta must be in (0,1)")
	}
	// Determine number of levels B (we choose such that levels geometrically decrease to a small size).
	// For simplicity, let B = 3 or 4, and allocate levels with decreasing sizes.
	B := 3
	if B < 1 {
		B = 1
	}
	// Total allowed elements:
	maxElems := int((1 - delta) * float64(N))
	ht := &FunnelHashTable{
		levels:   make([]Level, B),
		special:  []int{},
		b:        b,
		size:     0,
		capacity: maxElems,
	}
	// Example sizing strategy: level0 = 50% of N, level1 = 30% of N, level2 = 15% of N (sums to 95%, leaving 5% for special).
	// (In practice, sizes can be tuned to optimize δ and probability guarantees)
	sizes := []float64{0.5, 0.3, 0.15}  // for B=3 example
	if B > len(sizes) {
		// If more levels needed, fill uniformly smaller fractions
		frac := 0.1
		for len(sizes) < B {
			sizes = append(sizes, frac)
			remainingFrac := 1.0
			for _, f := range sizes {
				remainingFrac -= f
			}
			if remainingFrac < 0.1 {
				break
			}
		}
	}
	// Allocate levels
	allocated := 0
	for i := 0; i < B; i++ {
		size_i := int(sizes[i] * float64(N))
		if i == B-1 {
			// last level gets whatever is left (excluding special)
			size_i = int(sizes[i] * float64(N))
		}
		if size_i < b {
			size_i = b
		}
		// number of buckets = size_i / b (truncate)
		numB := size_i / b
		levelSlots := make([]int, numB*b)
		for j := range levelSlots {
			levelSlots[j] = EMPTY
		}
		ht.levels[i] = Level{slots: levelSlots, numBuckets: numB}
		allocated += numB * b
	}
	// Special array gets remaining slots
	specialSize := N - allocated
	if specialSize < 1 {
		specialSize = 1
	}
	ht.special = make([]int, specialSize)
	for j := range ht.special {
		ht.special[j] = EMPTY
	}
	return ht
}

// hashFunc for funnel hashing: (key, level) -> bucket index in that level.
func (ht *FunnelHashTable) hashFunc(key int, levelIdx int) int {
	// Simple 32-bit mix for demonstration
	h := uint32(key) * 0x9e3779b1
	h ^= h >> 15
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	level := ht.levels[levelIdx]
	return int(h % uint32(level.numBuckets))
}

// Insert inserts a key into the funnel hash table.
func (ht *FunnelHashTable) Insert(key int) error {
	if ht.size >= ht.capacity {
		return errors.New("hash table is full")
	}
	if ht.Contains(key) {
		return nil  // no duplicates for set semantics
	}
	// Try each level in order
	for i := 0; i < len(ht.levels); i++ {
		lvl := &ht.levels[i]
		bucketIdx := ht.hashFunc(key, i)
		start := bucketIdx * ht.b  // index of first slot in this bucket
		// Probe all slots in this bucket
		emptySlot := -1
		for j := 0; j < ht.b; j++ {
			slotIndex := start + j
			if lvl.slots[slotIndex] == key {
				return nil // already exists (shouldn't happen since we checked Contains)
			}
			if lvl.slots[slotIndex] == EMPTY {
				emptySlot = slotIndex
				break
			}
		}
		if emptySlot != -1 {
			// Found a free slot; insert and stop
			lvl.slots[emptySlot] = key
			ht.size++
			return nil
		}
		// If bucket is full, fall through to next level
	}
	// If all levels failed, insert into special overflow (linear probing)
	m := len(ht.special)
	h0 := ht.hashFunc(key, 0)  // reuse level0 hash as base for special (or use another hash)
	for offset := 0; offset < m; offset++ {
		pos := (h0 + offset) % m
		if ht.special[pos] == EMPTY || ht.special[pos] == key {
			ht.special[pos] = key
			ht.size++
			return nil
		}
	}
	return errors.New("special array is full - insertion failed")
}

// Contains checks if a key exists in the table.
func (ht *FunnelHashTable) Contains(key int) bool {
	// Check each level's corresponding bucket
	for i := 0; i < len(ht.levels); i++ {
		lvl := &ht.levels[i]
		bucketIdx := ht.hashFunc(key, i)
		start := bucketIdx * ht.b
		for j := 0; j < ht.b; j++ {
			slotIndex := start + j
			if lvl.slots[slotIndex] == key {
				return true
			}
			if lvl.slots[slotIndex] == EMPTY {
				// If we find an empty, the key cannot be in this level (it would have been placed in this empty slot if it were here).
				break
			}
		}
		// not found in this level, move to next
	}
	// Check special overflow array
	m := len(ht.special)
	h0 := ht.hashFunc(key, 0)
	for offset := 0; offset < m; offset++ {
		pos := (h0 + offset) % m
		if ht.special[pos] == key {
			return true
		}
		if ht.special[pos] == EMPTY {
			return false
		}
	}
	return false
}

// Size returns the current number of elements in the table.
func (ht *FunnelHashTable) Size() int {
	return ht.size
}

// Capacity returns the maximum number of elements the table can hold.
func (ht *FunnelHashTable) Capacity() int {
	return ht.capacity
}

// String returns a debug representation of the hash table.
func (ht *FunnelHashTable) String() string {
	str := fmt.Sprintf("FunnelHashTable: size=%d, capacity=%d, bucketSize=%d\n", ht.size, ht.capacity, ht.b)
	for i := 0; i < len(ht.levels); i++ {
		lvl := ht.levels[i]
		str += fmt.Sprintf("Level %d (%d buckets): %v\n", i, lvl.numBuckets, lvl.slots)
	}
	str += fmt.Sprintf("Special: %v\n", ht.special)
	return str
}