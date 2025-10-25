package ray

import (
	rand_v2 "math/rand/v2"
	"sync"
	"testing"
)

// BenchmarkSharedRand benchmarks using a single shared rand.Rand source
// across multiple goroutines (with mutex protection).
func BenchmarkSharedRand(b *testing.B) {
	rng := rand_v2.New(rand_v2.NewPCG(1, 2))
	var mu sync.Mutex

	b.RunParallel(func(pb *testing.PB) {
		var sum float64
		for pb.Next() {
			mu.Lock()
			v := rng.NormFloat64()
			mu.Unlock()
			sum += v
		}
		// Prevent optimization
		_ = sum
	})
}

// BenchmarkPerGoRoutineRand benchmarks using one rand.Rand source per goroutine.
func BenchmarkPerGoRoutineRand(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		rng := rand_v2.New(rand_v2.NewPCG(1, 2))
		var sum float64
		for pb.Next() {
			v := rng.NormFloat64()
			sum += v
		}
		// Prevent optimization
		_ = sum
	})
}

// BenchmarkGlobalRand benchmarks using the global rand functions (for comparison).
// Note: global rand uses ChaCha8, not PCG.
func BenchmarkGlobalRand(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var sum float64
		for pb.Next() {
			v := rand_v2.NormFloat64()
			sum += v
		}
		// Prevent optimization
		_ = sum
	})
}

// BenchmarkPerGoRoutineChaCha8 benchmarks using one ChaCha8 source per goroutine
// for a fair comparison with the global rand (which also uses ChaCha8).
func BenchmarkPerGoRoutineChaCha8(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		rng := rand_v2.New(rand_v2.NewChaCha8([32]byte{1, 2, 3}))
		var sum float64
		for pb.Next() {
			v := rng.NormFloat64()
			sum += v
		}
		// Prevent optimization
		_ = sum
	})
}

// BenchmarkPCGUint64 benchmarks raw PCG Uint64 generation.
func BenchmarkPCGUint64(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		src := rand_v2.NewPCG(1, 2)
		var sum uint64
		for pb.Next() {
			sum += src.Uint64()
		}
		_ = sum
	})
}

// BenchmarkChaCha8Uint64 benchmarks raw ChaCha8 Uint64 generation.
func BenchmarkChaCha8Uint64(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		src := rand_v2.NewChaCha8([32]byte{1, 2, 3})
		var sum uint64
		for pb.Next() {
			sum += src.Uint64()
		}
		_ = sum
	})
}
