package ray

import (
	"image/color"
	"math"
	"math/rand/v2"

	"fortio.org/terminal/ansipixels/tcolor"
)

// Vec3 represents a 3D vector.
// Many of the functions are implemented via generics and thus not methods.
type Vec3 [3]float64

// ColorF is a RGB color with float components.
type ColorF [3]float64

// Methods for both Vec3 and ColorF

// Add: vector addition. returns u + v.
func Add[T ~[3]float64](u, v T) T {
	return T{v[0] + u[0], v[1] + u[1], v[2] + u[2]}
}

// Sub: vector subtraction, returns u - v.
func Sub[T ~[3]float64](u, v T) T {
	return T{u[0] - v[0], u[1] - v[1], u[2] - v[2]}
}

// AddMultiple: sums all the input vectors.
func AddMultiple[T ~[3]float64](u T, vs ...T) T {
	for _, v := range vs {
		u = Add(u, v)
	}
	return u
}

// SubMultiple: subtracts all the other input vectors from u.
// returns u - v0 - v1 - ...
func SubMultiple[T ~[3]float64](u T, v0 T, vs ...T) T {
	toSub := AddMultiple(v0, vs...)
	return Sub(u, toSub)
}

// Minus subtracts one or more vectors from v.
// Returns v - u0 - more[0] - more[1] - ...
// This is a convenience method wrapper around SubMultiple.
// Example: camera.Minus(offset1, offset2, offset3).
func (v Vec3) Minus(u0 Vec3, more ...Vec3) Vec3 {
	return SubMultiple(v, u0, more...)
}

// Dot: dot product of two vectors.
func Dot[T ~[3]float64](u, v T) float64 {
	return u[0]*v[0] + u[1]*v[1] + u[2]*v[2]
}

// Plus adds one or more vectors to v.
// Returns v + others[0] + others[1] + ...
// This is a convenience method wrapper around AddMultiple.
// Example: position.Plus(velocity, acceleration).
func (v Vec3) Plus(others ...Vec3) Vec3 {
	return AddMultiple(v, others...)
}

// Times multiplies vector v by scalar t.
// Returns v * t.
// This is a convenience method wrapper around SMul.
// Example: direction.Times(distance).
func (v Vec3) Times(t float64) Vec3 {
	return SMul(v, t)
}

// SMul: multiply by scalar.
func SMul[T ~[3]float64](v T, t float64) T {
	return T{v[0] * t, v[1] * t, v[2] * t}
}

// Mul: component-wise multiplication. returns u * v.
func Mul[T ~[3]float64](u, v T) T {
	return T{u[0] * v[0], u[1] * v[1], u[2] * v[2]}
}

// SDiv: divide by scalar.
func SDiv[T ~[3]float64](v T, t float64) T {
	return T{v[0] / t, v[1] / t, v[2] / t}
}

// Length: returns the length of the vector.
func Length[T ~[3]float64](v T) float64 {
	return math.Sqrt(LengthSquared(v))
}

// LengthSquared: returns the squared length of the vector.
func LengthSquared[T ~[3]float64](v T) float64 {
	return v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
}

// Unit: returns the unit vector in the direction of v
// (normalized to length 1).
func Unit[T ~[3]float64](v T) T {
	l := Length(v)
	return T{v[0] / l, v[1] / l, v[2] / l}
}

// Neg: returns the negation of the vector.
func Neg[T ~[3]float64](v T) T {
	return T{-v[0], -v[1], -v[2]}
}

// Random generates a random vector with each component in [0,1).
func Random[T ~[3]float64]() T {
	return T{rand.Float64(), rand.Float64(), rand.Float64()} //nolint:gosec // not crypto use.
}

// NearZero returns true if the vector is close to zero in all dimensions.
func NearZero[T ~[3]float64](v T) bool {
	s := 1e-8
	return (math.Abs(v[0]) < s) && (math.Abs(v[1]) < s) && (math.Abs(v[2]) < s)
}

// Reflect returns the reflection of vector v around normal n.
func Reflect[T ~[3]float64](v, n T) T {
	return Sub(v, SMul(n, 2*Dot(v, n)))
}

// Refract computes the refraction of vector uv through normal n
// with the given ratio of indices of refraction etaiOverEtat.
func Refract[T ~[3]float64](uv, n T, etaiOverEtat float64) T {
	cosTheta := math.Min(Dot(Neg(uv), n), 1.0)
	rOutPerp := SMul(Add(uv, SMul(n, cosTheta)), etaiOverEtat)
	rOutParallel := SMul(n, -math.Sqrt(math.Abs(1.0-LengthSquared(rOutPerp))))
	return Add(rOutPerp, rOutParallel)
}

// RandomInRange generates a random vector with each component in the Interval
// excluding the end.
//
//nolint:gosec // not crypto use.
func RandomInRange[T ~[3]float64](intv Interval) T {
	minV := intv.Start
	l := intv.Length()
	return T{
		minV + l*rand.Float64(),
		minV + l*rand.Float64(),
		minV + l*rand.Float64(),
	}
}

