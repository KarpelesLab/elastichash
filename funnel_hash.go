package elastichash

import (
	"errors"
	"fmt"
)

type FunnelHashTable struct {
	levels    []Level   // slice of levels 0..B-1
	special   []int     // special overflow array
	b         int       // bucket size (slots per bucket)
	size      int
	capacity  int
}

// Each level has an array of buckets. We store as a flat slice and compute bucket indices.
type Level struct {
	slots      []int  // length = number of buckets * b
	numBuckets int
	mask       uint32 // bit mask for fast modulo (power of 2 optimization)
}

// NewFunnelHashTable creates a FunnelHashTable with given total size N, bucket size b, and empty fraction delta.
func NewFunnelHashTable(N int, b int, delta float64) *FunnelHashTable {
	if delta < 0 || delta >= 1 {
		panic("delta must be in (0,1)")
	}
	
	// Determine number of levels B, with optimized distribution
	B := 3
	if delta < 0.1 {
		// For very low delta, use more levels
		B = 4
	}
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
	
	// Revised sizing strategy based on paper analysis
	// Designed for better load distribution
	sizes := []float64{0.6, 0.25, 0.1}  // default for B=3
	if B == 4 {
		sizes = []float64{0.5, 0.25, 0.15, 0.05} // for B=4
	}
	
	if B > len(sizes) {
		// If more levels needed, fill uniformly smaller fractions
		frac := 0.1
		for len(sizes) < B {
			sizes = append(sizes, frac)
			remainingFrac := 1.0
			for _, f := range sizes {
				remainingFrac -= f
			}
			if remainingFrac < 0.05 {
				break
			}
		}
	}
	
	// Allocate levels, try to use power of 2 sizes for faster modulo operation
	allocated := 0
	for i := 0; i < B; i++ {
		size_i := int(sizes[i] * float64(N))
		if i == B-1 {
			// Ensure last level has enough space
			size_i = int(sizes[i] * float64(N))
		}
		
		// Ensure minimum bucket size
		if size_i < b {
			size_i = b
		}
		
		// Number of buckets = size_i / b (truncate)
		numB := size_i / b
		
		// Try to round to power of 2 for faster modulo operation
		powerOf2 := 1
		for powerOf2 < numB {
			powerOf2 <<= 1
		}
		
		// Use power of 2 if it doesn't increase size too much
		if powerOf2 <= numB*5/4 {
			numB = powerOf2
		}
		
		// Compute mask for fast modulo if numB is power of 2
		var mask uint32 = 0
		if numB > 0 && (numB & (numB-1)) == 0 {
			mask = uint32(numB - 1)
		}
		
		levelSlots := make([]int, numB*b)
		for j := range levelSlots {
			levelSlots[j] = EMPTY
		}
		
		ht.levels[i] = Level{
			slots:      levelSlots, 
			numBuckets: numB,
			mask:       mask,
		}
		allocated += numB * b
	}
	
	// Special array gets remaining slots
	specialSize := N - allocated
	if specialSize < 1 {
		specialSize = 1
	}
	
	// Round special array to power of 2 for better performance if reasonable
	powerOf2 := 1
	for powerOf2 < specialSize {
		powerOf2 <<= 1
	}
	if powerOf2 <= specialSize*5/4 {
		specialSize = powerOf2
	}
	
	ht.special = make([]int, specialSize)
	for j := range ht.special {
		ht.special[j] = EMPTY
	}
	return ht
}

// hashFunc for funnel hashing: (key, level) -> bucket index in that level.
// Uses fast modulo if level's numBuckets is a power of 2
func (ht *FunnelHashTable) hashFunc(key int, levelIdx int) int {
	// Simple 32-bit mix for demonstration
	h := uint32(key) * 0x9e3779b1
	h ^= h >> 15
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	
	level := ht.levels[levelIdx]
	
	// Use bit masking for fast modulo if numBuckets is power of 2
	if level.mask > 0 {
		return int(h & level.mask)
	}
	
	return int(h % uint32(level.numBuckets))
}

// Insert inserts a key into the funnel hash table.
func (ht *FunnelHashTable) Insert(key int) error {
	if ht.size >= ht.capacity {
		return errors.New("hash table is full")
	}
	
	// Try each level in order
	for i := 0; i < len(ht.levels); i++ {
		lvl := &ht.levels[i]
		bucketIdx := ht.hashFunc(key, i)
		start := bucketIdx * ht.b  // index of first slot in this bucket
		
		// First check if key already exists in this bucket
		for j := 0; j < ht.b; j++ {
			slotIndex := start + j
			if lvl.slots[slotIndex] == key {
				return nil // already exists
			}
		}
		
		// Now look for an empty slot
		for j := 0; j < ht.b; j++ {
			slotIndex := start + j
			if lvl.slots[slotIndex] == EMPTY {
				lvl.slots[slotIndex] = key
				ht.size++
				return nil
			}
		}
		// If bucket is full, fall through to next level
	}
	
	// If all levels failed, insert into special overflow
	// Optimize special array for power of 2 size if possible
	m := len(ht.special)
	h0 := uint32(key) * 0x9e3779b1  // different hash for special array
	
	// Fast path if m is power of 2
	if m > 0 && (m & (m-1)) == 0 {
		mask := uint32(m - 1)
		start := h0 & mask
		
		// First check if key already exists
		for offset := uint32(0); offset < uint32(m); offset++ {
			pos := int((start + offset) & mask)
			if ht.special[pos] == key {
				return nil
			}
			if ht.special[pos] == EMPTY {
				ht.special[pos] = key
				ht.size++
				return nil
			}
		}
	} else {
		// Standard linear probing for non-power-of-2 sizes
		start := int(h0 % uint32(m))
		for offset := 0; offset < m; offset++ {
			pos := (start + offset) % m
			if ht.special[pos] == key {
				return nil
			}
			if ht.special[pos] == EMPTY {
				ht.special[pos] = key
				ht.size++
				return nil
			}
		}
	}
	
	return errors.New("special array is full - insertion failed")
}

