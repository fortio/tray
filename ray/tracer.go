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
		t.Position = Vec3{0, .5, 5}
		t.LookAt = Vec3{-0.1, 0, -0.75} // look slight left and down and in front of the sphere
		t.FocalLength = 5
		t.ViewportHeight = 1.5
	}
	// Need some/any light to get rays that aren't all black:
	if scene.Background == nil {
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

	// Parallel rendering: divide work into horizontal bands
	var wg sync.WaitGroup
	rowsPerWorker := t.height / t.NumWorkers
	remainder := t.height % t.NumWorkers
	startY := 0
	for i := range t.NumWorkers {
		// Distribute remainder rows to first workers (one extra row each)
		endY := startY + rowsPerWorker
		if i < remainder {
			endY++
		}
		wg.Add(1)
		go (func(yStart, yEnd int) {
			t.RenderLines(yStart, yEnd, scene)
			wg.Done()
		})(startY, endY)
		startY = endY
	}
	wg.Wait()
	return t.imageData
}

func (t *Tracer) RenderLines(yStart, yEnd int, scene *Scene) {
	rng := NewRandomSource()
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
				deltaX, deltaY := 0.0, 0.0
				if multipleRays {
					deltaX, deltaY = rng.SampleDisc(t.RayRadius)
				}
				pixel := t.pixel00.Plus(t.pixelXVector.Times(float64(x)+deltaX), t.pixelYVector.Times(float64(y)+deltaY))
				rayDirection := pixel.Minus(t.Position)
				ray := rng.NewRay(t.Position, rayDirection)
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
