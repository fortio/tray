package ray

import (
	"math"

	"fortio.org/rand"
)

// HitRecord holds information about a ray-object intersection.
// Note: it's too big to be returned by value efficiently.
type HitRecord struct {
	Point     Vec3
	Normal    Vec3
	T         float64
	Mat       Material
	FrontFace bool
}

func (hr *HitRecord) SetFaceNormal(r *Ray, outwardNormal Vec3) {
	hr.FrontFace = Dot(r.Direction, outwardNormal) < 0
	if hr.FrontFace {
		hr.Normal = outwardNormal
	} else {
		hr.Normal = Neg(outwardNormal)
	}
}

type Hittable interface {
	Hit(r *Ray, interval Interval, hr *HitRecord) bool
	BoundingBox() AABB
}

type Scene struct {
	Objects    []Hittable
	Background AmbientLight
	BBox       AABB
}

func (s *Scene) Hit(r *Ray, interval Interval, hr *HitRecord) (hitAnything bool) {
	closestSoFar := interval.End
	for _, object := range s.Objects {
		if hit := object.Hit(r, Interval{Start: interval.Start, End: closestSoFar}, hr); hit {
			hitAnything = true
			closestSoFar = hr.T
		}
	}
	return hitAnything
}

// RayColor is the main function for computing the color of a ray (thus a pixel).
func (s *Scene) RayColor(r *Ray, depth int) ColorF {
	if depth <= 0 {
		return ColorF{0, 0, 0}
	}
	hr := &HitRecord{}
	if hit := s.Hit(r, FrontEpsilon, hr); hit {
		if didScatter, attenuation, scattered := hr.Mat.Scatter(r, hr); didScatter {
			return Mul(attenuation, s.RayColor(scattered, depth-1))
		}
		return ColorF{0, 0, 0}
	}
	// later we can allow not having a background (put back the nil check) but for now it's the only light source
	return s.Background.Hit(r)
}

func (s *Scene) BoundingBox() AABB {
	return s.BBox
}

type AmbientLight struct {
	ColorA, ColorB ColorF
}

func (al AmbientLight) Hit(r *Ray) ColorF {
	unit := Unit(r.Direction)
	a := 0.5 * (unit.Y() + 1.0)
	blend := Add(SMul(al.ColorA, 1.0-a), SMul(al.ColorB, a))
	return blend
}

type Sphere struct {
	Center Vec3
	Radius float64
	Mat    Material
	BBox   AABB
}

func NewSphere(center Vec3, radius float64, mat Material) *Sphere {
	s := &Sphere{
		Center: center,
		Radius: radius,
		Mat:    mat,
	}
	s.BBox = s.boundingBox()
	return s
}

func (s *Sphere) boundingBox() AABB {
	r := s.Radius
	/* See the box test:
	if r < 5 {
		r *= .5
	}
	*/
	rVec := Vec3{r, r, r}
	return NewAABB(s.Center.Minus(rVec), s.Center.Plus(rVec))
}

func (s *Sphere) BoundingBox() AABB {
	return s.BBox
}

func (s *Sphere) Hit(r *Ray, i Interval, hr *HitRecord) bool {
	/*if !s.BBox.Hit(r, i) {
		return false
	}*/
	oc := Sub(s.Center, r.Origin)
	a := LengthSquared(r.Direction)
	h := Dot(r.Direction, oc)
	c := LengthSquared(oc) - s.Radius*s.Radius
	discriminant := h*h - a*c
	if discriminant < 0 {
		return false
	}
	sqrtD := math.Sqrt(discriminant)
	root := (h - sqrtD) / a
	if !i.Surrounds(root) {
		root = (h + sqrtD) / a
		if !i.Surrounds(root) {
			return false
		}
	}
	hr.Point = r.At(root)
	hr.T = root
	outwardNormal := SDiv(Sub(hr.Point, s.Center), s.Radius)
	hr.SetFaceNormal(r, outwardNormal)
	hr.Mat = s.Mat
	return true
}

func DefaultBackground() AmbientLight {
	white := ColorF{1.0, 1.0, 1.0}
	blue := ColorF{0.4, 0.65, 1.0}
	return AmbientLight{ColorA: white, ColorB: blue}
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
			NewSphere(Vec3{0, 0, -1.2}, 0.5, center),
			NewSphere(Vec3{0, -100.5, -1}, 100, ground),
			NewSphere(Vec3{-1.0, 0, -1}, 0.5, left),
			NewSphere(Vec3{-1.0, 0, -1}, 0.4, bubble),
			NewSphere(Vec3{1.0, 0, -1}, 0.5, right),
		},
		Background: DefaultBackground(),
	}
}

func RichScene(rng rand.Rand) *Scene {
	ground := Lambertian{Albedo: ColorF{0.5, 0.5, 0.5}}
	world := &Scene{}
	world.Objects = append(world.Objects, NewSphere(Vec3{0, -1000, 0}, 1000, ground))

	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			chooseMat := rng.Float64()
			center := Vec3{float64(a) + 0.9*rng.Float64(), 0.2, float64(b) + 0.9*rng.Float64()}

			if Length(center.Minus(XYZ(4, 0.2, 0))) > 0.9 {
				var sphereMaterial Material
				switch {
				case chooseMat < 0.8:
					// diffuse
					albedo := Mul(Random(rng), Random(rng))
					sphereMaterial = Lambertian{Albedo: albedo}
					world.Objects = append(world.Objects, NewSphere(center, 0.2, sphereMaterial))
				case chooseMat < 0.95:
					// metal
					albedo := RandomInRange(rng, Interval{0.5, 1.0})
					fuzz := rng.Float64() * 0.5
					sphereMaterial = Metal{Albedo: albedo, Fuzz: fuzz}
					world.Objects = append(world.Objects, NewSphere(center, 0.2, sphereMaterial))
				default:
					// glass
					sphereMaterial = Dielectric{RefIdx: 1.5}
					world.Objects = append(world.Objects, NewSphere(center, 0.2, sphereMaterial))
				}
			}
		}
	}

	material1 := Dielectric{RefIdx: 1.5}
	world.Objects = append(world.Objects, NewSphere(Vec3{0, 1, 0}, 1.0, material1))

	material2 := Lambertian{Albedo: ColorF{0.4, 0.2, 0.1}}
	world.Objects = append(world.Objects, NewSphere(Vec3{-4, 1, 0}, 1.0, material2))

	material3 := Metal{Albedo: ColorF{0.7, 0.6, 0.5}, Fuzz: 0.0}
	world.Objects = append(world.Objects, NewSphere(Vec3{4, 1, 0}, 1.0, material3))

	return world
}
