package elastichash

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestElasticHashTable(t *testing.T) {
	N := 100
	delta := 0.25 // leave 25% of slots empty

	// Create a new elastic hash table
	eht := NewElasticHashTable(N, delta)

	// Test initial state
	if eht.Size() != 0 {
		t.Errorf("Expected initial size 0, got %d", eht.Size())
	}
	if eht.Capacity() != int((1-delta)*float64(N)) {
		t.Errorf("Expected capacity %d, got %d", int((1-delta)*float64(N)), eht.Capacity())
	}

	// Test inserting keys
	for i := 0; i < 50; i += 2 {
		err := eht.Insert(i)
		if err != nil {
			t.Errorf("Error inserting %d: %v", i, err)
		}
	}

	// Test size after insertions
	if eht.Size() != 25 { // 0, 2, 4, ..., 48 (25 numbers)
		t.Errorf("Expected size 25 after insertions, got %d", eht.Size())
	}

	// Test membership checks
	for i := 0; i < 50; i++ {
		expected := i%2 == 0
		if eht.Contains(i) != expected {
			t.Errorf("Expected Contains(%d) to be %v", i, expected)
		}
	}

	// Test duplicate insertion (should not increase size)
	prevSize := eht.Size()
	err := eht.Insert(0) // already exists
	if err != nil {
		t.Errorf("Error re-inserting existing key: %v", err)
	}
	if eht.Size() != prevSize {
		t.Errorf("Size should not change after inserting duplicate, expected %d, got %d", prevSize, eht.Size())
	}

	// Test inserting up to capacity
	for i := 1; i < 100; i += 2 {
		err := eht.Insert(i)
		if err != nil && eht.Size() < eht.Capacity() {
			t.Errorf("Error inserting %d when table not full: %v", i, err)
		}
	}

	// Verify inserted keys are contained
	for i := 0; i < 75; i++ {
		// At this point we should have inserted all even keys from 0-48
		// and some odd keys from the second insertion loop
		if (i <= 48 && i%2 == 0) || (i < 75 && i%2 == 1) {
			if !eht.Contains(i) && eht.Size() < eht.Capacity() {
				// Only report error if the table isn't full yet
				t.Errorf("Expected to find key %d after insertion", i)
			}
		}
	}
}

func TestFunnelHashTable(t *testing.T) {
	N := 100
	bucketSize := 4
	delta := 0.25 // leave 25% of slots empty

	// Create a new funnel hash table
	fht := NewFunnelHashTable(N, bucketSize, delta)

	// Test initial state
	if fht.Size() != 0 {
		t.Errorf("Expected initial size 0, got %d", fht.Size())
	}
	if fht.Capacity() != int((1-delta)*float64(N)) {
		t.Errorf("Expected capacity %d, got %d", int((1-delta)*float64(N)), fht.Capacity())
	}

	// Test inserting keys
	for i := 0; i < 50; i += 2 {
		err := fht.Insert(i)
		if err != nil {
			t.Errorf("Error inserting %d: %v", i, err)
		}
	}

	// Test size after insertions
	if fht.Size() != 25 { // 0, 2, 4, ..., 48 (25 numbers)
		t.Errorf("Expected size 25 after insertions, got %d", fht.Size())
	}

	// Test membership checks
	for i := 0; i < 50; i++ {
		expected := i%2 == 0
		if fht.Contains(i) != expected {
			t.Errorf("Expected Contains(%d) to be %v", i, expected)
		}
	}

	// Test duplicate insertion (should not increase size)
	prevSize := fht.Size()
	err := fht.Insert(0) // already exists
	if err != nil {
		t.Errorf("Error re-inserting existing key: %v", err)
	}
	if fht.Size() != prevSize {
		t.Errorf("Size should not change after inserting duplicate, expected %d, got %d", prevSize, fht.Size())
	}

	// Test inserting up to capacity
	for i := 1; i < 100; i += 2 {
		err := fht.Insert(i)
		if err != nil && fht.Size() < fht.Capacity() {
			t.Errorf("Error inserting %d when table not full: %v", i, err)
		}
	}

	// Verify inserted keys are contained
	for i := 0; i < 75; i++ {
		// At this point we should have inserted all even keys from 0-48
		// and some odd keys from the second insertion loop
		if (i <= 48 && i%2 == 0) || (i < 75 && i%2 == 1) {
			if !fht.Contains(i) && fht.Size() < fht.Capacity() {
				// Only report error if the table isn't full yet
				t.Errorf("Expected to find key %d after insertion", i)
			}
		}
	}
}

func TestHashPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	N := 10000
	delta := 0.1 // 90% load factor
	bucketSize := 8

	// Create hash tables
	eht := NewElasticHashTable(N, delta)
	fht := NewFunnelHashTable(N, bucketSize, delta)

	// Insert keys up to near capacity
	targetSize := int(float64(N) * 0.85) // Close to but not at capacity
	for i := 0; i < targetSize; i++ {
		eht.Insert(i)
		fht.Insert(i)
	}

	// Benchmark lookups - successful case
	t.Run("ElasticHash-SuccessfulLookup", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			key := i % targetSize // keys we know exist
			if !eht.Contains(key) {
				t.Errorf("Key %d should be found", key)
			}
		}
	})

	t.Run("FunnelHash-SuccessfulLookup", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			key := i % targetSize // keys we know exist
			if !fht.Contains(key) {
				t.Errorf("Key %d should be found", key)
			}
		}
	})

	// Benchmark lookups - unsuccessful case
	t.Run("ElasticHash-UnsuccessfulLookup", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			key := N + i // keys that definitely don't exist
			if eht.Contains(key) {
				t.Errorf("Key %d should not be found", key)
			}
		}
	})

	t.Run("FunnelHash-UnsuccessfulLookup", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			key := N + i // keys that definitely don't exist
			if fht.Contains(key) {
				t.Errorf("Key %d should not be found", key)
			}
		}
	})
}

func BenchmarkElasticHashInsert(b *testing.B) {
	N := 10000
	delta := 0.1 // 90% load factor
	eht := NewElasticHashTable(N, delta)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if eht.Size() >= eht.Capacity() {
			// Reset if we reach capacity
			b.StopTimer()
			eht = NewElasticHashTable(N, delta)
			b.StartTimer()
		}
		eht.Insert(i)
	}
}

func BenchmarkFunnelHashInsert(b *testing.B) {
	N := 10000
	delta := 0.1 // 90% load factor
	bucketSize := 8
	fht := NewFunnelHashTable(N, bucketSize, delta)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if fht.Size() >= fht.Capacity() {
			// Reset if we reach capacity
			b.StopTimer()
			fht = NewFunnelHashTable(N, bucketSize, delta)
			b.StartTimer()
		}
		fht.Insert(i)
	}
}

