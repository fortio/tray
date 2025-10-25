package ray

import "math"

type Camera struct {
	// Position is where the camera is located in 3D space.
	Position Vec3
	// LookAt is the point in 3D space the camera is looking at.
	LookAt Vec3
	// Up is the upward direction for the camera. (typically {0,1,0} - Y axis up)
	Up Vec3
	// ViewportHeight is the height of the viewport in world units.
	ViewportHeight float64
	// FocalLength is the distance from the camera to the image plane.
	FocalLength float64
	// Aperture is the diameter of the camera's aperture. Smaller values produce more depth of field blur.
	Aperture float64
}

func DefaultCamera() *Camera {
	return &Camera{
		Position:       Vec3{0, 0, 0},
		LookAt:         Vec3{0, 0, -1},
		Up:             Vec3{0, 1, 0},
		FocalLength:    1.0,
		ViewportHeight: 2,
		Aperture:       math.Inf(1), // no blur
	}
}
