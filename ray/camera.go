package ray

import "math"

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
	// VerticalFoV is the vertical field of view in degrees.
	// Typical values: 40-60° for normal view, 90° for wide angle, 20° for telephoto.
	// If zero, defaults to 90°.
	VerticalFoV float64
	// FocalLength is the distance from the camera to the image plane.
	// If zero, defaults to 1.0. Usually you don't need to change this.
	FocalLength float64
	// FocusDistance is the distance from the camera to the plane that will be in sharp focus.
	// Objects at this distance appear sharp; closer/farther objects are blurred based on Aperture.
	// If zero, defaults to FocalLength.
	FocusDistance float64
	// Aperture is the diameter of the camera's aperture. Zero means pinhole (no blur).
	// Larger aperture = more blur for out-of-focus objects (shallower depth of field).
	Aperture float64
	// Computed fields (initialized by Initialize)
	pixel00      Vec3
	pixelXVector Vec3
	pixelYVector Vec3
	defocusDiskU Vec3 // basis vector for lens disk (right)
	defocusDiskV Vec3 // basis vector for lens disk (up)
}

// Initialize computes the viewport parameters for the given image dimensions.
// Sets default values for any zero-valued fields. Must be called before rendering.
func (c *Camera) Initialize(width, height int) {
	var zero Vec3
	// Set defaults for zero-valued fields
	if c.FocalLength == 0 {
		c.FocalLength = 1.0
	}
	if c.VerticalFoV == 0 {
		c.VerticalFoV = 90.0 // Default to 90 degree field of view
	}
	if c.Up == zero {
		c.Up = Vec3{0, 1, 0}
	}
	if c.FocusDistance == 0 {
		c.FocusDistance = c.FocalLength
	}
	// If both Position and LookAt are at origin, set LookAt to look down -Z
	if c.Position == zero && c.LookAt == zero {
		c.LookAt = Vec3{0, 0, -1}
	}
	// Position default is (0,0,0) which is already the zero value
	// Aperture default is 0 (pinhole) which is already the zero value

	// Validate that Position and LookAt are different
	viewDirection := Sub(c.Position, c.LookAt)
	if NearZero(viewDirection) {
		// Position == LookAt, can't determine view direction.
		// Default to looking down -Z axis (w points toward +Z).
		viewDirection = Vec3{0, 0, 1}
	}

	// Compute camera basis vectors from LookAt and Up.
	// This forms a right-handed orthonormal coordinate system:
	// w: points from LookAt back to camera (opposite of view direction)
	// u: points to the right (Up × w, perpendicular to both)
	// v: points up in camera space (w × u, perpendicular to both, adjusted by Up)
	// Note: Changing Up rotates the camera around the view axis (roll).
	w := Unit(viewDirection)
	u := Unit(Cross(c.Up, w))
	v := Cross(w, u)

	// Compute defocus disk basis vectors for depth of field
	// The disk radius is aperture/2, and these vectors define the disk's orientation
	defocusRadius := c.Aperture / 2
	c.defocusDiskU = SMul(u, defocusRadius)
	c.defocusDiskV = SMul(v, defocusRadius)

	// Compute viewport dimensions from field of view
	// tan(fov/2) = (viewportHeight/2) / focalLength
	// viewportHeight = 2 * focalLength * tan(fov/2)
	theta := c.VerticalFoV * (math.Pi / 180.0) // degrees to radians
	viewportHeight := 2.0 * c.FocalLength * math.Tan(theta/2.0)
	aspectRatio := float64(width) / float64(height)
	viewportWidth := aspectRatio * viewportHeight

	// Viewport edges in world coordinates
	horizontal := SMul(u, viewportWidth)
	vertical := SMul(v, -viewportHeight) // negative because image y goes down
	c.pixelXVector = SDiv(horizontal, float64(width))
	c.pixelYVector = SDiv(vertical, float64(height))
	// Upper left corner of viewport
	upperLeftCorner := c.Position.Minus(SMul(w, c.FocalLength), horizontal.Times(0.5), vertical.Times(0.5))
	// pixel00 is the upper-left corner - offsets are added to get to pixel centers
	c.pixel00 = upperLeftCorner
}

// GetRay generates a ray from the camera through the specified pixel coordinates,
// with optional depth of field blur if Aperture > 0.
// The offsets (offsetX, offsetY) allow for sub-pixel sampling:
//   - (0, 0) = pixel center
//   - (-0.5, -0.5) = upper-left corner
//   - (0.5, 0.5) = lower-right corner
func (c *Camera) GetRay(rng Rand, pixelX, pixelY, offsetX, offsetY float64) *Ray {
	// Compute the point on the viewport
	// offset (0,0) = pixel center, so we add 0.5 to get from upper-left corner to center
	pixelSample := c.pixel00.Plus(
		c.pixelXVector.Times(pixelX+0.5+offsetX),
		c.pixelYVector.Times(pixelY+0.5+offsetY),
	)

	// Ray from camera position through the pixel sample
	rayOrigin := c.Position
	rayDirection := Sub(pixelSample, c.Position)

	// If aperture > 0, simulate depth of field by sampling from lens disk
	if c.Aperture > 0 {
		// Sample random point on lens disk
		dx, dy := rng.SampleDisc(1.0) // Sample unit disk
		offset := Add(SMul(c.defocusDiskU, dx), SMul(c.defocusDiskV, dy))

		// Compute the focus point: where the center ray hits the focus plane
		// Focus plane is FocusDistance away from camera along view direction
		focusTime := c.FocusDistance / c.FocalLength
		focusPoint := Add(c.Position, SMul(rayDirection, focusTime))

		// Ray now originates from offset position on lens disk and aims at focus point
		rayOrigin = Add(c.Position, offset)
		rayDirection = Sub(focusPoint, rayOrigin)
	}

	return rng.NewRay(rayOrigin, rayDirection)
}
