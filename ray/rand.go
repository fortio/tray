package ray

import (
	"math"

	"fortio.org/rand"
)

// NewVec3 creates a Vec3 from three float64 values.
func NewVec3(x, y, z float64) Vec3 {
	return Vec3{x, y, z}
}

// Random generates a random vector with each component in [0,1).
func Random(r rand.Rand) Vec3 {
	x, y, z := r.Random3()
	return Vec3{x, y, z}
}

// RandomInRange generates a random vector with each component in the Interval [Start, End).
func RandomInRange(r rand.Rand, intv Interval) Vec3 {
	x := r.RandomInRange(intv.Start, intv.End)
	y := r.RandomInRange(intv.Start, intv.End)
	z := r.RandomInRange(intv.Start, intv.End)
	return Vec3{x, y, z}
}

// RandomUnitVector generates a random unit vector using the shared rand package.
// Returns a Vec3 instead of three separate floats.
func RandomUnitVector(r rand.Rand) Vec3 {
	x, y, z := r.RandomUnitVector()
	return Vec3{x, y, z}
}

// RandomOnHemisphere returns a random unit vector on the hemisphere oriented by the given normal.
func RandomOnHemisphere(r rand.Rand, normal Vec3) Vec3 {
	onUnitSphere := RandomUnitVector(r)
	if Dot(onUnitSphere, normal) > 0.0 { // In the same hemisphere as the normal
		return onUnitSphere
	}
	return Neg(onUnitSphere)
}

// The following functions are kept for backward compatibility with existing tests
// that compare different random unit vector generation methods.

// RandomUnitVectorRej generates a random unit vector using rejection sampling.
// It repeatedly samples random vectors in the cube [-1,1)^3 until one is
// found inside the unit sphere, then normalizes it to length 1.
// This is the slowest of the three methods provided here.
func RandomUnitVectorRej(r rand.Rand) Vec3 {
	for {
		v := RandomInRange(r, Interval{Start: -1, End: 1})
		lensq := LengthSquared(v)
		if lensq > 1e-48 && lensq <= 1 {
			return SDiv(v, math.Sqrt(lensq))
		}
	}
}

// RandomUnitVectorAngle generates a random unit vector using spherical coordinates.
// This method is faster than rejection sampling but involves trigonometric functions.
func RandomUnitVectorAngle(r rand.Rand) Vec3 {
	angle := r.Float64() * 2 * math.Pi
	z := r.Float64()*2 - 1 // in [-1,1)
	radius := math.Sqrt(1 - z*z)
	x := radius * math.Cos(angle)
	y := radius * math.Sin(angle)
	return Vec3{x, y, z}
}
