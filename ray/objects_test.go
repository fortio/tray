package ray

import (
	"math"
	"testing"
)

// Test helper to preserve the original return pattern (bool, *HitRecord).
func testHit(h Hittable, r *Ray, i Interval) (bool, *HitRecord) {
	rec := &HitRecord{}
	hit := h.Hit(r, i, rec)
	return hit, rec
}

func TestSetFaceNormalFrontFace(t *testing.T) {
	rnd := RandForTests()
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{0, 0, -1})
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
	rnd := RandForTests()
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{0, 0, 1})
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
	rnd := RandForTests()
	sphere := NewSphere(
		Vec3{0, 0, -1},
		0.5,
		Lambertian{Albedo: ColorF{1, 0, 0}},
	)
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{0, 0, -1})

	hit, rec := testHit(sphere, ray, FrontEpsilon)

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
	rnd := RandForTests()
	sphere := NewSphere(
		Vec3{0, 0, -1},
		0.5,
		Lambertian{Albedo: ColorF{1, 0, 0}},
	)
	// Ray that misses the sphere
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{2, 0, -1})

	hit, _ := testHit(sphere, ray, FrontEpsilon)

	if hit {
		t.Error("Expected no hit")
	}
}

func TestSphereHitNormal(t *testing.T) {
	rnd := RandForTests()
	sphere := NewSphere(
		Vec3{0, 0, 0},
		1.0,
		Lambertian{Albedo: ColorF{1, 0, 0}},
	)
	// Ray from positive X hitting sphere
	ray := NewRay(rnd, Vec3{2, 0, 0}, Vec3{-1, 0, 0})

	hit, rec := testHit(sphere, ray, FrontEpsilon)

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
	rnd := RandForTests()
	sphere := NewSphere(
		Vec3{0, 0, 0},
		1.0,
		Lambertian{Albedo: ColorF{1, 0, 0}},
	)
	// Ray from inside sphere going out
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{1, 0, 0})

	hit, rec := testHit(sphere, ray, Front)

	if !hit {
		t.Fatal("Expected hit")
	}
	if rec.FrontFace {
		t.Error("Expected back face hit (from inside)")
	}
}

func TestSphereHitInterval(t *testing.T) {
	rnd := RandForTests()
	sphere := NewSphere(
		Vec3{0, 0, -5},
		1.0,
		Lambertian{Albedo: ColorF{1, 0, 0}},
	)
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{0, 0, -1})

	// Hit with acceptable interval
	hit, _ := testHit(sphere, ray, Interval{Start: 0, End: 10})
	if !hit {
		t.Error("Expected hit with interval [0, 10]")
	}

	// Miss with interval that excludes the hit
	hit, _ = testHit(sphere, ray, Interval{Start: 0, End: 3})
	if hit {
		t.Error("Expected no hit with interval [0, 3]")
	}

	hit, _ = testHit(sphere, ray, Interval{Start: 10, End: 20})
	if hit {
		t.Error("Expected no hit with interval [10, 20]")
	}
}

func TestSceneHitSingleObject(t *testing.T) {
	rnd := RandForTests()
	sphere := NewSphere(
		Vec3{0, 0, -1},
		0.5,
		Lambertian{Albedo: ColorF{1, 0, 0}},
	)
	scene := Scene{Objects: []Hittable{sphere}}
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{0, 0, -1})

	hit, rec := testHit(&scene, ray, FrontEpsilon)

	if !hit {
		t.Fatal("Expected hit")
	}
	if rec.Mat == nil {
		t.Error("Expected material to be set")
	}
}

func TestSceneHitMultipleObjects(t *testing.T) {
	rnd := RandForTests()
	sphere1 := NewSphere(
		Vec3{0, 0, -1},
		0.5,
		Lambertian{Albedo: ColorF{1, 0, 0}},
	)
	sphere2 := NewSphere(
		Vec3{0, 0, -2},
		0.5,
		Metal{Albedo: ColorF{0.8, 0.8, 0.8}},
	)
	scene := Scene{Objects: []Hittable{sphere1, sphere2}}
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{0, 0, -1})

	hit, rec := testHit(&scene, ray, FrontEpsilon)

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
	rnd := RandForTests()
	sphere := NewSphere(
		Vec3{0, 0, -1},
		0.5,
		Lambertian{Albedo: ColorF{1, 0, 0}},
	)
	scene := Scene{Objects: []Hittable{sphere}}
	// Ray that misses all objects
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{10, 0, -1})

	hit, _ := testHit(&scene, ray, FrontEpsilon)

	if hit {
		t.Error("Expected no hit")
	}
}

