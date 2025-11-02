package ray

import (
	"testing"

	"fortio.org/rand"
)

var zero Vec3

func RandForTests() rand.Rand {
	return rand.NewRand(42)
}

func TestCamera_Initialize_PositionEqualsLookAt(t *testing.T) {
	// Test that Initialize handles Position == LookAt without panicking
	camera := Camera{
		Position: Vec3{1, 2, 3},
		LookAt:   Vec3{1, 2, 3}, // Same as Position
	}

	// This should not panic
	camera.Initialize(100, 100)

	// Verify that computed fields are set (non-zero)
	if camera.pixel00 == zero {
		t.Error("pixel00 should be initialized")
	}
	if camera.pixelXVector == zero {
		t.Error("pixelXVector should be initialized")
	}
	if camera.pixelYVector == zero {
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
	if camera.pixel00 == zero {
		t.Error("pixel00 should be initialized")
	}
	if camera.pixelXVector == zero {
		t.Error("pixelXVector should be initialized")
	}
	if camera.pixelYVector == zero {
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
	if camera.pixel00 == zero {
		t.Error("pixel00 should be initialized")
	}
	if camera.pixelXVector == zero {
		t.Error("pixelXVector should be initialized")
	}
	if camera.pixelYVector == zero {
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

	rng := RandForTests()
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

	rng := RandForTests()
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

func TestCamera_GetRay_PixelCenter(t *testing.T) {
	// Test that offset (0,0) produces a ray through the exact pixel center
	// Simple camera setup for easy math verification
	camera := Camera{
		Position:    Vec3{0, 0, 0},  // Camera at origin
		LookAt:      Vec3{0, 0, -1}, // Looking down -Z
		VerticalFoV: 90.0,
		FocalLength: 1.0,
	}
	camera.Initialize(10, 10) // 10x10 image

	rng := rand.NewRand(42)

	// Get ray for pixel (5, 5) with offset (0, 0) - should be exact center
	ray := camera.GetRay(rng, 5, 5, 0.0, 0.0)

	// Ray origin should be camera position
	if ray.Origin != camera.Position {
		t.Errorf("Ray origin = %v, want %v", ray.Origin, camera.Position)
	}

	// The ray should point to pixel00 + 5*pixelXVector + 5*pixelYVector
	// This is the center of pixel (5, 5)
	expectedTarget := camera.pixel00.Plus(
		SMul(camera.pixelXVector, 5),
		SMul(camera.pixelYVector, 5),
	)

	// Ray direction should point toward expectedTarget
	// Normalize both to compare directions
	expectedDir := Unit(Sub(expectedTarget, camera.Position))
	actualDir := Unit(ray.Direction)

	// Check if directions are essentially the same (within floating point tolerance)
	diff := Sub(expectedDir, actualDir)
	if !NearZero(diff) {
		t.Errorf("Ray direction mismatch:\nExpected: %v\nActual: %v\nDiff: %v",
			expectedDir, actualDir, diff)
	}
}

func TestCamera_GetRay_OffsetFromCenter(t *testing.T) {
	// Test that non-zero offsets produce rays offset from pixel center
	camera := Camera{
		Position:    Vec3{0, 0, 0},
		LookAt:      Vec3{0, 0, -1},
		VerticalFoV: 90.0,
		FocalLength: 1.0,
	}
	camera.Initialize(10, 10)

	rng := RandForTests()

	// Get rays with different offsets for the same pixel
	rayCenter := camera.GetRay(rng, 5, 5, 0.0, 0.0)
	rayOffset := camera.GetRay(rng, 5, 5, 0.3, 0.2)

	// Directions should be different (pointing to different parts of the pixel)
	if rayCenter.Direction == rayOffset.Direction {
		t.Error("Expected different ray directions for different offsets")
	}

	// Origins should be the same (no aperture)
	if rayCenter.Origin != rayOffset.Origin {
		t.Error("Ray origins should be the same with aperture=0")
	}
}

func TestRichSceneCamera_RendersNonBlackImage(t *testing.T) {
	// Test that RichSceneCamera + RichScene produces a non-black image
	// Use very low resolution to keep test fast
	width, height := 20, 20
	rng := RandForTests()
	scene := RichScene(rng)

	tracer := New(width, height)
	tracer.Camera = RichSceneCamera()
	tracer.MaxDepth = 10
	tracer.NumRaysPerPixel = 2 // Low but not 1, to get some antialiasing

	img := tracer.Render(scene)

	// Check that the image is not all black
	// Count non-black pixels
	nonBlackPixels := 0
	totalPixels := width * height

	for y := range height {
		for x := range width {
			r, g, b, _ := img.At(x, y).RGBA()
			// RGBA() returns values in [0, 65535] range
			if r > 0 || g > 0 || b > 0 {
				nonBlackPixels++
			}
		}
	}

	// At least 50% of pixels should have some color
	// (the scene should render spheres against a background)
	minNonBlackPixels := totalPixels / 2
	if nonBlackPixels < minNonBlackPixels {
		t.Errorf("Image is mostly black: only %d/%d pixels are non-black (expected at least %d)",
			nonBlackPixels, totalPixels, minNonBlackPixels)
	}

	t.Logf("Rendered RichScene with %d/%d non-black pixels", nonBlackPixels, totalPixels)
}
