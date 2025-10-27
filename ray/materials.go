package ray

import "math"

type Material interface {
	Scatter(rIn *Ray, rec HitRecord) (bool, ColorF, *Ray)
}

type Lambertian struct {
	Albedo ColorF
}

func (l Lambertian) Scatter(rIn *Ray, rec HitRecord) (bool, ColorF, *Ray) {
	scatterDirection := rec.Normal.Add(RandomUnitVector(rIn.Rand))
	// Catch degenerate scatter direction
	if scatterDirection.NearZero() {
		scatterDirection = rec.Normal
	}
	scattered := rIn.NewRay(rec.Point, scatterDirection)
	return true, l.Albedo, scattered
}

type Metal struct {
	Albedo ColorF
	Fuzz   float64
}

func (m Metal) Scatter(rIn *Ray, rec HitRecord) (bool, ColorF, *Ray) {
	reflected := rIn.Direction.Unit().Reflect(rec.Normal)
	if m.Fuzz > 0.0 {
		reflected = reflected.Add(RandomUnitVector(rIn.Rand).SMul(m.Fuzz))
	}
	scattered := rIn.NewRay(rec.Point, reflected)
	if scattered.Direction.Dot(rec.Normal) > 0 {
		return true, m.Albedo, scattered
	}
	return false, ColorF{}, nil
}

type Dielectric struct {
	RefIdx float64
}

func (d Dielectric) Scatter(rIn *Ray, rec HitRecord) (bool, ColorF, *Ray) {
	attenuation := ColorF{Vec3{1.0, 1.0, 1.0}}
	var refractionRatio float64
	if rec.FrontFace {
		refractionRatio = 1.0 / d.RefIdx
	} else {
		refractionRatio = d.RefIdx
	}
	unitDirection := rIn.Direction.Unit()
	cosTheta := math.Min(unitDirection.Neg().Dot(rec.Normal), 1.0)
	sinTheta := math.Sqrt(1.0 - cosTheta*cosTheta)
	cannotRefract := (refractionRatio*sinTheta > 1.0)
	var direction Vec3
	if cannotRefract || Reflectance(cosTheta, refractionRatio) > rIn.Float64() {
		direction = unitDirection.Reflect(rec.Normal)
	} else {
		direction = unitDirection.Refract(rec.Normal, refractionRatio)
	}
	scattered := rIn.NewRay(rec.Point, direction)
	return true, attenuation, scattered
}

func Reflectance(cosine, refIdx float64) float64 {
	// Use Schlick's approximation for reflectance.
	r0 := (1 - refIdx) / (1 + refIdx)
	r0 *= r0
	return r0 + (1-r0)*math.Pow((1-cosine), 5)
}
