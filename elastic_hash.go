package elastichash

import (
	"errors"
	"fmt"
)

// We define a special empty marker. We assume non-negative int keys for simplicity.
const EMPTY = -1

type ElasticHashTable struct {
	levels    [][]int  // segments A0 ... A_{L-1}
	L         int      // number of levels
	R         int      // max probes per level (threshold)
	size      int      // current number of elements inserted
	capacity  int      // maximum allowed elements (respecting load factor)
}

// NewElasticHashTable creates a new ElasticHashTable with total array size N and fraction delta of slots left empty.
func NewElasticHashTable(N int, delta float64) *ElasticHashTable {
	if delta < 0 || delta >= 1 {
		panic("delta must be in (0,1)")
	}
	// Determine number of levels L (we use a small constant or derive from log(1/delta)).
	L := 4
	if L < 2 {
		L = 2
	}
	// Maximum elements allowed = floor((1-delta)*N)
	maxElems := int((1 - delta) * float64(N))
	table := &ElasticHashTable{
		levels:   make([][]int, L),
		L:        L,
		R:        L,         // for simplicity, R = L (could be tuned independently)
		size:     0,
		capacity: maxElems,
	}
	// Allocate levels. For simplicity, give first L-1 levels capacity = R (small constant),
	// and last level gets the remainder.
	for i := 0; i < L-1; i++ {
		segSize := table.R  // small segment
		if segSize > N {
			segSize = N
		}
		table.levels[i] = make([]int, segSize)
		for j := range table.levels[i] {
			table.levels[i][j] = EMPTY
		}
		N -= segSize
	}
	// Last level gets all remaining slots (at least 1).
	if N < 1 {
		N = 1
	}
	table.levels[L-1] = make([]int, N)
	for j := range table.levels[L-1] {
		table.levels[L-1][j] = EMPTY
	}
	return table
}

// hashFunc is a deterministic hash generator for (key, level, attempt) -> pseudo-random slot index.
func (ht *ElasticHashTable) hashFunc(key, level, attempt, mod int) int {
	// Combine key, level, attempt into a 64-bit state (this is a simple mix).
	// This is not cryptographic, just to simulate random-like distribution.
	x := uint64(key)
	x ^= (uint64(level) << 33) | uint64(attempt)  // combine bits
	// Mix bits (64-bit XOR-shift multiply)
	x ^= x >> 33
	x *= 0xff51afd7ed558ccd
	x ^= x >> 33
	x *= 0xc4ceb9fe1a85ec53
	x ^= x >> 33
	// Return a non-negative int index
	return int(x % uint64(mod))
}

// Insert adds a key to the hash table. Returns an error if the table is at capacity.
func (ht *ElasticHashTable) Insert(key int) error {
	if ht.size >= ht.capacity {
		return errors.New("hash table is full (max load reached)")
	}
	// Check if already present to avoid duplicates (not strictly required by paper's model, but typical in set)
	if ht.Contains(key) {
		return nil // already in table, nothing to do
	}
	// Try each level in order
	for i := 0; i < ht.L-1; i++ {
		m := len(ht.levels[i])
		// Generate up to R probe positions in Ai
		tried := make(map[int]bool, ht.R)
		for attempt := 0; attempt < ht.R; attempt++ {
			pos := ht.hashFunc(key, i, attempt, m)
			if tried[pos] {
				continue // avoid duplicate probe (rare)
			}
			tried[pos] = true
			if ht.levels[i][pos] == EMPTY { // found an empty slot
				ht.levels[i][pos] = key
				ht.size++
				return nil // inserted successfully
			}
		}
		// If we reach here, all R probes in A_i were occupied – move down to next level
	}
	// Final level (A_{L-1}): do standard open addressing (linear probing for simplicity)
	lastLevel := ht.L - 1
	m := len(ht.levels[lastLevel])
	// Use a double hashing or linear probing. Here, linear probe for simplicity.
	start := ht.hashFunc(key, lastLevel, 0, m)  // starting slot
	for offset := 0; offset < m; offset++ {
		pos := (start + offset) % m
		if ht.levels[lastLevel][pos] == EMPTY {
			ht.levels[lastLevel][pos] = key
			ht.size++
			return nil
		}
	}
	return errors.New("no empty slot found in final level (this should not happen under expected conditions)")
}

// Contains checks if the key is in the table.
func (ht *ElasticHashTable) Contains(key int) bool {
	// Search through the same probe sequence used in insertion.
	for i := 0; i < ht.L-1; i++ {
		m := len(ht.levels[i])
		tried := make(map[int]bool, ht.R)
		for attempt := 0; attempt < ht.R; attempt++ {
			pos := ht.hashFunc(key, i, attempt, m)
			if tried[pos] {
				continue
			}
			tried[pos] = true
			if ht.levels[i][pos] == key {
				return true
			}
			if ht.levels[i][pos] == EMPTY {
				// If we hit an empty slot during search, we can stop looking in this level – 
				// since insertion would have placed the key in the first empty encountered, 
				// not finding it here means it was never in this level.
				break
			}
		}
		// not found in level i; continue to next level
	}
	// Last level: do normal search (linear probe)
	lastLevel := ht.L - 1
	m := len(ht.levels[lastLevel])
	start := ht.hashFunc(key, lastLevel, 0, m)
	for offset := 0; offset < m; offset++ {
		pos := (start + offset) % m
		if ht.levels[lastLevel][pos] == key {
			return true
		}
		if ht.levels[lastLevel][pos] == EMPTY {
			return false
		}
	}
	return false
}

// Size returns the current number of elements in the table.
func (ht *ElasticHashTable) Size() int {
	return ht.size
}

// Capacity returns the maximum number of elements the table can hold.
func (ht *ElasticHashTable) Capacity() int {
	return ht.capacity
}

// String returns a debug representation of the hash table.
func (ht *ElasticHashTable) String() string {
	str := ""
	for i := 0; i < ht.L; i++ {
		str += fmt.Sprintf("Level %d: %v\n", i, ht.levels[i])
	}
	return str
}