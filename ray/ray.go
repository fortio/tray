// Package ray implements ray tracing on a small image.
// Inspired by https://raytracing.github.io/books/RayTracingInOneWeekend.html
package ray

import (
	"image"
	"image/color"
	"math"
)

// Tracer represents a ray tracing engine.
type Tracer struct {
	Camera         Vec3
	FocalLength    float64
	ViewportHeight float64
	width, height  int
	imageData      *image.RGBA
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

func (s *Scene) RayColor(r Ray) color.RGBA {
	if hit, hr := s.Hit(r, Front); hit {
		N := hr.Normal
		return SMul(ColorF{N.X() + 1, N.Y() + 1, N.Z() + 1}, 0.5).ToRGBA()
	}
	unit := Unit(r.Direction)
	a := 0.5 * (unit.Y() + 1.0)
	white := ColorF{1.0, 1.0, 1.0}
	blue := ColorF{0.5, 0.7, 1.0}
	blend := Add(SMul(white, 1.0-a), SMul(blue, a))
	return blend.ToRGBA()
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
		scene = &Scene{
			// Default scene with two spheres.
			Objects: []Hittable{
				&Sphere{Center: Vec3{0, 0, -1}, Radius: 0.5},
				&Sphere{Center: Vec3{0, -100.5, -1}, Radius: 100},
			},
		}
	}
	// Default camera / viewport setup
	if t.FocalLength == 0 {
		t.FocalLength = 1.0
	}
	if t.ViewportHeight == 0 {
		t.ViewportHeight = 2.0
	}
	aspectRatio := float64(t.width) / float64(t.height)
	viewportWidth := aspectRatio * t.ViewportHeight
	horizontal := XYZ(viewportWidth, 0, 0)
	vertical := XYZ(0, -t.ViewportHeight, 0) // y axis is inverted in image vs our world.
	pixelXVector := SDiv(horizontal, float64(t.width))
	pixelYVector := SDiv(vertical, float64(t.height))
	upperLeftCorner := t.Camera.Minus(horizontal.Times(0.5), vertical.Times(0.5), Vec3{0, 0, t.FocalLength})
	pixel00 := upperLeftCorner.Plus(Add(pixelXVector, pixelYVector).Times(0.5)) // up + (px + py)/2 (center of pixel)

	for y := range t.height {
		for x := range t.width {
			// Compute ray for pixel (x, y)
			pixel := pixel00.Plus(pixelXVector.Times(float64(x)), pixelYVector.Times(float64(y)))
			rayDirection := pixel.Minus(t.Camera)
			ray := Ray{Origin: t.Camera, Direction: rayDirection}
			color := scene.RayColor(ray)
			t.imageData.SetRGBA(x, y, color)
		}
	}
	return t.imageData
}

type Ray struct {
	Origin    Vec3
	Direction Vec3
}

func (r *Ray) At(t float64) Vec3 {
	return Add(r.Origin, SMul(r.Direction, t))
}

type Interval struct {
	Start, End float64
}

func (i Interval) Length() float64 {
	return i.End - i.Start
}

func (i Interval) Contains(t float64) bool {
	return t >= i.Start && t <= i.End
}

func (i Interval) Surrounds(t float64) bool {
	return t > i.Start && t < i.End
}

func (i Interval) Clamp(t float64) float64 {
	if t < i.Start {
		return i.Start
	}
	if t > i.End {
		return i.End
	}
	return t
}

var (
	Empty    = Interval{Start: math.Inf(1), End: math.Inf(-1)}
	Universe = Interval{Start: math.Inf(-1), End: math.Inf(1)}
	Front    = Interval{Start: 0, End: math.Inf(1)}
)
