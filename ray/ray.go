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
	pixelXVector := SDiv(horizontal, float64(t.width))
	pixelYVector := SDiv(vertical, float64(t.height))
	upperLeftCorner := Sub(Sub(Sub(camera, SDiv(horizontal, 2)), SDiv(vertical, 2)), Vec3{0, 0, focalLength})
	pixel00 := Add(upperLeftCorner, SDiv(Add(pixelXVector, pixelYVector), 2)) // up + (px + py)/2 (center of pixel)

	for y := range t.height {
		for x := range t.width {
			// Compute ray for pixel (x, y)
			pixel := Add(Add(pixel00, SMul(pixelXVector, float64(x))), SMul(pixelYVector, float64(y)))
			rayDirection := Sub(pixel, camera)
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
	return Add(r.Origin, SMul(r.Direction, t))
}
