package ray

import "math"

type Material interface {
	Scatter(rIn Ray, rec HitRecord) (bool, ColorF, Ray)
}

type Lambertian struct {
	Albedo ColorF
}

func (l Lambertian) Scatter(rIn Ray, rec HitRecord) (bool, ColorF, Ray) {
	scatterDirection := Add(rec.Normal, RandomUnitVectorRng[Vec3](rIn.rng))
	// Catch degenerate scatter direction
	if NearZero(scatterDirection) {
		scatterDirection = rec.Normal
	}
	scattered := Ray{Origin: rec.Point, Direction: scatterDirection, rng: rIn.rng}
	return true, l.Albedo, scattered
}

type Metal struct {
	Albedo ColorF
	Fuzz   float64
}

func (m Metal) Scatter(rIn Ray, rec HitRecord) (bool, ColorF, Ray) {
	reflected := Reflect(Unit(rIn.Direction), rec.Normal)
	if m.Fuzz > 0.0 {
		reflected = Add(reflected, SMul(RandomUnitVectorRng[Vec3](rIn.rng), m.Fuzz))
	}
	scattered := Ray{Origin: rec.Point, Direction: reflected, rng: rIn.rng}
	if Dot(scattered.Direction, rec.Normal) > 0 {
		return true, m.Albedo, scattered
	}
	return false, ColorF{}, Ray{}
}

type Dielectric struct {
	RefIdx float64
}

func (d Dielectric) Scatter(rIn Ray, rec HitRecord) (bool, ColorF, Ray) {
	attenuation := ColorF{1.0, 1.0, 1.0}
	var refractionRatio float64
	if rec.FrontFace {
		refractionRatio = 1.0 / d.RefIdx
	} else {
		refractionRatio = d.RefIdx
	}
	unitDirection := Unit(rIn.Direction)
	cosTheta := math.Min(Dot(Neg(unitDirection), rec.Normal), 1.0)
	sinTheta := math.Sqrt(1.0 - cosTheta*cosTheta)
	cannotRefract := (refractionRatio*sinTheta > 1.0)
	var direction Vec3
	if cannotRefract || Reflectance(cosTheta, refractionRatio) > rIn.rng.Float64() {
		direction = Reflect(unitDirection, rec.Normal)
	} else {
		direction = Refract(unitDirection, rec.Normal, refractionRatio)
	}
	scattered := Ray{Origin: rec.Point, Direction: direction, rng: rIn.rng}
	return true, attenuation, scattered
}

func Reflectance(cosine, refIdx float64) float64 {
	// Use Schlick's approximation for reflectance.
	r0 := (1 - refIdx) / (1 + refIdx)
	r0 *= r0
	return r0 + (1-r0)*math.Pow((1-cosine), 5)
}