func TestRayColorBackgroundGradient(t *testing.T) {
	rnd := RandForTests()
	scene := &Scene{Objects: []Hittable{}}
	// Ray pointing straight down (should give more blue)
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{0, -1, 0})

	color := scene.RayColor(ray, 10)

	// Should be more blue than white
	if color.z < color.x {
		t.Errorf("Expected blue component to be higher for downward ray, got %v", color)
	}
}

func TestRayColorDepthLimit(t *testing.T) {
	rnd := RandForTests()
	sphere := NewSphere(
		Vec3{0, 0, -1},
		0.5,
		Lambertian{Albedo: ColorF{1, 1, 1}},
	)
	scene := &Scene{Objects: []Hittable{sphere}, Background: DefaultBackground()}
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{0, 0, -1})

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

func TestRayColorDepthExhaustion(t *testing.T) {
	// Test that when rays keep scattering and depth runs out, we get black
	// A sphere with perfect reflection at very low depth should exhaust quickly
	rnd := RandForTests()

	// Create a scene where rays will keep bouncing
	sphere := NewSphere(
		Vec3{0, 0, -1},
		0.5,
		Lambertian{Albedo: ColorF{0.9, 0.9, 0.9}}, // High albedo, keeps bouncing
	)
	scene := &Scene{Objects: []Hittable{sphere}}

	// With maxDepth=1, after first bounce depth becomes 0 and returns black
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{0, 0, -1})
	color := scene.RayColor(ray, 1)

	// Should get some color from first bounce then black from second
	// The result will be attenuated but not pure black due to first bounce
	if color == (ColorF{1, 1, 1}) {
		t.Error("Expected color to be affected by depth limitation")
	}
}

func TestRayColorWithDifferentMaterials(t *testing.T) {
	rnd := RandForTests()
	tests := []struct {
		name string
		mat  Material
	}{
		{"Lambertian", Lambertian{Albedo: ColorF{0.5, 0.5, 0.5}}},
		{"Metal", Metal{Albedo: ColorF{0.8, 0.8, 0.8}, Fuzz: 0}},
		{"Dielectric", Dielectric{RefIdx: 1.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sphere := NewSphere(
				Vec3{0, 0, -1},
				0.5,
				tt.mat,
			)
			scene := &Scene{Objects: []Hittable{sphere}}
			ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{0, 0, -1})

			color := scene.RayColor(ray, 5)

			// All materials should produce valid colors (components in [0,1])
			for i, c := range color.Components() {
				if c < 0 || c > 1 {
					t.Errorf("color[%d] = %f, want in range [0,1]", i, c)
				}
			}
		})
	}
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
	rnd := RandForTests()
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{0, 0, -1})
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

func TestRayColorMaterialAbsorption(t *testing.T) {
	// Test that when material doesn't scatter, RayColor returns black
	// Metal with very high fuzz can absorb when fuzzed reflection goes below surface
	rnd := RandForTests()
	sphere := NewSphere(
		Vec3{0, 0, -5},
		1.0,
		Metal{Albedo: ColorF{0.8, 0.8, 0.8}, Fuzz: 5.0},
	)
	scene := &Scene{Objects: []Hittable{sphere}}
	ray := NewRay(rnd, Vec3{0, 0, 0}, Vec3{0, 0, -1})

	// Test that absorption path (didScatter=false) doesn't crash
	for range 100 {
		color := scene.RayColor(ray, 5)
		// Valid result is either absorbed (black) or scattered (some color in [0,1])
		for i, c := range color.Components() {
			if c < 0 || c > 1 {
				t.Fatalf("Invalid color[%d] = %f", i, c)
			}
		}
	}
}
