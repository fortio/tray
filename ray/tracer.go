// Package ray implements ray tracing on a small image.
// Inspired by https://raytracing.github.io/books/RayTracingInOneWeekend.html
package ray

import (
	"image"
	"math/rand/v2"
	"runtime"
	"sync"
)

// Tracer represents a ray tracing engine.
type Tracer struct {
	Camera          Vec3
	FocalLength     float64
	ViewportHeight  float64
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
		t.Camera = Vec3{0, .1, 5}
	}
	// Default camera / viewport setup
	if t.FocalLength <= 0 {
		t.FocalLength = 5
	}
	if t.ViewportHeight <= 0 {
		t.ViewportHeight = 1.5
	}
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

	aspectRatio := float64(t.width) / float64(t.height)
	viewportWidth := aspectRatio * t.ViewportHeight
	horizontal := XYZ(viewportWidth, 0, 0)
	vertical := XYZ(0, -t.ViewportHeight, 0) // y axis is inverted in image vs our world.
	pixelXVector := SDiv(horizontal, float64(t.width))
	pixelYVector := SDiv(vertical, float64(t.height))
	upperLeftCorner := t.Camera.Minus(horizontal.Times(0.5), vertical.Times(0.5), Vec3{0, 0, t.FocalLength})
	pixel00 := upperLeftCorner.Plus(Add(pixelXVector, pixelYVector).Times(0.5)) // up + (px + py)/2 (center of pixel)

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
			t.RenderLines(yStart, yEnd, pixel00, pixelXVector, pixelYVector, scene)
			wg.Done()
		})(startY, endY)
		startY = endY
	}
	wg.Wait()
	return t.imageData
}

func (t *Tracer) RenderLines(
	yStart, yEnd int, pixel00 Vec3, pixelXVector Vec3, pixelYVector Vec3, scene *Scene,
) {
	//nolint:gosec // not crypto use.
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
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
					deltaX, deltaY = SampleDiscRejRng(rng, t.RayRadius)
				}
				pixel := pixel00.Plus(pixelXVector.Times(float64(x)+deltaX), pixelYVector.Times(float64(y)+deltaY))
				rayDirection := pixel.Minus(t.Camera)
				ray := Ray{Origin: t.Camera, Direction: rayDirection, rng: rng}
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

type Ray struct {
	Origin    Vec3
	Direction Vec3
	rng       *rand.Rand
}

func (r *Ray) At(t float64) Vec3 {
	return Add(r.Origin, SMul(r.Direction, t))
}
