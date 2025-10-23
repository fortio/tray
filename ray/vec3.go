package ray

import (
	"image/color"
	"math"
)

// Vec3 represents a 3D vector.
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

// SMul: multiply by scalar.
func SMul[T ~[3]float64](v T, t float64) T {
	return T{v[0] * t, v[1] * t, v[2] * t}
}

// SDiv: divide by scalar.
func SDiv[T ~[3]float64](v T, t float64) T {
	return T{v[0] / t, v[1] / t, v[2] / t}
}

// Length: returns the length of the vector.
func Length[T ~[3]float64](v T) float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
}

// Unit: returns the unit vector in the direction of v
// (normalized to length 1).
func Unit[T ~[3]float64](v T) T {
	l := Length(v)
	return T{v[0] / l, v[1] / l, v[2] / l}
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

// ToRGBA converts ColorF to color.RGBA, clamping values to [0,1].
func (c ColorF) ToRGBA() color.RGBA {
	r := uint8(clamp(c[0], 0, 1) * 255)
	g := uint8(clamp(c[1], 0, 1) * 255)
	b := uint8(clamp(c[2], 0, 1) * 255)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func clamp(x, minV, maxV float64) float64 {
	if x < minV {
		return minV
	}
	if x > maxV {
		return maxV
	}
	return x
}
