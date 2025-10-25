// Package ray implements ray tracing on a small image.
// Inspired by https://raytracing.github.io/books/RayTracingInOneWeekend.html
package ray

import (
	"image"
	"math"
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
	width, height   int
	imageData       *image.RGBA
}

type HitRecord struct {
	Point     Vec3
	Normal    Vec3
	T         float64
	FrontFace bool
}

func (hr *HitRecord) SetFaceNormal(r Ray, outwardNormal Vec3) {
	hr.FrontFace = Dot(r.Direction, outwardNormal) < 0
	if hr.FrontFace {
		hr.Normal = outwardNormal
	} else {
		hr.Normal = Neg(outwardNormal)
	}
}

type Hittable interface {
	Hit(r Ray, interval Interval) (bool, HitRecord)
}

func (s *Scene) Hit(r Ray, interval Interval) (bool, HitRecord) {
	hitAnything := false
	closestSoFar := interval.End
	var tempRec HitRecord

	for _, object := range s.Objects {
		if hit, rec := object.Hit(r, Interval{Start: interval.Start, End: closestSoFar}); hit {
			hitAnything = true
			closestSoFar = rec.T
			tempRec = rec
		}
	}
	return hitAnything, tempRec
}

type Sphere struct {
	Center Vec3
	Radius float64
}

func (s *Sphere) Hit(r Ray, i Interval) (bool, HitRecord) {
	oc := Sub(s.Center, r.Origin)
	a := LengthSquared(r.Direction)
	h := Dot(r.Direction, oc)
	c := LengthSquared(oc) - s.Radius*s.Radius
	discriminant := h*h - a*c
	if discriminant < 0 {
		return false, HitRecord{}
	}
	sqrtD := math.Sqrt(discriminant)
	root := (h - sqrtD) / a
	if !i.Surrounds(root) {
		root = (h + sqrtD) / a
		if !i.Surrounds(root) {
			return false, HitRecord{}
		}
	}
	hr := HitRecord{Point: r.At(root), T: root}
	outwardNormal := SDiv(Sub(hr.Point, s.Center), s.Radius)
	hr.SetFaceNormal(r, outwardNormal)
	return true, hr
}

type Scene struct {
	Objects []Hittable
}

func (s *Scene) RayColor(r Ray, depth int) ColorF {
	if depth <= 0 {
		return ColorF{0, 0, 0}
	}
	if hit, hr := s.Hit(r, FrontEpsilon); hit {
		direction := Add(hr.Normal, RandomUnitVectorRng[Vec3](r.rng))
		newRay := Ray{Origin: hr.Point, Direction: direction, rng: r.rng}
		return SMul(s.RayColor(newRay, depth-1), 0.5)
	}
	unit := Unit(r.Direction)
	a := 0.5 * (unit.Y() + 1.0)
	white := ColorF{1.0, 1.0, 1.0}
	blue := ColorF{0.4, 0.65, 1.0}
	blend := Add(SMul(white, 1.0-a), SMul(blue, a))
	return blend
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

// SampleDiscRej returns a random point (x,y) within a disc of radius r.
// Rejection sampling method.
//
//nolint:gosec // not crypto use.
func SampleDiscRej(r float64) (x, y float64) {
	for {
		x = 2*rand.Float64() - 1.0
		y = 2*rand.Float64() - 1.0
		if x*x+y*y <= 1 {
			break
		}
	}
	return r * x, r * y
}

// SampleDiscRejRng returns a random point (x,y) within a disc of radius r
// using the provided random source.
func SampleDiscRejRng(rng *rand.Rand, r float64) (x, y float64) {
	for {
		x = 2*rng.Float64() - 1.0
		y = 2*rng.Float64() - 1.0
		if x*x+y*y <= 1 {
			break
		}
	}
	return r * x, r * y
}

// SampleDiscAngle returns a random point (x,y) within a disc of radius r.
// Angle method.
//
//nolint:gosec // not crypto use.
func SampleDiscAngle(r float64) (x, y float64) {
	theta := 2.0 * math.Pi * rand.Float64()
	rad := r * math.Sqrt(rand.Float64())
	x = rad * math.Cos(theta)
	y = rad * math.Sin(theta)
	return x, y
}

// Render performs the ray tracing and returns the resulting image data.
func (t *Tracer) Render(scene *Scene) *image.RGBA {
	if scene == nil {
		scene = &Scene{
			// Default scene with two spheres.
			Objects: []Hittable{
				&Sphere{Center: Vec3{0, 0, -1}, Radius: 0.5},
				&Sphere{Center: Vec3{0, -100.5, -1}, Radius: 100},
			},
		}
	}
	// Default camera / viewport setup
	if t.FocalLength <= 0 {
		t.FocalLength = 1.0
	}
	if t.ViewportHeight <= 0 {
		t.ViewportHeight = 2.0
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
	// And zero value (0,0,0) for Camera is the right default.

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

func (t *Tracer) RenderLines(yStart, yEnd int, pixel00 Vec3, pixelXVector Vec3, pixelYVector Vec3, scene *Scene) {
	//nolint:gosec // not crypto use.
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	multipleRays := t.NumRaysPerPixel > 1
	colorSumDiv := 1.0 / float64(t.NumRaysPerPixel)
	pix := t.imageData.Pix
	for y := yStart; y < yEnd; y++ {
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
