package ray

import (
	"image/color"
	"math"

	"fortio.org/terminal/ansipixels/tcolor"
)

// Vec3 represents a 3D vector.
// Many of the functions are implemented via generics and thus not methods.
type Vec3 struct {
	x, y, z float64
}

// ColorF is a RGB color with float components.
// Sadly go generics on [3]float64 has a huge negative performance impact and I don't
// want to copy-pasta the same implementation twice, and can't use generics on structs.
// So... type safety is no more.
type ColorF = Vec3

// Methods for both Vec3 and ColorF

// Add: vector addition. returns u + v.
func Add(u, v Vec3) Vec3 {
	return Vec3{v.x + u.x, v.y + u.y, v.z + u.z}
}

// Sub: vector subtraction, returns u - v.
func Sub(u, v Vec3) Vec3 {
	return Vec3{u.x - v.x, u.y - v.y, u.z - v.z}
}

// AddMultiple: sums all the input vectors.
func AddMultiple(u Vec3, vs ...Vec3) Vec3 {
	for _, v := range vs {
		u = Add(u, v)
	}
	return u
}

// SubMultiple: subtracts all the other input vectors from u.
// returns u - v0 - v1 - ...
func SubMultiple(u Vec3, v0 Vec3, vs ...Vec3) Vec3 {
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
func Dot(u, v Vec3) float64 {
	return u.x*v.x + u.y*v.y + u.z*v.z
}

// Cross computes the cross product of two vectors.
// The result is a vector perpendicular to both u and v, with magnitude equal to
// the area of the parallelogram formed by u and v. The direction follows the
// right-hand rule: point fingers along u, curl them toward v, thumb points along u×v.
// Common uses:
//   - Finding perpendicular vectors (e.g., camera right = up × forward)
//   - Computing surface normals from two edge vectors
//   - Determining rotation axis between two vectors
func Cross(u, v Vec3) Vec3 {
	return Vec3{u.y*v.z - u.z*v.y, u.z*v.x - u.x*v.z, u.x*v.y - u.y*v.x}
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
func SMul(v Vec3, t float64) Vec3 {
	return Vec3{v.x * t, v.y * t, v.z * t}
}

// Mul: component-wise multiplication. returns u * v.
func Mul(u, v Vec3) Vec3 {
	return Vec3{u.x * v.x, u.y * v.y, u.z * v.z}
}

// SDiv: divide by scalar.
func SDiv(v Vec3, t float64) Vec3 {
	return Vec3{v.x / t, v.y / t, v.z / t}
}

// Length: returns the length of the vector.
func Length(v Vec3) float64 {
	return math.Sqrt(LengthSquared(v))
}

// LengthSquared: returns the squared length of the vector.
func LengthSquared(v Vec3) float64 {
	return v.x*v.x + v.y*v.y + v.z*v.z
}

// Unit: returns the unit vector in the direction of v
// (normalized to length 1).
func Unit(v Vec3) Vec3 {
	l := Length(v)
	return Vec3{v.x / l, v.y / l, v.z / l}
}

// Neg: returns the negation of the vector.
func Neg(v Vec3) Vec3 {
	return Vec3{-v.x, -v.y, -v.z}
}

// NearZero returns true if the vector is close to zero in all dimensions.
func NearZero(v Vec3) bool {
	s := 1e-8
	return (math.Abs(v.x) < s) && (math.Abs(v.y) < s) && (math.Abs(v.z) < s)
}

// Reflect returns the reflection of vector v around normal n.
func Reflect(v, n Vec3) Vec3 {
	return Sub(v, SMul(n, 2*Dot(v, n)))
}

// Refract computes the refraction of vector uv through normal n
// with the given ratio of indices of refraction etaiOverEtat.
func Refract(uv, n Vec3, etaiOverEtat float64) Vec3 {
	cosTheta := math.Min(Dot(Neg(uv), n), 1.0)
	rOutPerp := SMul(Add(uv, SMul(n, cosTheta)), etaiOverEtat)
	rOutParallel := SMul(n, -math.Sqrt(math.Abs(1.0-LengthSquared(rOutPerp))))
	return Add(rOutPerp, rOutParallel)
}

// X: returns the X component.
func (v Vec3) X() float64 {
	return v.x
}

// Y: returns the Y component.
func (v Vec3) Y() float64 {
	return v.y
}

// Z: returns the Z component.
func (v Vec3) Z() float64 {
	return v.z
}

// Components returns the vector components as an array for iteration.
func (v Vec3) Components() [3]float64 {
	return [3]float64{v.x, v.y, v.z}
}

// XYZ: creates a Vec3 from its components.
func XYZ(x, y, z float64) Vec3 {
	return Vec3{x, y, z}
}

// ToSRGBA converts a linear ColorF to sRGB color.RGBA, clamping values to [0,1].
func (c ColorF) ToSRGBA() color.RGBA {
	return color.RGBA{
		R: tcolor.LinearToSrgb(c.x),
		G: tcolor.LinearToSrgb(c.y),
		B: tcolor.LinearToSrgb(c.z),
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
