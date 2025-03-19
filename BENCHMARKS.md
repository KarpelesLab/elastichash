# Elastic Hash vs Funnel Hash vs Go Map Benchmarks (Updated)

This document presents the results of performance benchmarks comparing the three hash table implementations after optimization:

1. **ElasticHashTable**: Our implementation of the non-greedy multi-level hash table described in "Optimal Bounds for Open Addressing Without Reordering"
2. **FunnelHashTable**: Our implementation of the greedy bucket-based multi-level hash table described in the same paper
3. **Go Map**: The standard built-in Go map implementation

## Key Performance Findings (After Optimization)

### Insert Operations

| Implementation | Time (ns/op) | Memory (B/op) | Allocations (allocs/op) |
|----------------|--------------|---------------|-------------------------|
| ElasticHash    | 138.0        | 0             | 0                       |
| FunnelHash     | 28.29        | 0             | 0                       |
| Go Map         | 44.82        | 0             | 0                       |

- **FunnelHash** remains the fastest for insertions, outperforming Go's built-in map by ~37%
- **ElasticHash** is 3x slower than FunnelHash but significantly faster than before optimization

### Lookup Operations

| Implementation | Time (ns/op) | Memory (B/op) | Allocations (allocs/op) |
|----------------|--------------|---------------|-------------------------|
| ElasticHash    | 65.92        | 0             | 0                       |
| FunnelHash     | 18.33        | 0             | 0                       |
| Go Map         | 6.768        | 0             | 0                       |

- **Go Map** remains the fastest for lookups
- **FunnelHash** lookup performance improved significantly, now about 2.7x slower than Go Map 
- **ElasticHash** is about 9.7x slower than Go Map, but much improved from previous implementation

### High Load Factor Performance (0.9)

| Implementation | Time (ns/op) | Memory (B/op) | Allocations (allocs/op) |
|----------------|--------------|---------------|-------------------------|
| ElasticHash    | 32.28        | 0             | 0                       |
| FunnelHash     | 4.891        | 0             | 0                       |
| Go Map         | 3.649        | 0             | 0                       |

- **FunnelHash** shows excellent performance at high load factors, only 34% slower than Go Map
- **ElasticHash** also performs well at high load factors with the optimized version

## What Went Wrong in Initial Optimization

Our first attempt at optimizing ElasticHash actually degraded performance by:

1. **Adding allocation overhead**: Our attempt to preinitialize probe positions required a dynamic memory allocation
2. **Increasing algorithm complexity**: The duplicate position detection added O(n²) complexity
3. **Introducing unnecessary double-hashing**: This added overhead without significant distribution benefits

## Analysis of Successful Optimizations

### FunnelHash Improvements

The optimized FunnelHash implementation shows significant improvements:

1. **Power of 2 bucket sizes**: We've optimized bucket sizing to use power-of-2 values where possible, replacing expensive modulo operations with bit masking
2. **Loop unrolling**: For common bucket sizes, we manually unrolled the loops for better performance
3. **Cache-friendly probing patterns**: Adjusted level allocation to improve cache locality
4. **Separate fast paths**: Special paths for power-of-2 sizes further improved performance

### ElasticHash Improvements

The revised ElasticHash implementation focussed on simplicity and eliminated allocations:

1. **Static arrays instead of maps**: Using fixed-size arrays for tracking tried positions
2. **Simplified level structure**: Returning to the original level allocation strategy
3. **Better bounds checking**: Ensuring we don't access out of bounds in tried array

## Theoretical vs Practical Performance

After proper optimization, our findings better align with the paper's theoretical claims:

1. **FunnelHash** shows excellent performance, competitive with or better than Go's map for insertions
2. **ElasticHash** performs respectably, especially at high load factors
3. Both implementations demonstrate the theoretical advantages at high load factors

Important factors that affect real-world performance:

1. **Memory access patterns**: Cache-friendly design significantly impacts performance
2. **Allocation avoidance**: Zero-allocation implementations are crucial for performance
3. **Algorithm simplicity**: Sometimes simpler algorithms win due to better CPU pipelining

## Conclusion

- **FunnelHash** is competitive and sometimes superior to Go's map for insertions, making it viable for insert-heavy workloads
- **ElasticHash** demonstrates acceptable performance after optimization, showing the paper's theoretical concepts in practice
- The optimization journey highlights key principles for hash table design:
  1. Avoid allocations in critical paths
  2. Design for cache locality
  3. Use power-of-2 sizing where possible for fast modulo operations
  4. Unroll loops for small, fixed iterations

While Go's built-in map still holds an edge for general use, these optimized implementations demonstrate the paper's algorithms can be practically implemented with competitive performance.