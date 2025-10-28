// Package ray implements ray tracing on a small image.
// Inspired by https://raytracing.github.io/books/RayTracingInOneWeekend.html
//
// Coordinate System:
// This package uses a right-handed coordinate system with:
//   - +X points right
//   - +Y points up
//   - +Z points backward (toward the camera)
//   - -Z points forward (into the scene)
//
// Scene objects should be positioned at negative Z values to appear in front
// of a camera at the origin. For example, a sphere at Vec3{0, 0, -5} is 5 units
// in front of a camera at Vec3{0, 0, 0} looking at Vec3{0, 0, -1}.
package ray

import (
	"image"
	"runtime"
	"sync"
)

// Tracer represents a ray tracing engine.
type Tracer struct {
	Camera
	MaxDepth        int
	NumRaysPerPixel int
	RayRadius       float64
	NumWorkers      int // Number of parallel workers; defaults to GOMAXPROCS if <= 0
	ProgressFunc    func(delta int)
	Seed            uint64 // Seed for random number generators; 0 means randomized each time
	width, height   int
	imageData       *image.RGBA
}

// New creates and initializes a new Tracer.
func New(width, height int) *Tracer {
	// Implementation of ray tracer initialization.
	return &Tracer{
		width:     width,
		height:    height,
		imageData: image.NewRGBA(image.Rect(0, 0, width, height)),
	}
}

// Render performs the ray tracing and returns the resulting image data.
func (t *Tracer) Render(scene *Scene) *image.RGBA {
	if scene == nil {
		scene = DefaultScene()
		// For now/for this scene:
		// t.Position = Vec3{0, .5, 5}
		t.Position = Vec3{-2, 2, 1}
		t.LookAt = Vec3{0, 0, -1}
		t.VerticalFoV = 20.0
		// t.LookAt = Vec3{-0.1, 0, -0.75} // look slight left and down and in front of the sphere
		// t.FocalLength = 5
		// t.VerticalFoV = 40.0
		t.Aperture = .1
		t.FocusDistance = Length(Sub(t.Position, t.LookAt))
	}
	// Need some/any light to get rays that aren't all black:
	if scene.Background.ColorA == (ColorF{}) && scene.Background.ColorB == (ColorF{}) {
		scene.Background = DefaultBackground()
	}
	// Other default values:
	if t.MaxDepth <= 0 {
		t.MaxDepth = 10
	}
	if t.NumRaysPerPixel <= 0 {
		t.NumRaysPerPixel = 1
	}
	if t.RayRadius <= 0 {
		t.RayRadius = 0.5
	}
	if t.NumWorkers <= 0 {
		t.NumWorkers = runtime.GOMAXPROCS(0)
	}
	// And zero value (0,0,0) for Camera is the right default
	// (when not hardcoded in nil scene case above).

	// Initialize camera viewport parameters (and set camera defaults if needed)
	t.Camera.Initialize(t.width, t.height)

	// Parallel rendering
	var wg sync.WaitGroup
	if t.NumWorkers == 1 {
		// Special case: single worker renders entire image (preserves exact RNG sequence)
		t.RenderLines(0, 0, t.height, scene)
	} else {
		// Work queue approach for dynamic load balancing across multiple workers
		// Divide image into chunks (smaller than worker count for better distribution)
		chunkSize := max(4, t.height/(t.NumWorkers*4))
		type workChunk struct{ startY, endY int }
		// numChunks = ceiling of t.height/chunkSize
		numChunks := (t.height + chunkSize - 1) / chunkSize
		workQueue := make(chan workChunk, numChunks)

		// Fill work queue with chunks
		for y := 0; y < t.height; y += chunkSize {
			workQueue <- workChunk{y, min(y+chunkSize, t.height)}
		}
		close(workQueue)

		// Workers pull chunks from queue until empty
		for range t.NumWorkers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for chunk := range workQueue {
					t.RenderLines(chunk.startY, chunk.startY, chunk.endY, scene)
				}
			}()
		}
		wg.Wait()
	}
	return t.imageData
}

func (t *Tracer) RenderLines(idx, yStart, yEnd int, scene *Scene) {
	rng := NewRandIdx(idx, t.Seed)
	multipleRays := t.NumRaysPerPixel > 1
	colorSumDiv := 1.0 / float64(t.NumRaysPerPixel)
	pix := t.imageData.Pix
	for y := yStart; y < yEnd; y++ {
		if t.ProgressFunc != nil {
			t.ProgressFunc(t.width)
		}
		for x := range t.width {
			// Compute ray for pixel (x, y)
			// Multiple rays per pixel for antialiasing (alternative from scaling the image up/down).
			colorSum := ColorF{0, 0, 0}
			for range t.NumRaysPerPixel {
				// Sub-pixel offset for antialiasing
				offsetX, offsetY := 0.0, 0.0 // Default to pixel center (0,0)
				if multipleRays {
					// Random offset within pixel for antialiasing
					offsetX, offsetY = rng.SampleDisc(t.RayRadius)
				}
				// Generate ray with depth of field (if Aperture > 0)
				ray := t.Camera.GetRay(rng, float64(x), float64(y), offsetX, offsetY)
				color := scene.RayColor(ray, t.MaxDepth)
				colorSum = Add(colorSum, color)
			}
			c := SMul(colorSum, colorSumDiv).ToSRGBA()
			// inline SetRGBA for performance
			off := t.imageData.PixOffset(x, y)
			s := pix[off : off+4 : off+4]
			s[0] = c.R
			s[1] = c.G
			s[2] = c.B
			s[3] = 255
		}
	}
}
