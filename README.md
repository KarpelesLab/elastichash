# ElasticHash

This project provides a Go implementation of two novel open-addressing hash table algorithms described in the paper [Optimal Bounds for Open Addressing Without Reordering](https://arxiv.org/abs/2501.02305) by Martin Farach-Colton, Shalev Ben-David, and Meng-Tsung Tsai.

## About the Algorithms

The paper introduces two innovative hash table strategies that achieve better expected probe complexities than classical methods without moving elements after insertion (no reordering):

1. **Elastic Hashing** (Non-greedy): Uses a multi-level table and a two-dimensional probe sequence, achieving O(1) amortized expected search cost and O(log(1/δ)) worst-case expected search cost (for a table load factor of 1-δ).

2. **Funnel Hashing** (Greedy): A simpler greedy strategy that partitions the hash table into geometrically decreasing levels with fixed-size buckets, achieving O((log(1/δ))²) worst-case expected probes.

Both methods cleverly structure the table into multiple segments to break the traditional coupon-collector bottleneck in hash table probing.

## Implementation

This repository contains:

- `elastic_hash.go`: Implementation of Elastic Hashing
- `funnel_hash.go`: Implementation of Funnel Hashing
- `hash_test.go`: Tests and benchmarks for both implementations
- `example/example.go`: Example usage of both hash tables

## Usage

```go
import "elastichash"

// Create an Elastic Hash Table
N := 100         // Total size of the table
delta := 0.25    // Fraction of slots to leave empty
eht := elastichash.NewElasticHashTable(N, delta)

// Insert keys
eht.Insert(42)

// Check if key exists
exists := eht.Contains(42)

// Create a Funnel Hash Table
bucketSize := 4  // Number of slots per bucket
fht := elastichash.NewFunnelHashTable(N, bucketSize, delta)

// Insert and lookup operations are the same
fht.Insert(42)
exists = fht.Contains(42)
```

## Performance

Both hash tables are designed to offer better theoretical guarantees than traditional open addressing at high load factors. In general:

- Elastic Hashing provides O(1) amortized expected search cost
- Funnel Hashing is simpler and may have better practical performance in some cases

Run the benchmarks to compare their performance on your machine:

```
go test -bench=.
```

See [BENCHMARKS.md](BENCHMARKS.md) for detailed performance comparisons between the ElasticHash, FunnelHash, and Go's built-in map implementation. Our benchmark results show that:

1. FunnelHash performs competitively with Go's built-in map for insertions
2. Go's built-in map generally outperforms both custom implementations for lookups
3. Performance characteristics vary based on load factor and table size

## Paper Abstract

> Farach-Colton et al. paper "Optimal Bounds for Open Addressing Without Reordering" introduces two novel open-address hash table strategies that achieve much better expected probe complexities than classical methods. Both methods avoid reordering (once an item is placed, it never moves), yet they cleverly structure the table into multiple segments (or "levels") to break the traditional coupon-collector bottleneck in hash table probing.

## License

This implementation is provided under the MIT License. Copyright (c) 2025 Karpeles Lab Inc.

## Reference

Farach-Colton, M., Ben-David, S., & Tsai, M. (2024). Optimal Bounds for Open Addressing Without Reordering. arXiv:2501.02305. [https://arxiv.org/abs/2501.02305](https://arxiv.org/abs/2501.02305)