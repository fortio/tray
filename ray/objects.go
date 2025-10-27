package ray

import "math"

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
}

type Scene struct {
	Objects    []Hittable
	Background AmbientLight
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
}

func (s *Sphere) Hit(r *Ray, i Interval, hr *HitRecord) bool {
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
			&Sphere{Center: Vec3{0, 0, -1.2}, Radius: 0.5, Mat: center},
			&Sphere{Center: Vec3{0, -100.5, -1}, Radius: 100, Mat: ground},
			&Sphere{Center: Vec3{-1.0, 0, -1}, Radius: 0.5, Mat: left},
			&Sphere{Center: Vec3{-1.0, 0, -1}, Radius: 0.4, Mat: bubble},
			&Sphere{Center: Vec3{1.0, 0, -1}, Radius: 0.5, Mat: right},
		},
		Background: DefaultBackground(),
	}
}

func RichScene(rand Rand) *Scene {
	ground := Lambertian{Albedo: ColorF{0.5, 0.5, 0.5}}
	world := &Scene{}
	world.Objects = append(world.Objects, &Sphere{Center: Vec3{0, -1000, 0}, Radius: 1000, Mat: ground})

	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			chooseMat := rand.Float64()
			center := Vec3{float64(a) + 0.9*rand.Float64(), 0.2, float64(b) + 0.9*rand.Float64()}

			if Length(center.Minus(XYZ(4, 0.2, 0))) > 0.9 {
				var sphereMaterial Material
				switch {
				case chooseMat < 0.8:
					// diffuse
					albedo := Mul(Random(rand), Random(rand))
					sphereMaterial = Lambertian{Albedo: albedo}
					world.Objects = append(world.Objects, &Sphere{Center: center, Radius: 0.2, Mat: sphereMaterial})
				case chooseMat < 0.95:
					// metal
					albedo := RandomInRange(rand, Interval{0.5, 1.0})
					fuzz := rand.Float64() * 0.5
					sphereMaterial = Metal{Albedo: albedo, Fuzz: fuzz}
					world.Objects = append(world.Objects, &Sphere{Center: center, Radius: 0.2, Mat: sphereMaterial})
				default:
					// glass
					sphereMaterial = Dielectric{RefIdx: 1.5}
					world.Objects = append(world.Objects, &Sphere{Center: center, Radius: 0.2, Mat: sphereMaterial})
				}
			}
		}
	}

	material1 := Dielectric{RefIdx: 1.5}
	world.Objects = append(world.Objects, &Sphere{Center: Vec3{0, 1, 0}, Radius: 1.0, Mat: material1})

	material2 := Lambertian{Albedo: ColorF{0.4, 0.2, 0.1}}
	world.Objects = append(world.Objects, &Sphere{Center: Vec3{-4, 1, 0}, Radius: 1.0, Mat: material2})

	material3 := Metal{Albedo: ColorF{0.7, 0.6, 0.5}, Fuzz: 0.0}
	world.Objects = append(world.Objects, &Sphere{Center: Vec3{4, 1, 0}, Radius: 1.0, Mat: material3})

	return world
}
