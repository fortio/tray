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
type ColorF struct{ v Vec3 }

// Methods for both Vec3 and ColorF

func (v Vec3) Add(w Vec3) Vec3 {
	return Vec3{
		v.x + w.x,
		v.y + w.y,
		v.z + w.z,
	}
}
func (c ColorF) Add(d ColorF) ColorF {
	return ColorF{c.v.Add(d.v)}
}

func (v Vec3) Sub(w Vec3) Vec3 {
	return Vec3{
		v.x - w.x,
		v.y - w.y,
		v.z - w.z,
	}
}
func (c ColorF) Sub(d ColorF) ColorF {
	return ColorF{c.v.Sub(d.v)}
}

func (v Vec3) AddMultiple(vs ...Vec3) Vec3 {
	for _, w := range vs {
		v = v.Add(w)
	}
	return v
}
func (v ColorF) AddMultiple(vs ...ColorF) ColorF {
	for _, w := range vs {
		v = v.Add(w)
	}
	return v
}

func (v Vec3) SubMultiple(vs ...Vec3) Vec3 {
	for _, w := range vs {
		v = v.Sub(w)
	}
	return v
}
func (v ColorF) SubMultiple(vs ...ColorF) ColorF {
	for _, w := range vs {
		v = v.Sub(w)
	}
	return v
}

// Minus subtracts one or more vectors from v.
// Returns v - u0 - more[0] - more[1] - ...
// This is a convenience method wrapper around SubMultiple.
// Example: camera.Minus(offset1, offset2, offset3).
func (v Vec3) Minus(u0 Vec3, more ...Vec3) Vec3 {
	return v.Sub(u0).SubMultiple(more...)
}
func (v ColorF) Minus(u0 ColorF, more ...ColorF) ColorF {
	return v.Sub(u0).SubMultiple(more...)
}

func (v Vec3) Dot(w Vec3) float64 {
	return v.x*w.x + v.y*w.y + v.z*w.z
}
func (c ColorF) Dot(w ColorF) float64 {
	return c.v.Dot(w.v)
}

// Cross computes the cross product of two vectors.
// The result is a vector perpendicular to both u and v, with magnitude equal to
// the area of the parallelogram formed by u and v. The direction follows the
// right-hand rule: point fingers along u, curl them toward v, thumb points along u×v.
// Common uses:
//   - Finding perpendicular vectors (e.g., camera right = up × forward)
//   - Computing surface normals from two edge vectors
//   - Determining rotation axis between two vectors
func (v Vec3) Cross(w Vec3) Vec3 {
	return Vec3{
		v.y*w.z - v.z*w.y,
		v.z*w.x - v.x*w.z,
		v.x*w.y - v.y*w.x,
	}
}
func (c ColorF) Cross(d ColorF) ColorF {
	return ColorF{c.v.Cross(d.v)}
}

// Plus adds one or more vectors to v.
// Returns v + others[0] + others[1] + ...
// This is a convenience method wrapper around AddMultiple.
// Example: position.Plus(velocity, acceleration).
func (v Vec3) Plus(others ...Vec3) Vec3 {
	return v.AddMultiple(others...)
}
func (c ColorF) Plus(others ...ColorF) ColorF {
	return c.AddMultiple(others...)
}

// Times multiplies vector v by scalar t.
// Returns v * t.
// This is a convenience method wrapper around SMul.
// Example: direction.Times(distance).
func (v Vec3) Times(t float64) Vec3 {
	return v.SMul(t)
}
func (c ColorF) Times(t float64) ColorF {
	return ColorF{c.v.Times(t)}
}

func (c Vec3) SMul(t float64) Vec3 {
	return Vec3{
		c.x * t,
		c.y * t,
		c.z * t,
	}
}
func (c ColorF) SMul(t float64) ColorF {
	return ColorF{c.v.SMul(t)}
}

func (v Vec3) Mul(w Vec3) Vec3 {
	return Vec3{
		v.x * w.x,
		v.y * w.y,
		v.z * w.z,
	}
}
func (c ColorF) Mul(d ColorF) ColorF {
	return ColorF{c.v.Mul(d.v)}
}

func (v Vec3) SDiv(t float64) Vec3 {
	return Vec3{
		v.x / t,
		v.y / t,
		v.z / t,
	}
}
func (c ColorF) SDiv(t float64) ColorF {
	return ColorF{c.v.SDiv(t)}
}

func (v Vec3) Length() float64 {
	return math.Sqrt(v.LengthSquared())
}
func (c ColorF) Length() float64 {
	return c.v.Length()
}

func (v Vec3) LengthSquared() float64 {
	return v.x*v.x + v.y*v.y + v.z*v.z
}
func (c ColorF) LengthSquared() float64 {
	return c.v.LengthSquared()
}

func (v Vec3) Unit() Vec3 {
	l := v.Length()
	return v.SDiv(l)
}
func (c ColorF) Unit() ColorF {
	return ColorF{c.v.Unit()}
}

func (v Vec3) Neg() Vec3 {
	return Vec3{
		-v.x,
		-v.y,
		-v.z,
	}
}
func (c ColorF) Neg() ColorF {
	return ColorF{c.v.Neg()}
}

func (v Vec3) NearZero() bool {
	s := 1e-8
	return (math.Abs(v.x) < s) && (math.Abs(v.y) < s) && (math.Abs(v.z) < s)
}
func (c ColorF) NearZero() bool {
	return c.v.NearZero()
}

func (v Vec3) Reflect(n Vec3) Vec3 {
	return v.Sub(n.SMul(2 * v.Dot(n)))
}
func (c ColorF) Reflect(n ColorF) ColorF {
	return ColorF{c.v.Reflect(n.v)}
}

func (v Vec3) Refract(n Vec3, etaiOverEtat float64) Vec3 {
	cosTheta := math.Min(v.Neg().Dot(n), 1.0)
	rOutPerp := v.Plus(n.Times(cosTheta)).Times(etaiOverEtat)
	rOutParallel := n.Times(-math.Sqrt(math.Abs(1.0 - rOutPerp.LengthSquared())))
	return rOutPerp.Plus(rOutParallel)
}
func (c ColorF) Refract(n ColorF, etaiOverEtat float64) ColorF {
	return ColorF{c.v.Refract(n.v, etaiOverEtat)}
}

// X: returns the X component.
func (v Vec3) X() float64 {
	return v.x
}
func (c ColorF) R() float64 {
	return c.v.x
}

// Y: returns the Y component.
func (v Vec3) Y() float64 {
	return v.y
}
func (c ColorF) G() float64 {
	return c.v.y
}

// Z: returns the Z component.
func (v Vec3) Z() float64 {
	return v.z
}
func (c ColorF) B() float64 {
	return c.v.z
}

// XYZ: creates a Vec3 from its components.
func XYZ(x, y, z float64) Vec3 {
	return Vec3{x, y, z}
}

// ToSRGBA converts a linear ColorF to sRGB color.RGBA, clamping values to [0,1].
func (c ColorF) ToSRGBA() color.RGBA {
	return color.RGBA{
		R: tcolor.LinearToSrgb(c.R()),
		G: tcolor.LinearToSrgb(c.G()),
		B: tcolor.LinearToSrgb(c.B()),
		A: 255,
	}
}

// ToRGBALinear converts a linear ColorF to linear color.RGBA, values must be in [0,1].
func (c ColorF) ToRGBALinear() color.RGBA {
	return color.RGBA{
		R: uint8(math.Round(255. * float64(c.R()))),
		G: uint8(math.Round(255. * float64(c.G()))),
		B: uint8(math.Round(255. * float64(c.B()))),
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
