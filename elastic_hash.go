package elastichash

import (
	"errors"
	"fmt"
	"sync/atomic"
)

// We define special markers. We assume non-negative int keys for simplicity.
const (
	EMPTY     = -1 // Slot has never been used
	TOMBSTONE = -2 // Slot was used but now deleted
)

type ElasticHashTable struct {
	levels    [][]int  // segments A0 ... A_{L-1}
	L         int      // number of levels
	R         int      // max probes per level (threshold)
	size      int32    // current number of elements inserted (atomic)
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
// This implementation uses SplitMix64 algorithm for fast high-quality hashing
func (ht *ElasticHashTable) hashFunc(key, level, attempt, mod int) int {
	// Combine key, level, attempt into a 64-bit state
	x := uint64(key)
	x ^= (uint64(level) << 33) | uint64(attempt)
	
	// SplitMix64 mixing - extremely fast and high quality bit mixing
	x += 0x9E3779B97F4A7C15  // Golden ratio constant
	x = (x ^ (x >> 30)) * 0xBF58476D1CE4E5B9
	x = (x ^ (x >> 27)) * 0x94D049BB133111EB
	x = x ^ (x >> 31)
	
	// Return a non-negative int index
	return int(x % uint64(mod))
}

// Insert adds a key to the hash table. Returns an error if the table is at capacity.
func (ht *ElasticHashTable) Insert(key int) error {
	if atomic.LoadInt32(&ht.size) >= int32(ht.capacity) {
		return errors.New("hash table is full (max load reached)")
	}
	
	// Check if key already exists in any level
	if ht.Contains(key) {
		return nil // already in table, nothing to do
	}
	
	// Try each level in order
	for i := 0; i < ht.L-1; i++ {
		m := len(ht.levels[i])
		// Generate up to R probe positions in Ai
		// Use a fixed-size array instead of map for tracking tried positions
		var tried [16]bool // Assuming R <= 16; adjust size if needed
		for attempt := 0; attempt < ht.R; attempt++ {
			pos := ht.hashFunc(key, i, attempt, m)
			if pos < len(tried) && tried[pos] {
				continue // avoid duplicate probe (rare)
			}
			if pos < len(tried) {
				tried[pos] = true
			}
			
			// Found an empty or deleted slot
			if ht.levels[i][pos] == EMPTY || ht.levels[i][pos] == TOMBSTONE {
				ht.levels[i][pos] = key
				atomic.AddInt32(&ht.size, 1)
				return nil // inserted successfully
			}
		}
		// If we reach here, all R probes in A_i were occupied – move down to next level
	}
	
	// Final level (A_{L-1}): optimize for power of 2 sizes when possible
	lastLevel := ht.L - 1
	m := len(ht.levels[lastLevel])
	
	// Check if m is a power of 2 for fast modulo with bitwise AND
	isPowerOfTwo := (m & (m - 1)) == 0
	
	// Starting slot
	start := ht.hashFunc(key, lastLevel, 0, m)
	
	if isPowerOfTwo {
		// Fast path with bitwise AND for modulo
		mask := m - 1
		for offset := 0; offset < m; offset++ {
			pos := (start + offset) & mask
			if ht.levels[lastLevel][pos] == EMPTY || ht.levels[lastLevel][pos] == TOMBSTONE {
				ht.levels[lastLevel][pos] = key
				atomic.AddInt32(&ht.size, 1)
				return nil
			}
		}
	} else {
		// Standard path with modulo
		for offset := 0; offset < m; offset++ {
			pos := (start + offset) % m
			if ht.levels[lastLevel][pos] == EMPTY || ht.levels[lastLevel][pos] == TOMBSTONE {
				ht.levels[lastLevel][pos] = key
				atomic.AddInt32(&ht.size, 1)
				return nil
			}
		}
	}
	
	return errors.New("no empty slot found in final level (this should not happen under expected conditions)")
}

// Contains checks if the key is in the table.
func (ht *ElasticHashTable) Contains(key int) bool {
	// Search through the same probe sequence used in insertion.
	for i := 0; i < ht.L-1; i++ {
		m := len(ht.levels[i])
		// Use a fixed-size array instead of map for tracking tried positions
		var tried [16]bool // Assuming R <= 16; adjust size if needed
		
		// Unrolled loop for first few attempts for better performance
		if ht.R >= 1 {
			pos := ht.hashFunc(key, i, 0, m)
			if pos < len(tried) {
				tried[pos] = true
				if ht.levels[i][pos] == key {
					return true
				}
				if ht.levels[i][pos] == EMPTY {
					goto nextLevel
				}
				// Tombstones require us to continue searching (unlike empty slots)
			}
		}
		
		if ht.R >= 2 {
			pos := ht.hashFunc(key, i, 1, m)
			if pos < len(tried) && !tried[pos] {
				tried[pos] = true
				if ht.levels[i][pos] == key {
					return true
				}
				if ht.levels[i][pos] == EMPTY {
					goto nextLevel
				}
			}
		}
		
		if ht.R >= 3 {
			pos := ht.hashFunc(key, i, 2, m)
			if pos < len(tried) && !tried[pos] {
				tried[pos] = true
				if ht.levels[i][pos] == key {
					return true
				}
				if ht.levels[i][pos] == EMPTY {
					goto nextLevel
				}
			}
		}
		
		if ht.R >= 4 {
			pos := ht.hashFunc(key, i, 3, m)
			if pos < len(tried) && !tried[pos] {
				tried[pos] = true
				if ht.levels[i][pos] == key {
					return true
				}
				if ht.levels[i][pos] == EMPTY {
					goto nextLevel
				}
			}
		}
		
		// Check remaining attempts
		for attempt := 4; attempt < ht.R; attempt++ {
			pos := ht.hashFunc(key, i, attempt, m)
			if pos < len(tried) && tried[pos] {
				continue
			}
			if pos < len(tried) {
				tried[pos] = true
			}
			
			if ht.levels[i][pos] == key {
				return true
			}
			if ht.levels[i][pos] == EMPTY {
				// If we hit an empty slot during search, we can stop looking in this level – 
				// since insertion would have placed the key in the first empty encountered, 
				// not finding it here means it was never in this level.
				goto nextLevel
			}
			// Tombstones require us to continue searching
		}
		
	nextLevel:
		// not found in level i; continue to next level
	}
	
	// Last level: optimize for power of 2 sizes
	lastLevel := ht.L - 1
	m := len(ht.levels[lastLevel])
	isPowerOfTwo := (m & (m - 1)) == 0
	start := ht.hashFunc(key, lastLevel, 0, m)
	
	if isPowerOfTwo {
		// Fast path with bitwise AND
		mask := m - 1
		for offset := 0; offset < m; offset++ {
			pos := (start + offset) & mask
			if ht.levels[lastLevel][pos] == key {
				return true
			}
			if ht.levels[lastLevel][pos] == EMPTY {
				return false
			}
			// Continue on tombstones
		}
	} else {
		// Standard path
		for offset := 0; offset < m; offset++ {
			pos := (start + offset) % m
			if ht.levels[lastLevel][pos] == key {
				return true
			}
			if ht.levels[lastLevel][pos] == EMPTY {
				return false
			}
			// Continue on tombstones
		}
	}
	
	return false
}

// Remove deletes a key from the hash table if it exists.
// Returns true if the key was found and removed, false otherwise.
func (ht *ElasticHashTable) Remove(key int) bool {
	// Search through the same probe sequence used in insertion and Contains.
	for i := 0; i < ht.L-1; i++ {
		m := len(ht.levels[i])
		var tried [16]bool // Assuming R <= 16
		
		for attempt := 0; attempt < ht.R; attempt++ {
			pos := ht.hashFunc(key, i, attempt, m)
			if pos < len(tried) && tried[pos] {
				continue
			}
			if pos < len(tried) {
				tried[pos] = true
			}
			
			if ht.levels[i][pos] == key {
				// Found the key - mark as deleted
				ht.levels[i][pos] = TOMBSTONE
				atomic.AddInt32(&ht.size, -1)
				return true
			}
			if ht.levels[i][pos] == EMPTY {
				break // Not in this level
			}
		}
		// Not found in this level, continue to next level
	}
	
	// Last level
	lastLevel := ht.L - 1
	m := len(ht.levels[lastLevel])
	isPowerOfTwo := (m & (m - 1)) == 0
	start := ht.hashFunc(key, lastLevel, 0, m)
	
	if isPowerOfTwo {
		mask := m - 1
		for offset := 0; offset < m; offset++ {
			pos := (start + offset) & mask
			if ht.levels[lastLevel][pos] == key {
				ht.levels[lastLevel][pos] = TOMBSTONE
				atomic.AddInt32(&ht.size, -1)
				return true
			}
			if ht.levels[lastLevel][pos] == EMPTY {
				return false
			}
		}
	} else {
		for offset := 0; offset < m; offset++ {
			pos := (start + offset) % m
			if ht.levels[lastLevel][pos] == key {
				ht.levels[lastLevel][pos] = TOMBSTONE
				atomic.AddInt32(&ht.size, -1)
				return true
			}
			if ht.levels[lastLevel][pos] == EMPTY {
				return false
			}
		}
	}
	
	return false
}

// Size returns the current number of elements in the table.
func (ht *ElasticHashTable) Size() int {
	return int(atomic.LoadInt32(&ht.size))
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