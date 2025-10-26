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
	if camera.VerticalFoV != 90.0 {
		t.Errorf("VerticalFoV = %f, want 90.0", camera.VerticalFoV)
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
		Position:    Vec3{0, 0, 5},
		LookAt:      Vec3{0, 0, 0},
		Up:          Vec3{0, 1, 0},
		FocalLength: 2.0,
		VerticalFoV: 60.0,
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
	if camera.VerticalFoV != 60.0 {
		t.Errorf("VerticalFoV changed: %f", camera.VerticalFoV)
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

func TestCamera_GetRay_Pinhole(t *testing.T) {
	// Test that aperture=0 produces rays from the same origin
	camera := Camera{
		Position:    Vec3{0, 0, 5},
		LookAt:      Vec3{0, 0, 0},
		Aperture:    0, // Pinhole camera
		FocalLength: 1.0,
	}
	camera.Initialize(100, 100)

	rng := NewRandomSource()
	ray1 := camera.GetRay(rng, 50, 50, 0.0, 0.0)
	ray2 := camera.GetRay(rng, 50, 50, 0.0, 0.0)
	ray3 := camera.GetRay(rng, 25, 75, 0.0, 0.0)

	// All rays should originate from camera position
	if ray1.Origin != camera.Position {
		t.Errorf("Ray origin = %v, want %v", ray1.Origin, camera.Position)
	}
	if ray2.Origin != camera.Position {
		t.Errorf("Ray origin = %v, want %v", ray2.Origin, camera.Position)
	}
	if ray3.Origin != camera.Position {
		t.Errorf("Ray origin = %v, want %v", ray3.Origin, camera.Position)
	}
}

func TestCamera_GetRay_DepthOfField(t *testing.T) {
	// Test that aperture>0 produces rays from different origins
	camera := Camera{
		Position:      Vec3{0, 0, 5},
		LookAt:        Vec3{0, 0, 0},
		Aperture:      0.5, // Non-zero aperture for depth of field
		FocalLength:   1.0,
		FocusDistance: 5.0, // Focus at distance 5
	}
	camera.Initialize(100, 100)

	rng := NewRandomSource()
	ray1 := camera.GetRay(rng, 50, 50, 0.0, 0.0)
	ray2 := camera.GetRay(rng, 50, 50, 0.0, 0.0)

	// With aperture>0, rays should originate from different positions (sampling lens disk)
	if ray1.Origin == ray2.Origin {
		t.Error("Expected different ray origins with aperture>0, got same origin")
	}

	// Origins should be near camera position (within aperture radius)
	dist1 := Length(Sub(ray1.Origin, camera.Position))
	dist2 := Length(Sub(ray2.Origin, camera.Position))
	maxDist := camera.Aperture / 2
	if dist1 > maxDist {
		t.Errorf("Ray origin too far from camera: %f > %f", dist1, maxDist)
	}
	if dist2 > maxDist {
		t.Errorf("Ray origin too far from camera: %f > %f", dist2, maxDist)
	}
}

func TestCamera_FocusDistance_Default(t *testing.T) {
	// Test that FocusDistance defaults to FocalLength
	camera := Camera{
		FocalLength: 2.5,
		// FocusDistance not set
	}
	camera.Initialize(100, 100)

	if camera.FocusDistance != camera.FocalLength {
		t.Errorf("FocusDistance = %f, want %f (FocalLength)", camera.FocusDistance, camera.FocalLength)
	}
}
