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

// Add: vector addition.
func Add[T ~[3]float64](u, v T) T {
	return T{v[0] + u[0], v[1] + u[1], v[2] + u[2]}
}

func Sub[T ~[3]float64](u, v T) T {
	return T{u[0] - v[0], u[1] - v[1], u[2] - v[2]}
}

// SMul: multiply by scalar.
func SMul[T ~[3]float64](v T, t float64) T {
	return T{v[0] * t, v[1] * t, v[2] * t}
}

func SDiv[T ~[3]float64](v T, t float64) T {
	return T{v[0] / t, v[1] / t, v[2] / t}
}

func Length[T ~[3]float64](v T) float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
}

func Unit[T ~[3]float64](v T) T {
	l := Length(v)
	return T{v[0] / l, v[1] / l, v[2] / l}
}

func (v Vec3) X() float64 {
	return v[0]
}

func (v Vec3) Y() float64 {
	return v[1]
}

func (v Vec3) Z() float64 {
	return v[2]
}

func XYZ(x, y, z float64) Vec3 {
	return Vec3{x, y, z}
}

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
