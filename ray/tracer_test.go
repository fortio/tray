package ray

import (
	"runtime"
	"sync/atomic"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"small", 10, 10},
		{"wide", 100, 50},
		{"tall", 50, 100},
		{"large", 1920, 1080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracer := New(tt.width, tt.height)
			if tracer == nil {
				t.Fatal("New() returned nil")
			}
			if tracer.width != tt.width {
				t.Errorf("width = %d, want %d", tracer.width, tt.width)
			}
			if tracer.height != tt.height {
				t.Errorf("height = %d, want %d", tracer.height, tt.height)
			}
			if tracer.imageData == nil {
				t.Fatal("imageData is nil")
			}
			bounds := tracer.imageData.Bounds()
			if bounds.Dx() != tt.width {
				t.Errorf("image width = %d, want %d", bounds.Dx(), tt.width)
			}
			if bounds.Dy() != tt.height {
				t.Errorf("image height = %d, want %d", bounds.Dy(), tt.height)
			}
		})
	}
}

func TestRender_DefaultScene(t *testing.T) {
	tracer := New(10, 10)
	img := tracer.Render(nil)

	if img == nil {
		t.Fatal("Render() returned nil")
	}
	if img != tracer.imageData {
		t.Error("Render() should return the tracer's imageData")
	}
	bounds := img.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("image size = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}

	// Check that pixels were actually set (not all black)
	allBlack := true
	for y := range 10 {
		for x := range 10 {
			r, g, b, a := img.At(x, y).RGBA()
			if a == 0 {
				t.Errorf("pixel (%d,%d) has zero alpha", x, y)
			}
			if r != 0 || g != 0 || b != 0 {
				allBlack = false
			}
		}
	}
	if allBlack {
		t.Error("all pixels are black, expected some color from scene")
	}
}

func TestRender_CustomScene(t *testing.T) {
	tracer := New(5, 5)
	scene := &Scene{
		Objects: []Hittable{
			&Sphere{
				Center: Vec3{0, 0, -1},
				Radius: 0.5,
				Mat:    Lambertian{Albedo: ColorF{1, 0, 0}},
			},
		},
	}

	img := tracer.Render(scene)
	if img == nil {
		t.Fatal("Render() returned nil")
	}

	// Verify alpha channel is set
	for y := range 5 {
		for x := range 5 {
			_, _, _, a := img.At(x, y).RGBA()
			if a != 0xffff {
				t.Errorf("pixel (%d,%d) alpha = %d, want %d", x, y, a>>8, 255)
			}
		}
	}
}

func TestRender_DefaultParameters(t *testing.T) {
	tracer := New(5, 5)
	// Don't set any parameters, let them all be defaults
	_ = tracer.Render(nil)

	// Verify defaults were applied
	if tracer.FocalLength != 5 {
		t.Errorf("FocalLength = %f, want 5", tracer.FocalLength)
	}
	if tracer.ViewportHeight != 1.5 {
		t.Errorf("ViewportHeight = %f, want 1.5", tracer.ViewportHeight)
	}
	if tracer.MaxDepth != 10 {
		t.Errorf("MaxDepth = %d, want 10", tracer.MaxDepth)
	}
	if tracer.NumRaysPerPixel != 1 {
		t.Errorf("NumRaysPerPixel = %d, want 1", tracer.NumRaysPerPixel)
	}
	if tracer.RayRadius != 0.5 {
		t.Errorf("RayRadius = %f, want 0.5", tracer.RayRadius)
	}
	expectedWorkers := runtime.GOMAXPROCS(0)
	if tracer.NumWorkers != expectedWorkers {
		t.Errorf("NumWorkers = %d, want %d", tracer.NumWorkers, expectedWorkers)
	}
}

func TestRender_CustomParameters(t *testing.T) {
	tracer := New(5, 5)
	tracer.Camera = &Camera{}
	tracer.Position = Vec3{1, 2, 3}
	tracer.FocalLength = 10
	tracer.ViewportHeight = 2.0
	tracer.MaxDepth = 20
	tracer.NumRaysPerPixel = 4
	tracer.RayRadius = 1.0
	tracer.NumWorkers = 2

	_ = tracer.Render(DefaultScene())

	// Verify custom values are preserved
	if tracer.Camera.Position != (Vec3{1, 2, 3}) {
		t.Errorf("Camera.Position = %v, want {1, 2, 3}", tracer.Camera.Position)
	}
	if tracer.FocalLength != 10 {
		t.Errorf("FocalLength = %f, want 10", tracer.FocalLength)
	}
	if tracer.ViewportHeight != 2.0 {
		t.Errorf("ViewportHeight = %f, want 2.0", tracer.ViewportHeight)
	}
	if tracer.MaxDepth != 20 {
		t.Errorf("MaxDepth = %d, want 20", tracer.MaxDepth)
	}
	if tracer.NumRaysPerPixel != 4 {
		t.Errorf("NumRaysPerPixel = %d, want 4", tracer.NumRaysPerPixel)
	}
	if tracer.RayRadius != 1.0 {
		t.Errorf("RayRadius = %f, want 1.0", tracer.RayRadius)
	}
	if tracer.NumWorkers != 2 {
		t.Errorf("NumWorkers = %d, want 2", tracer.NumWorkers)
	}
}

func TestRender_ProgressCallback(t *testing.T) {
	tracer := New(10, 8)
	var totalProgress atomic.Int32
	tracer.ProgressFunc = func(delta int) {
		totalProgress.Add(int32(delta))
	}

	_ = tracer.Render(DefaultScene())

	// Should have called progress once per row with width pixels
	expected := int32(10 * 8)
	if totalProgress.Load() != expected {
		t.Errorf("total progress = %d, want %d", totalProgress.Load(), expected)
	}
}

func TestRender_ParallelRendering(t *testing.T) {
	// Test with different worker counts
	tests := []struct {
		name       string
		numWorkers int
		width      int
		height     int
	}{
		{"single_worker", 1, 10, 10},
		{"two_workers", 2, 10, 10},
		{"more_workers_than_rows", 20, 10, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracer := New(tt.width, tt.height)
			tracer.NumWorkers = tt.numWorkers
			img := tracer.Render(DefaultScene())

			if img == nil {
				t.Fatal("Render() returned nil")
			}

			// Verify all pixels have been set (non-zero alpha)
			for y := range tt.height {
				for x := range tt.width {
					_, _, _, a := img.At(x, y).RGBA()
					if a == 0 {
						t.Errorf("pixel (%d,%d) not rendered (zero alpha)", x, y)
					}
				}
			}
		})
	}
}

