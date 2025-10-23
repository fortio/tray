package ray

import (
	"image/color"
	"math"
)

// Vec3 represents a 3D vector.
type Vec3 [3]float64

// Add: vector addition.
func (v Vec3) Add(u Vec3) Vec3 {
	return Vec3{v[0] + u[0], v[1] + u[1], v[2] + u[2]}
}

func (v Vec3) Sub(u Vec3) Vec3 {
	return Vec3{v[0] - u[0], v[1] - u[1], v[2] - u[2]}
}

// SMul: multiply by scalar.
func (v Vec3) SMul(t float64) Vec3 {
	return Vec3{v[0] * t, v[1] * t, v[2] * t}
}

func (v Vec3) SDiv(t float64) Vec3 {
	return Vec3{v[0] / t, v[1] / t, v[2] / t}
}

func (v Vec3) Length() float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
}

func (v Vec3) Unit() Vec3 {
	l := v.Length()
	return Vec3{v[0] / l, v[1] / l, v[2] / l}
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

// FloatColor for when using vector as RGB floating color.
type FloatColor = Vec3

func ColorF(r, g, b float64) FloatColor {
	return FloatColor{r, g, b}
}

func (c FloatColor) ToRGBA() color.RGBA {
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