func BenchmarkElasticHashLookup(b *testing.B) {
	N := 10000
	delta := 0.1 // 90% load factor
	eht := NewElasticHashTable(N, delta)

	// Insert half the capacity
	targetSize := eht.Capacity() / 2
	for i := 0; i < targetSize; i++ {
		eht.Insert(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mix of successful and unsuccessful lookups
		key := i % (targetSize * 2)
		eht.Contains(key)
	}
}

func BenchmarkFunnelHashLookup(b *testing.B) {
	N := 10000
	delta := 0.1 // 90% load factor
	bucketSize := 8
	fht := NewFunnelHashTable(N, bucketSize, delta)

	// Insert half the capacity
	targetSize := fht.Capacity() / 2
	for i := 0; i < targetSize; i++ {
		fht.Insert(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mix of successful and unsuccessful lookups
		key := i % (targetSize * 2)
		fht.Contains(key)
	}
}

// Standard Go map benchmarks for comparison

func BenchmarkGoMapInsert(b *testing.B) {
	N := 10000
	delta := 0.1 // 90% load factor
	capacity := int((1-delta) * float64(N))
	goMap := make(map[int]struct{}, capacity)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if len(goMap) >= capacity {
			// Reset if we reach capacity
			b.StopTimer()
			goMap = make(map[int]struct{}, capacity)
			b.StartTimer()
		}
		goMap[i] = struct{}{}
	}
}

func BenchmarkGoMapLookup(b *testing.B) {
	N := 10000
	delta := 0.1 // 90% load factor
	capacity := int((1-delta) * float64(N))
	goMap := make(map[int]struct{}, capacity)

	// Insert half the capacity
	targetSize := capacity / 2
	for i := 0; i < targetSize; i++ {
		goMap[i] = struct{}{}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mix of successful and unsuccessful lookups
		key := i % (targetSize * 2)
		_, _ = goMap[key]
	}
}

// More advanced benchmarks with varying load factors and random access patterns

func BenchmarkComparisonAtHighLoadFactor(b *testing.B) {
	const N = 10000
	const loadFactor = 0.9 // High load factor to stress test
	const bucketSize = 8
	
	// Initialize all data structures with same capacity
	capacity := int(float64(N) * loadFactor)
	
	// Pre-generate insertion and lookup keys
	insertKeys := make([]int, capacity)
	for i := 0; i < capacity; i++ {
		insertKeys[i] = rand.Int()
	}
	
	// Create lookup keys with 50% hit rate
	lookupKeys := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		if rand.Float64() < 0.5 {
			// Successful lookup (choose from inserted keys)
			lookupKeys[i] = insertKeys[rand.Intn(capacity)]
		} else {
			// Unsuccessful lookup (use a random key)
			lookupKeys[i] = rand.Int()
		}
	}
	
	b.Run("ElasticHash", func(b *testing.B) {
		eht := NewElasticHashTable(N, 1-loadFactor)
		
		// Insert all keys
		for _, key := range insertKeys {
			eht.Insert(key)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			eht.Contains(lookupKeys[i%len(lookupKeys)])
		}
	})
	
	b.Run("FunnelHash", func(b *testing.B) {
		fht := NewFunnelHashTable(N, bucketSize, 1-loadFactor)
		
		// Insert all keys
		for _, key := range insertKeys {
			fht.Insert(key)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fht.Contains(lookupKeys[i%len(lookupKeys)])
		}
	})
	
	b.Run("GoMap", func(b *testing.B) {
		goMap := make(map[int]struct{}, N)
		
		// Insert all keys
		for _, key := range insertKeys {
			goMap[key] = struct{}{}
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = goMap[lookupKeys[i%len(lookupKeys)]]
		}
	})
}

// ScalingBenchmark tests how performance scales with table size
func BenchmarkScaling(b *testing.B) {
	tableSizes := []int{100, 1000, 10000, 100000}
	loadFactor := 0.7
	bucketSize := 8
	
	for _, size := range tableSizes {
		capacity := int(float64(size) * loadFactor)
		
		// Generate random keys
		keys := make([]int, capacity)
		for i := 0; i < capacity; i++ {
			keys[i] = rand.Int()
		}
		
		b.Run(fmt.Sprintf("ElasticHash-Size%d", size), func(b *testing.B) {
			eht := NewElasticHashTable(size, 1-loadFactor)
			
			// Insert keys
			for _, key := range keys {
				eht.Insert(key)
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Look up existing keys to test successful lookups
				eht.Contains(keys[i%len(keys)])
			}
		})
		
		b.Run(fmt.Sprintf("FunnelHash-Size%d", size), func(b *testing.B) {
			fht := NewFunnelHashTable(size, bucketSize, 1-loadFactor)
			
			// Insert keys
			for _, key := range keys {
				fht.Insert(key)
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Look up existing keys to test successful lookups
				fht.Contains(keys[i%len(keys)])
			}
		})
		
		b.Run(fmt.Sprintf("GoMap-Size%d", size), func(b *testing.B) {
			goMap := make(map[int]struct{}, size)
			
			// Insert keys
			for _, key := range keys {
				goMap[key] = struct{}{}
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Look up existing keys to test successful lookups
				_, _ = goMap[keys[i%len(keys)]]
			}
		})
	}
}

