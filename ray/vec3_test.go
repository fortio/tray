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
	u = Neg(u)
	result2 := Add(v, u)
	if result2 != expected {
		t.Errorf("Add with Neg() = %v, want %v", result2, expected)
	}
}

func TestAddMultiple(t *testing.T) {
	tests := []struct {
		name     string
		u        Vec3
		vs       []Vec3
		expected Vec3
	}{
		{"single vector", Vec3{1, 2, 3}, []Vec3{{4, 5, 6}}, Vec3{5, 7, 9}},
		{"two vectors", Vec3{1, 2, 3}, []Vec3{{4, 5, 6}, {7, 8, 9}}, Vec3{12, 15, 18}},
		{"three vectors", Vec3{1, 1, 1}, []Vec3{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}, Vec3{2, 2, 2}},
		{"no additional vectors", Vec3{5, 10, 15}, []Vec3{}, Vec3{5, 10, 15}},
		{"with negatives", Vec3{10, 10, 10}, []Vec3{{-5, 0, 5}, {3, -3, 0}}, Vec3{8, 7, 15}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddMultiple(tt.u, tt.vs...)
			if result != tt.expected {
				t.Errorf("AddMultiple() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSubMultiple(t *testing.T) {
	tests := []struct {
		name     string
		u        Vec3
		v0       Vec3
		vs       []Vec3
		expected Vec3
	}{
		{"single subtraction", Vec3{10, 10, 10}, Vec3{1, 2, 3}, []Vec3{}, Vec3{9, 8, 7}},
		{"two subtractions", Vec3{10, 10, 10}, Vec3{1, 2, 3}, []Vec3{{2, 3, 4}}, Vec3{7, 5, 3}},
		{"three subtractions", Vec3{20, 20, 20}, Vec3{5, 5, 5}, []Vec3{{3, 3, 3}, {2, 2, 2}}, Vec3{10, 10, 10}},
		{"with negatives", Vec3{10, 10, 10}, Vec3{5, 5, 5}, []Vec3{{-2, -2, -2}}, Vec3{7, 7, 7}},
		{"equivalent to nested Sub", Vec3{20, 30, 40}, Vec3{5, 10, 15}, []Vec3{{2, 3, 4}, {1, 1, 1}}, Vec3{12, 16, 20}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubMultiple(tt.u, tt.v0, tt.vs...)
			if result != tt.expected {
				t.Errorf("SubMultiple() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSubMultipleUsageExample(t *testing.T) {
	// Test the actual usage from ray.go
	camera := Vec3{0, 0, 0}
	horizontalHalf := Vec3{2, 0, 0}
	verticalHalf := Vec3{0, 1, 0}
	focal := Vec3{0, 0, 1}

	// SubMultiple(camera, horizontalHalf, verticalHalf, focal)
	// should equal camera - horizontalHalf - verticalHalf - focal
	result := SubMultiple(camera, horizontalHalf, verticalHalf, focal)
	expected := Vec3{-2, -1, -1}

	if result != expected {
		t.Errorf("SubMultiple() = %v, want %v", result, expected)
	}

	// Verify equivalence with nested Sub
	nested := Sub(Sub(Sub(camera, horizontalHalf), verticalHalf), focal)
	if result != nested {
		t.Errorf("SubMultiple() = %v, nested Sub() = %v, should be equal", result, nested)
	}
}

func TestMethodStyleAPI(t *testing.T) {
	// Test that method-style API produces same results as function-style
	v1 := Vec3{10, 20, 30}
	v2 := Vec3{1, 2, 3}
	v3 := Vec3{4, 5, 6}

	// Test Plus
	methodPlus := v1.Plus(v2, v3)
	funcPlus := AddMultiple(v1, v2, v3)
	if methodPlus != funcPlus {
		t.Errorf("Plus() = %v, AddMultiple() = %v, should be equal", methodPlus, funcPlus)
	}

	// Test Minus
	methodMinus := v1.Minus(v2, v3)
	funcMinus := SubMultiple(v1, v2, v3)
	if methodMinus != funcMinus {
		t.Errorf("Minus() = %v, SubMultiple() = %v, should be equal", methodMinus, funcMinus)
	}

	// Test Times
	methodTimes := v1.Times(2.5)
	funcTimes := SMul(v1, 2.5)
	if methodTimes != funcTimes {
		t.Errorf("Times() = %v, SMul() = %v, should be equal", methodTimes, funcTimes)
	}
}

func TestMethodStyleChaining(t *testing.T) {
	// Test realistic chaining as used in ray.go
	camera := XYZ(0, 0, 0)
	horizontal := XYZ(4, 0, 0)
	vertical := XYZ(0, 2, 0)
	focal := XYZ(0, 0, 1)

	// Method style (readable)
	upperLeft := camera.Minus(horizontal.Times(0.5), vertical.Times(0.5), focal)

	// Function style (equivalent)
	upperLeftFunc := SubMultiple(camera, SMul(horizontal, 0.5), SMul(vertical, 0.5), focal)

	if upperLeft != upperLeftFunc {
		t.Errorf("Method style = %v, function style = %v, should be equal", upperLeft, upperLeftFunc)
	}

	// Expected result
	expected := Vec3{-2, -1, -1}
	if upperLeft != expected {
		t.Errorf("upperLeft = %v, want %v", upperLeft, expected)
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

func TestDot(t *testing.T) {
	v1 := Vec3{1, 2, 3}
	v2 := Vec3{4, 5, 6}
	result := Dot(v1, v2)
	expected := 32.0 // 1*4 + 2*5 + 3*6
	if result != expected {
		t.Errorf("Dot() = %v, want %v", result, expected)
	}
}
