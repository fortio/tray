package ray

import (
	"math"
	"testing"
)

func TestSetFaceNormalFrontFace(t *testing.T) {
	rnd := NewRandomSource()
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, 0, -1})
	outwardNormal := Vec3{0, 0, 1}

	hr := HitRecord{}
	hr.SetFaceNormal(ray, outwardNormal)

	if !hr.FrontFace {
		t.Error("Expected front face")
	}
	if hr.Normal != outwardNormal {
		t.Errorf("Expected normal %v, got %v", outwardNormal, hr.Normal)
	}
}

func TestSetFaceNormalBackFace(t *testing.T) {
	rnd := NewRandomSource()
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, 0, 1})
	outwardNormal := Vec3{0, 0, 1}

	hr := HitRecord{}
	hr.SetFaceNormal(ray, outwardNormal)

	if hr.FrontFace {
		t.Error("Expected back face")
	}
	expected := Vec3{0, 0, -1}
	if hr.Normal != expected {
		t.Errorf("Expected normal %v, got %v", expected, hr.Normal)
	}
}

func TestSphereHitSimple(t *testing.T) {
	rnd := NewRandomSource()
	sphere := Sphere{
		Center: Vec3{0, 0, -1},
		Radius: 0.5,
		Mat:    Lambertian{Albedo: ColorF{1, 0, 0}},
	}
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, 0, -1})

	hit, rec := sphere.Hit(ray, FrontEpsilon)

	if !hit {
		t.Fatal("Expected hit")
	}
	// Check that t is reasonable
	if rec.T <= 0 {
		t.Errorf("Expected positive t, got %v", rec.T)
	}
	// Check that hit point is on sphere surface
	distFromCenter := Length(Sub(rec.Point, sphere.Center))
	if math.Abs(distFromCenter-sphere.Radius) > 1e-10 {
		t.Errorf("Hit point not on sphere surface: distance=%v, radius=%v", distFromCenter, sphere.Radius)
	}
}

func TestSphereNoHitMiss(t *testing.T) {
	rnd := NewRandomSource()
	sphere := Sphere{
		Center: Vec3{0, 0, -1},
		Radius: 0.5,
		Mat:    Lambertian{Albedo: ColorF{1, 0, 0}},
	}
	// Ray that misses the sphere
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{2, 0, -1})

	hit, _ := sphere.Hit(ray, FrontEpsilon)

	if hit {
		t.Error("Expected no hit")
	}
}

func TestSphereHitNormal(t *testing.T) {
	rnd := NewRandomSource()
	sphere := Sphere{
		Center: Vec3{0, 0, 0},
		Radius: 1.0,
		Mat:    Lambertian{Albedo: ColorF{1, 0, 0}},
	}
	// Ray from positive X hitting sphere
	ray := rnd.NewRay(Vec3{2, 0, 0}, Vec3{-1, 0, 0})

	hit, rec := sphere.Hit(ray, FrontEpsilon)

	if !hit {
		t.Fatal("Expected hit")
	}
	if !rec.FrontFace {
		t.Error("Expected front face hit")
	}
	// Normal should point outward (in +X direction)
	expectedNormal := Vec3{1, 0, 0}
	if Length(Sub(rec.Normal, expectedNormal)) > 1e-10 {
		t.Errorf("Expected normal %v, got %v", expectedNormal, rec.Normal)
	}
}

func TestSphereHitFromInside(t *testing.T) {
	rnd := NewRandomSource()
	sphere := Sphere{
		Center: Vec3{0, 0, 0},
		Radius: 1.0,
		Mat:    Lambertian{Albedo: ColorF{1, 0, 0}},
	}
	// Ray from inside sphere going out
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{1, 0, 0})

	hit, rec := sphere.Hit(ray, Front)

	if !hit {
		t.Fatal("Expected hit")
	}
	if rec.FrontFace {
		t.Error("Expected back face hit (from inside)")
	}
}

func TestSphereHitInterval(t *testing.T) {
	rnd := NewRandomSource()
	sphere := Sphere{
		Center: Vec3{0, 0, -5},
		Radius: 1.0,
		Mat:    Lambertian{Albedo: ColorF{1, 0, 0}},
	}
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, 0, -1})

	// Hit with acceptable interval
	hit, _ := sphere.Hit(ray, Interval{Start: 0, End: 10})
	if !hit {
		t.Error("Expected hit with interval [0, 10]")
	}

	// Miss with interval that excludes the hit
	hit, _ = sphere.Hit(ray, Interval{Start: 0, End: 3})
	if hit {
		t.Error("Expected no hit with interval [0, 3]")
	}

	hit, _ = sphere.Hit(ray, Interval{Start: 10, End: 20})
	if hit {
		t.Error("Expected no hit with interval [10, 20]")
	}
}

func TestSceneHitSingleObject(t *testing.T) {
	rnd := NewRandomSource()
	sphere := &Sphere{
		Center: Vec3{0, 0, -1},
		Radius: 0.5,
		Mat:    Lambertian{Albedo: ColorF{1, 0, 0}},
	}
	scene := Scene{Objects: []Hittable{sphere}}
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, 0, -1})

	hit, rec := scene.Hit(ray, FrontEpsilon)

	if !hit {
		t.Fatal("Expected hit")
	}
	if rec.Mat == nil {
		t.Error("Expected material to be set")
	}
}

