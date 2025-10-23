package ray

import (
	"image/color"
	"math"
	"testing"
)

func TestVec3Add(t *testing.T) {
	v := Vec3{1, 2, 3}
	u := XYZ(4, 5, 6)
	result := Add(v, u)
	expected := Vec3{5, 7, 9}
	if result != expected {
		t.Errorf("Add() = %v, want %v", result, expected)
	}
}

func TestVec3Sub(t *testing.T) {
	v := Vec3{5, 7, 9}
	u := Vec3{1, 2, 3}
	result := Sub(v, u)
	expected := Vec3{4, 5, 6}
	if result != expected {
		t.Errorf("Sub() = %v, want %v", result, expected)
	}
}

func TestVec3SMul(t *testing.T) {
	v := Vec3{1, 2, 3}
	result := SMul(v, 2.5)
	expected := Vec3{2.5, 5.0, 7.5}
	if result != expected {
		t.Errorf("SMul() = %v, want %v", result, expected)
	}
}

func TestVec3SDiv(t *testing.T) {
	v := Vec3{10, 20, 30}
	result := SDiv(v, 10)
	expected := Vec3{1, 2, 3}
	if result != expected {
		t.Errorf("SDiv() = %v, want %v", result, expected)
	}
}

func TestVec3Length(t *testing.T) {
	tests := []struct {
		name     string
		v        Vec3
		expected float64
	}{
		{"unit vector", Vec3{1, 0, 0}, 1.0},
		{"3-4-5 triangle", Vec3{3, 4, 0}, 5.0},
		{"zero vector", Vec3{0, 0, 0}, 0.0},
		{"negative values", Vec3{-1, -2, -2}, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Length(tt.v)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Length() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVec3Unit(t *testing.T) {
	v := Vec3{3, 4, 0}
	result := Unit(v)
	expected := Vec3{0.6, 0.8, 0.0}

	for i := range 3 {
		if math.Abs(result[i]-expected[i]) > 1e-9 {
			t.Errorf("Unit()[%d] = %v, want %v", i, result[i], expected[i])
		}
	}

	// Check that unit vector has length 1
	length := Length(result)
	if math.Abs(length-1.0) > 1e-9 {
		t.Errorf("Unit().Length() = %v, want 1.0", length)
	}
}

func TestVec3Accessors(t *testing.T) {
	v := Vec3{1.5, 2.5, 3.5}

	if v.X() != 1.5 {
		t.Errorf("X() = %v, want 1.5", v.X())
	}
	if v.Y() != 2.5 {
		t.Errorf("Y() = %v, want 2.5", v.Y())
	}
	if v.Z() != 3.5 {
		t.Errorf("Z() = %v, want 3.5", v.Z())
	}
}

func TestColorF(t *testing.T) {
	c := ColorF{0.5, 0.75, 1.0}
	if c[0] != 0.5 || c[1] != 0.75 || c[2] != 1.0 {
		t.Errorf("ColorF() = %v, want [0.5 0.75 1.0]", c)
	}
}

func TestFloatColorToRGBA(t *testing.T) {
	tests := []struct {
		name     string
		c        ColorF
		expected color.RGBA
	}{
		{"black", ColorF{0, 0, 0}, color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		{"white", ColorF{1, 1, 1}, color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"red", ColorF{1, 0, 0}, color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"green", ColorF{0, 1, 0}, color.RGBA{R: 0, G: 255, B: 0, A: 255}},
		{"blue", ColorF{0, 0, 1}, color.RGBA{R: 0, G: 0, B: 255, A: 255}},
		{"mid gray", ColorF{0.5, 0.5, 0.5}, color.RGBA{R: 127, G: 127, B: 127, A: 255}},
		{"clamped above", ColorF{1.5, 2.0, 3.0}, color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"clamped below", ColorF{-1.0, -0.5, -2.0}, color.RGBA{R: 0, G: 0, B: 0, A: 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.c.ToRGBA()
			if result != tt.expected {
				t.Errorf("ToRGBA() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		x        float64
		min      float64
		max      float64
		expected float64
	}{
		{"within range", 0.5, 0.0, 1.0, 0.5},
		{"below min", -0.5, 0.0, 1.0, 0.0},
		{"above max", 1.5, 0.0, 1.0, 1.0},
		{"at min", 0.0, 0.0, 1.0, 0.0},
		{"at max", 1.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clamp(tt.x, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("clamp(%v, %v, %v) = %v, want %v", tt.x, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}
