package ray

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
	// Aperture is the diameter of the camera's aperture. Zero means pinhole (no blur).
	Aperture float64
	// Computed fields (initialized by Initialize)
	pixel00      Vec3
	pixelXVector Vec3
	pixelYVector Vec3
}

func DefaultCamera() *Camera {
	return &Camera{
		Position:       Vec3{0, 0, 0},
		LookAt:         Vec3{0, 0, -1},
		Up:             Vec3{0, 1, 0},
		FocalLength:    1.0,
		ViewportHeight: 2,
		Aperture:       0, // pinhole camera (no blur)
	}
}

// Initialize computes the viewport parameters for the given image dimensions.
// Must be called before rendering.
func (c *Camera) Initialize(width, height int) {
	aspectRatio := float64(width) / float64(height)
	viewportWidth := aspectRatio * c.ViewportHeight
	horizontal := XYZ(viewportWidth, 0, 0)
	vertical := XYZ(0, -c.ViewportHeight, 0) // y axis is inverted in image vs our world.
	c.pixelXVector = SDiv(horizontal, float64(width))
	c.pixelYVector = SDiv(vertical, float64(height))
	upperLeftCorner := c.Position.Minus(horizontal.Times(0.5), vertical.Times(0.5), Vec3{0, 0, c.FocalLength})
	c.pixel00 = upperLeftCorner.Plus(Add(c.pixelXVector, c.pixelYVector).Times(0.5)) // up + (px + py)/2 (center of pixel)
}