// Contains checks if a key exists in the table.
func (ht *FunnelHashTable) Contains(key int) bool {
	// Use local variables to avoid repeated field accesses
	b := ht.b
	
	// Check each level's corresponding bucket
	for i := 0; i < len(ht.levels); i++ {
		lvl := &ht.levels[i]
		bucketIdx := ht.hashFunc(key, i)
		start := bucketIdx * b
		
		// Optimized unrolled version for common bucket sizes
		switch {
		case b >= 8:
			// Unroll first 8 slots
			if lvl.slots[start] == key {
				return true
			}
			if lvl.slots[start] == EMPTY {
				goto nextLevel
			}
			
			if lvl.slots[start+1] == key {
				return true
			}
			if lvl.slots[start+1] == EMPTY {
				goto nextLevel
			}
			
			if lvl.slots[start+2] == key {
				return true
			}
			if lvl.slots[start+2] == EMPTY {
				goto nextLevel
			}
			
			if lvl.slots[start+3] == key {
				return true
			}
			if lvl.slots[start+3] == EMPTY {
				goto nextLevel
			}
			
			if lvl.slots[start+4] == key {
				return true
			}
			if lvl.slots[start+4] == EMPTY {
				goto nextLevel
			}
			
			if lvl.slots[start+5] == key {
				return true
			}
			if lvl.slots[start+5] == EMPTY {
				goto nextLevel
			}
			
			if lvl.slots[start+6] == key {
				return true
			}
			if lvl.slots[start+6] == EMPTY {
				goto nextLevel
			}
			
			if lvl.slots[start+7] == key {
				return true
			}
			if lvl.slots[start+7] == EMPTY {
				goto nextLevel
			}
			
			// Check remaining slots if bucket size > 8
			for j := 8; j < b; j++ {
				slotIndex := start + j
				if lvl.slots[slotIndex] == key {
					return true
				}
				if lvl.slots[slotIndex] == EMPTY {
					goto nextLevel
				}
			}
			
		case b >= 4:
			// Unroll 4 slots for medium buckets
			if lvl.slots[start] == key {
				return true
			}
			if lvl.slots[start] == EMPTY {
				goto nextLevel
			}
			
			if lvl.slots[start+1] == key {
				return true
			}
			if lvl.slots[start+1] == EMPTY {
				goto nextLevel
			}
			
			if lvl.slots[start+2] == key {
				return true
			}
			if lvl.slots[start+2] == EMPTY {
				goto nextLevel
			}
			
			if lvl.slots[start+3] == key {
				return true
			}
			if lvl.slots[start+3] == EMPTY {
				goto nextLevel
			}
			
			// Check remaining slots if bucket size > 4
			for j := 4; j < b; j++ {
				slotIndex := start + j
				if lvl.slots[slotIndex] == key {
					return true
				}
				if lvl.slots[slotIndex] == EMPTY {
					goto nextLevel
				}
			}
			
		default:
			// Standard loop for small buckets
			for j := 0; j < b; j++ {
				slotIndex := start + j
				if lvl.slots[slotIndex] == key {
					return true
				}
				if lvl.slots[slotIndex] == EMPTY {
					goto nextLevel
				}
			}
		}
		
	nextLevel:
		// Continue to next level
	}
	
	// Check special overflow array
	m := len(ht.special)
	h0 := uint32(key) * 0x9e3779b1  // Different hash for special array
	
	// Fast path if m is power of 2
	if m > 0 && (m & (m-1)) == 0 {
		mask := uint32(m - 1)
		start := h0 & mask
		
		for offset := uint32(0); offset < uint32(m); offset++ {
			pos := int((start + offset) & mask)
			if ht.special[pos] == key {
				return true
			}
			if ht.special[pos] == EMPTY {
				return false
			}
		}
	} else {
		// Standard linear probing for non-power-of-2 sizes
		start := int(h0 % uint32(m))
		for offset := 0; offset < m; offset++ {
			pos := (start + offset) % m
			if ht.special[pos] == key {
				return true
			}
			if ht.special[pos] == EMPTY {
				return false
			}
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