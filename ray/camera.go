package ray

type Camera struct {
	// Position is where the camera is located in 3D space.
	Position Vec3
	// LookAt is the point in 3D space the camera is looking at.
	// Together with Position, this defines the view direction.
	LookAt Vec3
	// Up is the upward direction for the camera, controlling the camera's roll/rotation
	// around the view axis. Typically {0,1,0} for Y-up. Changing Up rotates the image
	// (e.g., {0,-1,0} would flip the image upside down, {1,0,0} would tilt 90 degrees).
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

// Initialize computes the viewport parameters for the given image dimensions.
// Sets default values for any zero-valued fields. Must be called before rendering.
func (c *Camera) Initialize(width, height int) {
	// Set defaults for zero-valued fields
	if c.FocalLength == 0 {
		c.FocalLength = 1.0
	}
	if c.ViewportHeight == 0 {
		c.ViewportHeight = 2.0
	}
	if c.Up == (Vec3{}) {
		c.Up = Vec3{0, 1, 0}
	}
	if c.LookAt == (Vec3{}) {
		c.LookAt = Vec3{0, 0, -1}
	}
	// Position default is (0,0,0) which is already the zero value
	// Aperture default is 0 (pinhole) which is already the zero value

	// Compute camera basis vectors from LookAt and Up.
	// This forms a right-handed orthonormal coordinate system:
	// w: points from LookAt back to camera (opposite of view direction)
	// u: points to the right (Up × w, perpendicular to both)
	// v: points up in camera space (w × u, perpendicular to both, adjusted by Up)
	// Note: Changing Up rotates the camera around the view axis (roll).
	w := Unit(Sub(c.Position, c.LookAt))
	u := Unit(Cross(c.Up, w))
	v := Cross(w, u)

	aspectRatio := float64(width) / float64(height)
	viewportWidth := aspectRatio * c.ViewportHeight
	// Viewport edges in world coordinates
	horizontal := SMul(u, viewportWidth)
	vertical := SMul(v, -c.ViewportHeight) // negative because image y goes down
	c.pixelXVector = SDiv(horizontal, float64(width))
	c.pixelYVector = SDiv(vertical, float64(height))
	// Upper left corner is: position - focal_length*w - horizontal/2 - vertical/2
	upperLeftCorner := c.Position.Minus(SMul(w, c.FocalLength), horizontal.Times(0.5), vertical.Times(0.5))
	c.pixel00 = upperLeftCorner.Plus(Add(c.pixelXVector, c.pixelYVector).Times(0.5)) // center of pixel (0,0)
}
