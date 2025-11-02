package ray

import "testing"

func TestNewRay(t *testing.T) {
	rnd := RandForTests()
	origin := Vec3{1, 2, 3}
	direction := Vec3{4, 5, 6}

	ray := NewRay(rnd, origin, direction)

	if ray.Origin != origin {
		t.Errorf("Expected origin %v, got %v", origin, ray.Origin)
	}
	if ray.Direction != direction {
		t.Errorf("Expected direction %v, got %v", direction, ray.Direction)
	}
}

func TestRayAt(t *testing.T) {
	rnd := RandForTests()
	origin := Vec3{1, 0, 0}
	direction := Vec3{0, 1, 0}
	ray := NewRay(rnd, origin, direction)

	tests := []struct {
		t        float64
		expected Vec3
	}{
		{0, Vec3{1, 0, 0}},
		{1, Vec3{1, 1, 0}},
		{2, Vec3{1, 2, 0}},
		{-1, Vec3{1, -1, 0}},
		{0.5, Vec3{1, 0.5, 0}},
	}

	for _, tt := range tests {
		result := ray.At(tt.t)
		if result != tt.expected {
			t.Errorf("At(%v): expected %v, got %v", tt.t, tt.expected, result)
		}
	}
}

func TestRayAtGeneral(t *testing.T) {
	rnd := RandForTests()
	origin := Vec3{1, 2, 3}
	direction := Vec3{2, 3, 4}
	ray := NewRay(rnd, origin, direction)

	t2 := 2.5
	result := ray.At(t2)
	expected := Vec3{1 + 2*t2, 2 + 3*t2, 3 + 4*t2}

	if result != expected {
		t.Errorf("At(%v): expected %v, got %v", t2, expected, result)
	}
}
