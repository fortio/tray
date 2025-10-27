package ray

import "testing"

// Non-generic versions for comparison.
func AddDirect(u, v Vec3) Vec3 {
	return Vec3{v.x + u.x, v.y + u.y, v.z + u.z}
}

func (v Vec3) AddMethod(u Vec3) Vec3 {
	return Vec3{v.x + u.x, v.y + u.y, v.z + u.z}
}

func SMulDirect(v Vec3, t float64) Vec3 {
	return Vec3{v.x * t, v.y * t, v.z * t}
}

// Benchmarks.
func BenchmarkAddGeneric(b *testing.B) {
	v1 := Vec3{1.0, 2.0, 3.0}
	v2 := Vec3{4.0, 5.0, 6.0}
	var result Vec3
	for b.Loop() {
		result = Add(v1, v2)
	}
	_ = result
}

func BenchmarkAddDirect(b *testing.B) {
	v1 := Vec3{1.0, 2.0, 3.0}
	v2 := Vec3{4.0, 5.0, 6.0}
	var result Vec3
	for b.Loop() {
		result = AddDirect(v1, v2)
	}
	_ = result
}

func BenchmarkAddMethod(b *testing.B) {
	v1 := Vec3{1.0, 2.0, 3.0}
	v2 := Vec3{4.0, 5.0, 6.0}
	var result Vec3
	for b.Loop() {
		result = v1.AddMethod(v2)
	}
	_ = result
}

// (Somewhat) More realistic: chain of operations.
func BenchmarkChainedGeneric(b *testing.B) {
	v1 := Vec3{1.0, 2.0, 3.0}
	v2 := Vec3{4.0, 5.0, 6.0}
	v3 := Vec3{7.0, 8.0, 9.0}
	var result Vec3
	for b.Loop() {
		result = Add(Add(v1, v2), SMul(v3, 2.0))
	}
	_ = result
}

func BenchmarkChainedDirect(b *testing.B) {
	v1 := Vec3{1.0, 2.0, 3.0}
	v2 := Vec3{4.0, 5.0, 6.0}
	v3 := Vec3{7.0, 8.0, 9.0}
	var result Vec3
	for b.Loop() {
		result = AddDirect(AddDirect(v1, v2), SMulDirect(v3, 2.0))
	}
	_ = result
}