func TestRender_MultipleRaysPerPixel(t *testing.T) {
	// Test that multiple rays per pixel doesn't crash and produces valid output
	tests := []struct {
		name    string
		numRays int
	}{
		{"single_ray", 1},
		{"four_rays", 4},
		{"ten_rays", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracer := New(5, 5)
			tracer.NumRaysPerPixel = tt.numRays
			img := tracer.Render(DefaultScene())

			if img == nil {
				t.Fatal("Render returned nil")
			}

			// Verify all pixels have valid colors
			for y := range 5 {
				for x := range 5 {
					_, _, _, a := img.At(x, y).RGBA()
					if a != 0xffff {
						t.Errorf("pixel (%d,%d) has invalid alpha", x, y)
					}
				}
			}
		})
	}
}

func TestRenderLines(t *testing.T) {
	tracer := New(10, 10)
	tracer.Camera = &Camera{}
	tracer.FocalLength = 5
	tracer.ViewportHeight = 1.5
	tracer.MaxDepth = 10
	tracer.NumRaysPerPixel = 1
	tracer.RayRadius = 0.5

	scene := DefaultScene()
	// Initialize camera viewport parameters
	tracer.Camera.Initialize(tracer.width, tracer.height)

	// Render just the first 3 lines
	tracer.RenderLines(0, 3, scene)

	// Check that first 3 rows are rendered (non-zero alpha)
	for y := range 3 {
		for x := range 10 {
			_, _, _, a := tracer.imageData.At(x, y).RGBA()
			if a == 0 {
				t.Errorf("pixel (%d,%d) not rendered", x, y)
			}
		}
	}

	// Check that remaining rows are still unrendered (zero values)
	allZero := true
	for y := 3; y < 10; y++ {
		for x := range 10 {
			r, g, b, a := tracer.imageData.At(x, y).RGBA()
			if r != 0 || g != 0 || b != 0 || a != 0 {
				allZero = false
				break
			}
		}
	}
	if !allZero {
		t.Error("rows 3-9 should not have been rendered")
	}
}

func TestRender_EmptyScene(t *testing.T) {
	tracer := New(5, 5)
	scene := &Scene{Objects: []Hittable{}}
	img := tracer.Render(scene)

	// Should render background gradient (sky)
	for y := range 5 {
		for x := range 5 {
			r, g, b, a := img.At(x, y).RGBA()
			if a == 0 {
				t.Errorf("pixel (%d,%d) has zero alpha", x, y)
			}
			// Background should have some blue component
			if b == 0 {
				t.Errorf("pixel (%d,%d) missing blue from sky gradient", x, y)
			}
			// Should not be pure black
			if r == 0 && g == 0 && b == 0 {
				t.Errorf("pixel (%d,%d) is black, expected sky gradient", x, y)
			}
		}
	}
}
