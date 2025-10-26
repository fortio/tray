package ray

import "testing"

func TestCamera_Initialize_PositionEqualsLookAt(t *testing.T) {
	// Test that Initialize handles Position == LookAt without panicking
	camera := Camera{
		Position: Vec3{1, 2, 3},
		LookAt:   Vec3{1, 2, 3}, // Same as Position
	}

	// This should not panic
	camera.Initialize(100, 100)

	// Verify that computed fields are set (non-zero)
	if camera.pixel00 == (Vec3{}) {
		t.Error("pixel00 should be initialized")
	}
	if camera.pixelXVector == (Vec3{}) {
		t.Error("pixelXVector should be initialized")
	}
	if camera.pixelYVector == (Vec3{}) {
		t.Error("pixelYVector should be initialized")
	}
}

func TestCamera_Initialize_Defaults(t *testing.T) {
	// Test that Initialize sets default values for zero fields
	camera := Camera{}
	camera.Initialize(100, 100)

	if camera.FocalLength != 1.0 {
		t.Errorf("FocalLength = %f, want 1.0", camera.FocalLength)
	}
	if camera.ViewportHeight != 2.0 {
		t.Errorf("ViewportHeight = %f, want 2.0", camera.ViewportHeight)
	}
	if camera.Up != (Vec3{0, 1, 0}) {
		t.Errorf("Up = %v, want {0, 1, 0}", camera.Up)
	}
	// When both Position and LookAt start at zero, LookAt gets set to {0,0,-1}
	if camera.LookAt != (Vec3{0, 0, -1}) {
		t.Errorf("LookAt = %v, want {0, 0, -1}", camera.LookAt)
	}

	// Verify computed fields are initialized
	if camera.pixel00 == (Vec3{}) {
		t.Error("pixel00 should be initialized")
	}
	if camera.pixelXVector == (Vec3{}) {
		t.Error("pixelXVector should be initialized")
	}
	if camera.pixelYVector == (Vec3{}) {
		t.Error("pixelYVector should be initialized")
	}
}

func TestCamera_Initialize_CustomValues(t *testing.T) {
	// Test that Initialize preserves custom values
	camera := Camera{
		Position:       Vec3{0, 0, 5},
		LookAt:         Vec3{0, 0, 0},
		Up:             Vec3{0, 1, 0},
		FocalLength:    2.0,
		ViewportHeight: 3.0,
	}
	camera.Initialize(100, 100)

	// Verify that input fields are preserved
	if camera.Position != (Vec3{0, 0, 5}) {
		t.Errorf("Position changed: %v", camera.Position)
	}
	if camera.LookAt != (Vec3{0, 0, 0}) {
		t.Errorf("LookAt changed: %v", camera.LookAt)
	}
	if camera.FocalLength != 2.0 {
		t.Errorf("FocalLength changed: %f", camera.FocalLength)
	}
	if camera.ViewportHeight != 3.0 {
		t.Errorf("ViewportHeight changed: %f", camera.ViewportHeight)
	}

	// Verify computed fields are initialized
	if camera.pixel00 == (Vec3{}) {
		t.Error("pixel00 should be initialized")
	}
	if camera.pixelXVector == (Vec3{}) {
		t.Error("pixelXVector should be initialized")
	}
	if camera.pixelYVector == (Vec3{}) {
		t.Error("pixelYVector should be initialized")
	}
}
