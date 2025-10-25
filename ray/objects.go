package ray

import "math"

type HitRecord struct {
	Point     Vec3
	Normal    Vec3
	T         float64
	FrontFace bool
	Mat       Material
}

func (hr *HitRecord) SetFaceNormal(r Ray, outwardNormal Vec3) {
	hr.FrontFace = Dot(r.Direction, outwardNormal) < 0
	if hr.FrontFace {
		hr.Normal = outwardNormal
	} else {
		hr.Normal = Neg(outwardNormal)
	}
}

type Hittable interface {
	Hit(r Ray, interval Interval) (bool, HitRecord)
}

type Scene struct {
	Objects []Hittable
}

func (s *Scene) Hit(r Ray, interval Interval) (bool, HitRecord) {
	hitAnything := false
	closestSoFar := interval.End
	var tempRec HitRecord

	for _, object := range s.Objects {
		if hit, rec := object.Hit(r, Interval{Start: interval.Start, End: closestSoFar}); hit {
			hitAnything = true
			closestSoFar = rec.T
			tempRec = rec
		}
	}
	return hitAnything, tempRec
}

// RayColor is the main function for computing the color of a ray (thus a pixel).
func (s *Scene) RayColor(r Ray, depth int) ColorF {
	if depth <= 0 {
		return ColorF{0, 0, 0}
	}
	if hit, hr := s.Hit(r, FrontEpsilon); hit {
		var scattered Ray
		var attenuation ColorF
		if didScatter, att, scat := hr.Mat.Scatter(r, hr); didScatter {
			attenuation = att
			scattered = scat
			return Mul(attenuation, s.RayColor(scattered, depth-1))
		}
		return ColorF{0, 0, 0}
	}
	unit := Unit(r.Direction)
	a := 0.5 * (unit.Y() + 1.0)
	white := ColorF{1.0, 1.0, 1.0}
	blue := ColorF{0.4, 0.65, 1.0}
	blend := Add(SMul(white, 1.0-a), SMul(blue, a))
	return blend
}

type Sphere struct {
	Center Vec3
	Radius float64
	Mat    Material
}

func (s *Sphere) Hit(r Ray, i Interval) (bool, HitRecord) {
	oc := Sub(s.Center, r.Origin)
	a := LengthSquared(r.Direction)
	h := Dot(r.Direction, oc)
	c := LengthSquared(oc) - s.Radius*s.Radius
	discriminant := h*h - a*c
	if discriminant < 0 {
		return false, HitRecord{}
	}
	sqrtD := math.Sqrt(discriminant)
	root := (h - sqrtD) / a
	if !i.Surrounds(root) {
		root = (h + sqrtD) / a
		if !i.Surrounds(root) {
			return false, HitRecord{}
		}
	}
	hr := HitRecord{Point: r.At(root), T: root}
	outwardNormal := SDiv(Sub(hr.Point, s.Center), s.Radius)
	hr.SetFaceNormal(r, outwardNormal)
	hr.Mat = s.Mat
	return true, hr
}

func DefaultScene() *Scene {
	ground := Lambertian{Albedo: ColorF{0.7, 0.8, 0.1}}
	center := Lambertian{Albedo: ColorF{0.1, 0.2, 0.5}}
	//		left := Metal{Albedo: ColorF{0.8, 0.8, 0.8}, Fuzz: 0}
	left := Dielectric{1.5}
	bubble := Dielectric{1.0 / 1.5}
	right := Metal{Albedo: ColorF{1, .8, .8}, Fuzz: 0.05}
	return &Scene{
		// Default scene with two spheres.
		Objects: []Hittable{
			&Sphere{Center: Vec3{0, 0, -1.2}, Radius: 0.5, Mat: center},
			&Sphere{Center: Vec3{0, -100.5, -1}, Radius: 100, Mat: ground},
			&Sphere{Center: Vec3{-1.0, 0, -1}, Radius: 0.5, Mat: left},
			&Sphere{Center: Vec3{-1.0, 0, -1}, Radius: 0.4, Mat: bubble},
			&Sphere{Center: Vec3{1.0, 0, -1}, Radius: 0.5, Mat: right},
		},
	}
}
