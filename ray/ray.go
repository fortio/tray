package ray

// Ray holds information about a ray in 3D space and a reference to a random number generator
// not to be shared across goroutines.
type Ray struct {
	Rand
	Origin    Vec3
	Direction Vec3
}

// NewRay creates a new Ray with the given origin and direction, transferring
// the Rand source.
func (r Rand) NewRay(origin, direction Vec3) *Ray {
	return &Ray{
		Rand:      r,
		Origin:    origin,
		Direction: direction,
	}
}

func (r *Ray) At(t float64) Vec3 {
	return r.Origin.Add(r.Direction.SMul(t))
}
