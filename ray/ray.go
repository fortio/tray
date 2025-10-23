// Package ray implements ray tracing on a small image.
// Inspired by https://raytracing.github.io/books/RayTracingInOneWeekend.html
package ray

import (
	"image"
	"image/color"
)

// Tracer represents a ray tracing engine.
type Tracer struct {
	// Fields for ray tracing state would go here.
	width, height int
	imageData     *image.RGBA
}

type Scene struct {
	// Fields defining the scene to be rendered.
}

func (s *Scene) TraceRay(r Ray) color.RGBA {
	// Placeholder implementation: return a color based on ray direction.

	unit := r.Direction.Unit()
	a := 0.5 * (unit.Y() + 1.0)
	white := ColorF(1.0, 1.0, 1.0)
	blue := ColorF(0.5, 0.7, 1.0)
	blend := white.SMul(1.0 - a).Add(blue.SMul(a))
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
func (t *Tracer) Render(scene Scene) *image.RGBA {
	// Implementation of ray tracing rendering.
	_ = scene                            // to avoid unused variable warning
	t.imageData.Set(10, 10, image.White) // Placeholder operation

	focalLength := 1.0
	camera := Vec3{0, 0, 0}
	viewportHeight := 2.0
	aspectRatio := float64(t.width) / float64(t.height)
	viewportWidth := aspectRatio * viewportHeight
	horizontal := Vec3{viewportWidth, 0, 0}
	vertical := Vec3{0, -viewportHeight, 0} // y axis is inverted in image vs our world.
	pixelXVector := horizontal.SDiv(float64(t.width))
	pixelYVector := vertical.SDiv(float64(t.height))
	upperLeftCorner := camera.Sub(horizontal.SDiv(2)).Sub(vertical.SDiv(2)).Sub(Vec3{0, 0, focalLength})
	pixel00 := upperLeftCorner.Add(pixelXVector.Add(pixelYVector).SDiv(2)) // up + (px + py)/2 (center of pixel)

	for y := range t.height {
		for x := range t.width {
			// Compute ray for pixel (x, y)
			pixel := pixel00.Add(pixelXVector.SMul(float64(x))).Add(pixelYVector.SMul(float64(y)))
			rayDirection := pixel.Sub(camera)
			ray := Ray{Origin: camera, Direction: rayDirection}
			color := scene.TraceRay(ray)
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
	return r.Origin.Add(r.Direction.SMul(t))
}
