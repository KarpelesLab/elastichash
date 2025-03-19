package main

import (
	"fmt"
	"elastichash"
)

func main() {
	fmt.Println("ElasticHash and FunnelHash Example")
	
	// Parameters
	N := 100
	delta := 0.25 // leave 25% of slots empty
	bucketSize := 4
	
	// Create both hash tables
	eht := elastichash.NewElasticHashTable(N, delta)
	fht := elastichash.NewFunnelHashTable(N, bucketSize, delta)
	
	// Insert some keys into both tables
	for i := 0; i < 50; i += 2 {
		eht.Insert(i)
		fht.Insert(i)
	}
	
	// Check membership in ElasticHashTable
	fmt.Println("\nElasticHashTable:")
	fmt.Printf("Size: %d, Capacity: %d\n", eht.Size(), eht.Capacity())
	for i := 0; i < 10; i++ {
		fmt.Printf("Key %d exists: %t\n", i, eht.Contains(i))
	}
	
	// Check membership in FunnelHashTable
	fmt.Println("\nFunnelHashTable:")
	fmt.Printf("Size: %d, Capacity: %d\n", fht.Size(), fht.Capacity())
	for i := 0; i < 10; i++ {
		fmt.Printf("Key %d exists: %t\n", i, fht.Contains(i))
	}
}