// RandomUnitVectorRej generates a random unit vector using rejection sampling.
// It repeatedly samples random vectors in the cube [-1,1)^3 until one is
// found inside the unit sphere, then normalizes it to length 1.
// This is the slowest of the three methods provided here.
func RandomUnitVectorRej[T ~[3]float64]() T {
	for {
		r := RandomInRange[T](Interval{Start: -1, End: 1})
		lensq := LengthSquared(r)
		if lensq > 1e-48 && lensq <= 1 {
			return SDiv(r, math.Sqrt(lensq))
		}
	}
}

// RandomUnitVectorAngle generates a random unit vector using spherical coordinates.
// This method is faster than rejection sampling but involves trigonometric functions.
//
//nolint:gosec // not crypto use.
func RandomUnitVectorAngle[T ~[3]float64]() T {
	angle := rand.Float64() * 2 * math.Pi
	z := rand.Float64()*2 - 1 // in [-1,1)
	r := math.Sqrt(1 - z*z)
	x := r * math.Cos(angle)
	y := r * math.Sin(angle)
	return T{x, y, z}
}

// RandomUnitVector generates a random unit vector using normal distribution.
// It is the fastest of the three methods provided here and produces uniformly
// distributed points on the unit sphere. Being both correct and most efficient,
// this is the preferred method for generating random unit vectors and thus gets
// the default name.
//
//nolint:gosec // not crypto use.
func RandomUnitVector[T ~[3]float64]() T {
	for {
		x, y, z := rand.NormFloat64(), rand.NormFloat64(), rand.NormFloat64()
		r := math.Sqrt(x*x + y*y + z*z)
		if r > 1e-24 {
			return T{x / r, y / r, z / r}
		}
	}
}

// RandomUnitVectorRng generates a random unit vector using normal distribution
// with the provided random source. This version allows per-goroutine rand sources.
func RandomUnitVectorRng[T ~[3]float64](rng *rand.Rand) T {
	for {
		x, y, z := rng.NormFloat64(), rng.NormFloat64(), rng.NormFloat64()
		r := math.Sqrt(x*x + y*y + z*z)
		if r > 1e-24 {
			return T{x / r, y / r, z / r}
		}
	}
}

// RandomOnHemisphere returns a random unit vector on the hemisphere oriented by the given normal.
func RandomOnHemisphere[T ~[3]float64](normal T) T {
	onUnitSphere := RandomUnitVector[T]()
	if Dot(onUnitSphere, normal) > 0.0 { // In the same hemisphere as the normal
		return onUnitSphere
	}
	return Neg(onUnitSphere)
}

// X: returns the X component.
func (v Vec3) X() float64 {
	return v[0]
}

// Y: returns the Y component.
func (v Vec3) Y() float64 {
	return v[1]
}

// Z: returns the Z component.
func (v Vec3) Z() float64 {
	return v[2]
}

// XYZ: creates a Vec3 from its components.
func XYZ(x, y, z float64) Vec3 {
	return Vec3{x, y, z}
}

// ToSRGBA converts a linear ColorF to sRGB color.RGBA, clamping values to [0,1].
func (c ColorF) ToSRGBA() color.RGBA {
	return color.RGBA{
		R: tcolor.LinearToSrgb(c[0]),
		G: tcolor.LinearToSrgb(c[1]),
		B: tcolor.LinearToSrgb(c[2]),
		A: 255,
	}
}

// ToRGBALinear converts a linear ColorF to linear color.RGBA, values must be in [0,1].
func (c ColorF) ToRGBALinear() color.RGBA {
	return color.RGBA{
		R: uint8(math.Round(255. * float64(c[0]))),
		G: uint8(math.Round(255. * float64(c[1]))),
		B: uint8(math.Round(255. * float64(c[2]))),
		A: 255,
	}
}

// Interval represents a closed interval [Start, End] on the real number line.
type Interval struct {
	Start, End float64
}

// Length returns the length of the interval (End - Start).
func (i Interval) Length() float64 {
	return i.End - i.Start
}

// Contains returns true if t is within the interval [Start, End] (inclusive).
func (i Interval) Contains(t float64) bool {
	return t >= i.Start && t <= i.End
}

// Surrounds returns true if t is strictly within the interval (Start, End) (exclusive).
func (i Interval) Surrounds(t float64) bool {
	return t > i.Start && t < i.End
}

// Clamp returns t clamped to the interval [Start, End].
// If t < Start, returns Start. If t > End, returns End. Otherwise returns t.
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
	Empty        = Interval{Start: math.Inf(1), End: math.Inf(-1)}
	Universe     = Interval{Start: math.Inf(-1), End: math.Inf(1)}
	Front        = Interval{Start: 0, End: math.Inf(1)}
	FrontEpsilon = Interval{Start: 1e-6, End: math.Inf(1)}
	ZeroOne      = Interval{Start: 0, End: 1}
)
