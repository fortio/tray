package ray

import (
	"math"
	"math/rand/v2"
)

// Rand wraps a random number generator, is meant to be embedded in other structs and
// reused during rendering but not shared across goroutines.
type Rand struct {
	rng *rand.Rand
}

func NewRandomSource() Rand {
	//nolint:gosec // not crypto use.
	return Rand{rng: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))}
}

func NewRand(seed uint64) Rand {
	//nolint:gosec // not crypto use.
	return Rand{rng: rand.New(rand.NewPCG(0, seed))}
}

func (r Rand) Float64() float64 {
	return r.rng.Float64()
}

// Random generates a random vector with each component in [0,1).
func Random[T ~[3]float64](r Rand) T {
	return T{r.rng.Float64(), r.rng.Float64(), r.rng.Float64()}
}

// RandomInRange generates a random vector with each component in the Interval
// excluding the end.
//

func RandomInRange[T ~[3]float64](r Rand, intv Interval) T {
	minV := intv.Start
	l := intv.Length()
	return T{
		minV + l*r.rng.Float64(),
		minV + l*r.rng.Float64(),
		minV + l*r.rng.Float64(),
	}
}

// RandomUnitVectorRej generates a random unit vector using rejection sampling.
// It repeatedly samples random vectors in the cube [-1,1)^3 until one is
// found inside the unit sphere, then normalizes it to length 1.
// This is the slowest of the three methods provided here.
func RandomUnitVectorRej[T ~[3]float64](r Rand) T {
	for {
		r := RandomInRange[T](r, Interval{Start: -1, End: 1})
		lensq := LengthSquared(r)
		if lensq > 1e-48 && lensq <= 1 {
			return SDiv(r, math.Sqrt(lensq))
		}
	}
}

// RandomUnitVectorAngle generates a random unit vector using spherical coordinates.
// This method is faster than rejection sampling but involves trigonometric functions.
//

func RandomUnitVectorAngle[T ~[3]float64](r Rand) T {
	angle := r.rng.Float64() * 2 * math.Pi
	z := r.rng.Float64()*2 - 1 // in [-1,1)
	radius := math.Sqrt(1 - z*z)
	x := radius * math.Cos(angle)
	y := radius * math.Sin(angle)
	return T{x, y, z}
}

// RandomUnitVector generates a random unit vector using normal distribution.
// It is the fastest of the three methods provided here and produces uniformly
// distributed points on the unit sphere. Being both correct and most efficient,
// this is the preferred method for generating random unit vectors and thus gets
// the default name.
//

func RandomUnitVector[T ~[3]float64](r Rand) T {
	for {
		x, y, z := r.rng.NormFloat64(), r.rng.NormFloat64(), r.rng.NormFloat64()
		radius := math.Sqrt(x*x + y*y + z*z)
		if radius > 1e-24 {
			return T{x / radius, y / radius, z / radius}
		}
	}
}

// RandomOnHemisphere returns a random unit vector on the hemisphere oriented by the given normal.
func RandomOnHemisphere[T ~[3]float64](r Rand, normal T) T {
	onUnitSphere := RandomUnitVector[T](r)
	if Dot(onUnitSphere, normal) > 0.0 { // In the same hemisphere as the normal
		return onUnitSphere
	}
	return Neg(onUnitSphere)
}

// SampleDisc returns a random point (x,y) within a disc of radius r
// using the provided random source (and currently implemented via rejection sampling).
func (r Rand) SampleDisc(radius float64) (x, y float64) {
	for {
		x = 2*r.rng.Float64() - 1.0
		y = 2*r.rng.Float64() - 1.0
		if x*x+y*y <= 1 {
			break
		}
	}
	return radius * x, radius * y
}

// SampleDiscAngle returns a random point (x,y) within a disc of radius r.
// Angle method.
//

func (r Rand) SampleDiscAngle(radius float64) (x, y float64) {
	theta := 2.0 * math.Pi * r.rng.Float64()
	rad := radius * math.Sqrt(r.rng.Float64())
	x = rad * math.Cos(theta)
	y = rad * math.Sin(theta)
	return x, y
}