func TestSceneHitMultipleObjects(t *testing.T) {
	rnd := NewRandomSource()
	sphere1 := &Sphere{
		Center: Vec3{0, 0, -1},
		Radius: 0.5,
		Mat:    Lambertian{Albedo: ColorF{1, 0, 0}},
	}
	sphere2 := &Sphere{
		Center: Vec3{0, 0, -2},
		Radius: 0.5,
		Mat:    Metal{Albedo: ColorF{0.8, 0.8, 0.8}},
	}
	scene := Scene{Objects: []Hittable{sphere1, sphere2}}
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, 0, -1})

	hit, rec := scene.Hit(ray, FrontEpsilon)

	if !hit {
		t.Fatal("Expected hit")
	}
	// Should hit the closer sphere (sphere1)
	expectedT := 0.5 // approximately
	if math.Abs(rec.T-expectedT) > 0.1 {
		t.Errorf("Expected t close to %v, got %v", expectedT, rec.T)
	}
}

func TestSceneNoHit(t *testing.T) {
	rnd := NewRandomSource()
	sphere := &Sphere{
		Center: Vec3{0, 0, -1},
		Radius: 0.5,
		Mat:    Lambertian{Albedo: ColorF{1, 0, 0}},
	}
	scene := Scene{Objects: []Hittable{sphere}}
	// Ray that misses all objects
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{10, 0, -1})

	hit, _ := scene.Hit(ray, FrontEpsilon)

	if hit {
		t.Error("Expected no hit")
	}
}

func TestRayColorBackgroundGradient(t *testing.T) {
	rnd := NewRandomSource()
	scene := &Scene{Objects: []Hittable{}}
	// Ray pointing straight down (should give more blue)
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, -1, 0})

	color := scene.RayColor(ray, 10)

	// Should be more blue than white
	if color[2] < color[0] {
		t.Errorf("Expected blue component to be higher for downward ray, got %v", color)
	}
}

func TestRayColorDepthLimit(t *testing.T) {
	rnd := NewRandomSource()
	sphere := &Sphere{
		Center: Vec3{0, 0, -1},
		Radius: 0.5,
		Mat:    Lambertian{Albedo: ColorF{1, 1, 1}},
	}
	scene := &Scene{Objects: []Hittable{sphere}}
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, 0, -1})

	// With depth 0, should return black
	color := scene.RayColor(ray, 0)
	expected := ColorF{0, 0, 0}
	if color != expected {
		t.Errorf("Expected black with depth 0, got %v", color)
	}

	// With positive depth, should scatter
	color = scene.RayColor(ray, 5)
	if color == expected {
		t.Error("Expected non-black color with positive depth")
	}
}

func TestRayColorWithLambertian(t *testing.T) {
	rnd := NewRandomSource()
	sphere := &Sphere{
		Center: Vec3{0, 0, -1},
		Radius: 0.5,
		Mat:    Lambertian{Albedo: ColorF{0.5, 0.5, 0.5}},
	}
	scene := &Scene{Objects: []Hittable{sphere}}
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, 0, -1})

	color := scene.RayColor(ray, 5)

	// Should not be black or pure white
	if color == (ColorF{0, 0, 0}) {
		t.Error("Expected non-black color")
	}
	if color == (ColorF{1, 1, 1}) {
		t.Error("Expected color not to be pure white")
	}
}

func TestRayColorWithMetal(t *testing.T) {
	t.Helper()
	rnd := NewRandomSource()
	sphere := &Sphere{
		Center: Vec3{0, 0, -1},
		Radius: 0.5,
		Mat:    Metal{Albedo: ColorF{0.8, 0.8, 0.8}, Fuzz: 0},
	}
	scene := &Scene{Objects: []Hittable{sphere}}
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, 0, -1})

	color := scene.RayColor(ray, 5)

	// Should produce some color (might be black if absorbed)
	_ = color // Just ensure it runs without panic
}

func TestRayColorWithDielectric(t *testing.T) {
	t.Helper()
	rnd := NewRandomSource()
	sphere := &Sphere{
		Center: Vec3{0, 0, -1},
		Radius: 0.5,
		Mat:    Dielectric{RefIdx: 1.5},
	}
	scene := &Scene{Objects: []Hittable{sphere}}
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, 0, -1})

	color := scene.RayColor(ray, 5)

	// Should produce some color
	_ = color // Just ensure it runs without panic
}

func TestDefaultScene(t *testing.T) {
	scene := DefaultScene()

	if scene == nil {
		t.Fatal("Expected non-nil scene")
	}
	if len(scene.Objects) == 0 {
		t.Error("Expected default scene to have objects")
	}

	// Test that default scene can be rendered
	rnd := NewRandomSource()
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, 0, -1})
	color := scene.RayColor(ray, 5)
	_ = color // Just ensure it runs without panic
}

func TestDefaultSceneHasDifferentMaterials(t *testing.T) {
	scene := DefaultScene()

	// Check that we have different material types
	hasLambertian := false
	hasMetal := false
	hasDielectric := false

	for _, obj := range scene.Objects {
		sphere, ok := obj.(*Sphere)
		if !ok {
			continue
		}
		switch sphere.Mat.(type) {
		case Lambertian:
			hasLambertian = true
		case Metal:
			hasMetal = true
		case Dielectric:
			hasDielectric = true
		}
	}

	if !hasLambertian {
		t.Error("Expected default scene to have Lambertian material")
	}
	if !hasMetal {
		t.Error("Expected default scene to have Metal material")
	}
	if !hasDielectric {
		t.Error("Expected default scene to have Dielectric material")
	}
}