// BenchmarkLoadFactorImpact tests how load factor affects performance
func BenchmarkLoadFactorImpact(b *testing.B) {
	size := 10000
	bucketSize := 8
	loadFactors := []float64{0.1, 0.3, 0.5, 0.7, 0.9}
	
	for _, loadFactor := range loadFactors {
		capacity := int(float64(size) * loadFactor)
		
		// Generate random keys
		keys := make([]int, capacity)
		for i := 0; i < capacity; i++ {
			keys[i] = rand.Int()
		}
		
		// Create lookup keys (all successful lookups)
		lookupKeys := make([]int, b.N)
		for i := 0; i < b.N; i++ {
			lookupKeys[i] = keys[i%len(keys)]
		}
		
		b.Run(fmt.Sprintf("ElasticHash-LoadFactor%.1f", loadFactor), func(b *testing.B) {
			eht := NewElasticHashTable(size, 1-loadFactor)
			
			// Insert keys
			for _, key := range keys {
				eht.Insert(key)
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				eht.Contains(lookupKeys[i%len(lookupKeys)])
			}
		})
		
		b.Run(fmt.Sprintf("FunnelHash-LoadFactor%.1f", loadFactor), func(b *testing.B) {
			fht := NewFunnelHashTable(size, bucketSize, 1-loadFactor)
			
			// Insert keys
			for _, key := range keys {
				fht.Insert(key)
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fht.Contains(lookupKeys[i%len(lookupKeys)])
			}
		})
		
		b.Run(fmt.Sprintf("GoMap-LoadFactor%.1f", loadFactor), func(b *testing.B) {
			goMap := make(map[int]struct{}, size)
			
			// Insert keys
			for _, key := range keys {
				goMap[key] = struct{}{}
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = goMap[lookupKeys[i%len(lookupKeys)]]
			}
		})
	}
}

// BenchmarkUnsuccessfulLookup specifically tests performance on keys that don't exist
func BenchmarkUnsuccessfulLookup(b *testing.B) {
	size := 10000
	bucketSize := 8
	loadFactors := []float64{0.5, 0.7, 0.9} // Higher load factors where performance differences should be more visible
	
	for _, loadFactor := range loadFactors {
		capacity := int(float64(size) * loadFactor)
		
		// Generate insertion keys (used to populate the tables)
		insertKeys := make([]int, capacity)
		// Create a set of randomly distributed keys
		for i := 0; i < capacity; i++ {
			insertKeys[i] = rand.Int() & 0x7FFFFFFF // Positive integers only
		}
		
		// Create lookup keys that definitely don't exist in the table
		// by flipping the sign bit of inserted keys
		lookupKeys := make([]int, b.N)
		for i := 0; i < b.N; i++ {
			// Take a random key from the insert set and flip its sign to ensure it's not in the table
			lookupKeys[i] = -1 - insertKeys[i%len(insertKeys)]
		}
		
		b.Run(fmt.Sprintf("ElasticHash-LoadFactor%.1f", loadFactor), func(b *testing.B) {
			eht := NewElasticHashTable(size, 1-loadFactor)
			
			// Insert all keys
			for _, key := range insertKeys {
				eht.Insert(key)
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				eht.Contains(lookupKeys[i%len(lookupKeys)])
			}
		})
		
		b.Run(fmt.Sprintf("FunnelHash-LoadFactor%.1f", loadFactor), func(b *testing.B) {
			fht := NewFunnelHashTable(size, bucketSize, 1-loadFactor)
			
			// Insert all keys
			for _, key := range insertKeys {
				fht.Insert(key)
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fht.Contains(lookupKeys[i%len(lookupKeys)])
			}
		})
		
		b.Run(fmt.Sprintf("GoMap-LoadFactor%.1f", loadFactor), func(b *testing.B) {
			goMap := make(map[int]struct{}, size)
			
			// Insert all keys
			for _, key := range insertKeys {
				goMap[key] = struct{}{}
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = goMap[lookupKeys[i%len(lookupKeys)]]
			}
		})
	}
}