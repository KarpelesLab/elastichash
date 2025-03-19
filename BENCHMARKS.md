# Elastic Hash vs Funnel Hash vs Go Map Benchmarks

This document presents the results of performance benchmarks comparing the three hash table implementations:

1. **ElasticHashTable**: Our implementation of the non-greedy multi-level hash table described in "Optimal Bounds for Open Addressing Without Reordering"
2. **FunnelHashTable**: Our implementation of the greedy bucket-based multi-level hash table described in the same paper
3. **Go Map**: The standard built-in Go map implementation

## Key Performance Findings

### Insert Operations

| Implementation | Time (ns/op) | Memory (B/op) | Allocations (allocs/op) |
|----------------|--------------|---------------|-------------------------|
| ElasticHash    | 354.2        | 0             | 0                       |
| FunnelHash     | 39.30        | 0             | 0                       |
| Go Map         | 45.05        | 0             | 0                       |

- **FunnelHash** is the fastest for insertions, slightly outperforming Go's built-in map
- **ElasticHash** is significantly slower for insertions, likely due to its non-greedy insertion strategy that requires examining multiple levels

### Lookup Operations (Successful)

| Implementation | Time (ns/op) | Memory (B/op) | Allocations (allocs/op) |
|----------------|--------------|---------------|-------------------------|
| ElasticHash    | 168.0        | 0             | 0                       |
| FunnelHash     | 20.04        | 0             | 0                       |
| Go Map         | 6.656        | 0             | 0                       |

- **Go Map** is the fastest for successful lookups
- **FunnelHash** is about 3x slower than Go Map
- **ElasticHash** is significantly slower, about 25x slower than Go Map

### Unsuccessful Lookup Operations (High Load Factor = 0.9)

| Implementation | Time (ns/op) | Memory (B/op) | Allocations (allocs/op) |
|----------------|--------------|---------------|-------------------------|
| ElasticHash    | 97.63        | 0             | 0                       |
| FunnelHash     | 17.61        | 0             | 0                       |
| Go Map         | 5.191        | 0             | 0                       |

- Surprisingly, **Go Map** remains the fastest even for unsuccessful lookups at high load factors
- **FunnelHash** remains about 3-4x slower than Go Map
- **ElasticHash** performs worse on unsuccessful lookups

### Performance Across Different Load Factors

| Implementation | Load Factor 0.1 | Load Factor 0.5 | Load Factor 0.9 |
|----------------|----------------|----------------|----------------|
| ElasticHash    | 11.97 ns/op    | 11.99 ns/op    | 11.91 ns/op    |
| FunnelHash     | 2.870 ns/op    | 2.848 ns/op    | 2.885 ns/op    |
| Go Map         | 3.060 ns/op    | 3.044 ns/op    | 3.076 ns/op    |

- Performance for successful lookups is remarkably consistent across different load factors for all implementations
- **FunnelHash** slightly outperforms **Go Map** in the successful lookup benchmark with pre-generated lookup keys

### Scaling with Table Size

| Implementation | Size 100    | Size 1,000   | Size 10,000  | Size 100,000  |
|----------------|-------------|-------------|-------------|--------------|
| ElasticHash    | 104.0 ns/op | 157.3 ns/op | 169.3 ns/op | 169.4 ns/op  |
| FunnelHash     | 6.203 ns/op | 5.894 ns/op | 15.36 ns/op | 19.30 ns/op  |
| Go Map         | 3.454 ns/op | 3.410 ns/op | 11.92 ns/op | 15.22 ns/op  |

- **Go Map** scales the best with increasing table size
- **ElasticHash** shows the worst scaling behavior
- **FunnelHash** remains competitive with Go Map even at larger sizes

## Analysis

### Go Map's Superior Performance

The standard Go map outperforms our custom implementations in most benchmarks, particularly for lookup operations. This is likely due to:

1. Highly optimized implementation in the Go runtime
2. Hardware-friendly memory layout
3. Sophisticated probing strategy that's been refined over years

### FunnelHash vs ElasticHash

FunnelHash consistently outperforms ElasticHash in all benchmarks. This is likely due to:

1. The bucket-based approach reducing the number of probe attempts
2. More cache-friendly memory access patterns
3. Simplified level structure compared to ElasticHash

### Theoretical vs Practical Performance

While the paper "Optimal Bounds for Open Addressing Without Reordering" proves that these hash table designs have optimal theoretical bounds, our implementation does not demonstrate practical performance benefits over the built-in Go map implementation:

1. The asymptotic advantages may only become apparent with extremely large tables or pathological inputs
2. Language-specific optimizations in Go's built-in map overcome theoretical disadvantages
3. Modern CPU architecture (caching, branch prediction, etc.) favors the simpler access patterns of Go's built-in map

## Conclusion

- **For practical use, Go's built-in map remains the best choice** for most applications
- **FunnelHash** shows competitive performance and could be useful in specific scenarios where its theoretical guarantees are important
- **ElasticHash** provides interesting theoretical guarantees but has poor practical performance in our implementation

The academic value of these implementations is in demonstrating the concepts from the paper, but further optimization would be needed to make them competitive with Go's built-in map for general-purpose use.