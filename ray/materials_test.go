package ray

import (
	"math"
	"testing"
)

func TestLambertianScatter(t *testing.T) {
	rnd := NewRandomSource()
	lambertian := Lambertian{Albedo: ColorF{0.5, 0.5, 0.5}}
	ray := rnd.NewRay(Vec3{0, 0, 0}, Vec3{0, 0, -1})
	rec := HitRecord{
		Point:  Vec3{0, 0, -1},
		Normal: Vec3{0, 0, 1},
	}

	didScatter, attenuation, scattered := lambertian.Scatter(ray, rec)

	if !didScatter {
		t.Error("Expected Lambertian to scatter")
	}
	if attenuation != lambertian.Albedo {
		t.Errorf("Expected attenuation %v, got %v", lambertian.Albedo, attenuation)
	}
	if scattered == nil {
		t.Error("Expected scattered ray to be non-nil")
	} else if scattered.Origin != rec.Point {
		t.Errorf("Expected scattered origin %v, got %v", rec.Point, scattered.Origin)
	}
}

func TestMetalScatter(t *testing.T) {
	rnd := NewRandomSource()
	metal := Metal{Albedo: ColorF{0.8, 0.8, 0.8}, Fuzz: 0}
	rayDir := Unit(Vec3{1, -1, 0})
	ray := rnd.NewRay(Vec3{0, 2, 0}, rayDir)
	rec := HitRecord{
		Point:  Vec3{1, 1, 0},
		Normal: Vec3{0, 1, 0},
	}

	didScatter, attenuation, scattered := metal.Scatter(ray, rec)

	if !didScatter {
		t.Error("Expected metal to scatter")
	}
	if attenuation != metal.Albedo {
		t.Errorf("Expected attenuation %v, got %v", metal.Albedo, attenuation)
	}
	if scattered == nil {
		t.Fatal("Expected scattered ray")
	}
	if scattered.Origin != rec.Point {
		t.Errorf("Expected scattered origin %v, got %v", rec.Point, scattered.Origin)
	}
}

func TestMetalScatterWithFuzz(t *testing.T) {
	rnd := NewRandomSource()
	metal := Metal{Albedo: ColorF{0.9, 0.9, 0.9}, Fuzz: 0.3}
	rayDir := Unit(Vec3{1, -1, 0})
	ray := rnd.NewRay(Vec3{0, 2, 0}, rayDir)
	rec := HitRecord{
		Point:  Vec3{1, 1, 0},
		Normal: Vec3{0, 1, 0},
	}

	didScatter, _, scattered := metal.Scatter(ray, rec)

	if !didScatter {
		t.Error("Expected metal to scatter")
	}
	if scattered == nil {
		t.Fatal("Expected scattered ray")
	}
	// With fuzz, direction should be perturbed
	if scattered.Origin != rec.Point {
		t.Errorf("Expected scattered origin %v, got %v", rec.Point, scattered.Origin)
	}
}

func TestMetalScatterAbsorbedWhenReflectionBelowSurface(t *testing.T) {
	rnd := NewRandomSource()
	// High fuzz (>1) can cause scatter to be absorbed when fuzzed reflection goes below surface
	metal := Metal{Albedo: ColorF{0.7, 0.7, 0.7}, Fuzz: 1.5}
	rayDir := Unit(Vec3{1, -1, 0})
	ray := rnd.NewRay(Vec3{0, 2, 0}, rayDir)
	rec := HitRecord{
		Point:  Vec3{1, 1, 0},
		Normal: Vec3{0, 1, 0},
	}

	// With high fuzz, absorption is possible (test that it doesn't panic)
	// Try a few times to exercise both scatter and absorption paths
	hasScattered := false
	hasAbsorbed := false
	for range 50 {
		didScatter, _, _ := metal.Scatter(ray, rec)
		if didScatter {
			hasScattered = true
		} else {
			hasAbsorbed = true
		}
		if hasScattered && hasAbsorbed {
			break // Both paths tested
		}
	}
	if !hasScattered {
		t.Error("Expected at least some rays to scatter with fuzz=1.5")
	}
}

func TestDielectricScatterFrontFace(t *testing.T) {
	rnd := NewRandomSource()
	dielectric := Dielectric{RefIdx: 1.5}
	rayDir := Unit(Vec3{0, -1, 0})
	ray := rnd.NewRay(Vec3{0, 2, 0}, rayDir)
	rec := HitRecord{
		Point:     Vec3{0, 0, 0},
		Normal:    Vec3{0, 1, 0},
		FrontFace: true,
	}

	didScatter, attenuation, scattered := dielectric.Scatter(ray, rec)

	if !didScatter {
		t.Error("Expected dielectric to scatter")
	}
	expected := ColorF{1.0, 1.0, 1.0}
	if attenuation != expected {
		t.Errorf("Expected attenuation %v, got %v", expected, attenuation)
	}
	if scattered == nil {
		t.Fatal("Expected scattered ray")
	}
	if scattered.Origin != rec.Point {
		t.Errorf("Expected scattered origin %v, got %v", rec.Point, scattered.Origin)
	}
}

func TestDielectricScatterBackFace(t *testing.T) {
	rnd := NewRandomSource()
	dielectric := Dielectric{RefIdx: 1.5}
	rayDir := Unit(Vec3{0, 1, 0})
	ray := rnd.NewRay(Vec3{0, -2, 0}, rayDir)
	rec := HitRecord{
		Point:     Vec3{0, 0, 0},
		Normal:    Vec3{0, 1, 0},
		FrontFace: false,
	}

	didScatter, attenuation, scattered := dielectric.Scatter(ray, rec)

	if !didScatter {
		t.Error("Expected dielectric to scatter")
	}
	expected := ColorF{1.0, 1.0, 1.0}
	if attenuation != expected {
		t.Errorf("Expected attenuation %v, got %v", expected, attenuation)
	}
	if scattered != nil && scattered.Origin != rec.Point {
		t.Errorf("Expected scattered origin %v, got %v", rec.Point, scattered.Origin)
	}
}

func TestDielectricScatterVariousAngles(t *testing.T) {
	// Test different incident angles to exercise both refraction and reflection paths
	rnd := NewRandomSource()
	dielectric := Dielectric{RefIdx: 1.5}

	tests := []struct {
		name      string
		rayDir    Vec3
		frontFace bool
	}{
		{"Front perpendicular", Vec3{0, -1, 0}, true},
		{"Front angled", Vec3{1, -1, 0}, true},
		{"Back perpendicular", Vec3{0, 1, 0}, false},
		{"Back angled", Vec3{1, 1, 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ray := rnd.NewRay(Vec3{0, 0, 0}, Unit(tt.rayDir))
			rec := HitRecord{
				Point:     Vec3{0, 1, 0},
				Normal:    Vec3{0, 1, 0},
				FrontFace: tt.frontFace,
			}

			didScatter, attenuation, scattered := dielectric.Scatter(ray, rec)
			if !didScatter {
				t.Error("Dielectric should always scatter")
			}
			if attenuation != (ColorF{1, 1, 1}) {
				t.Errorf("Expected white attenuation, got %v", attenuation)
			}
			if scattered == nil || scattered.Origin != rec.Point {
				t.Error("Expected valid scattered ray from hit point")
			}
		})
	}
}

func TestReflectance(t *testing.T) {
	tests := []struct {
		cosine float64
		refIdx float64
	}{
		{0.5, 1.5},
		{0.0, 1.5},
		{1.0, 1.5},
		{0.7, 1.33},
		{0.9, 2.0},
	}

	for _, tt := range tests {
		result := Reflectance(tt.cosine, tt.refIdx)
		if result < 0 || result > 1 {
			t.Errorf("Reflectance(%v, %v) = %v, expected value in [0,1]", tt.cosine, tt.refIdx, result)
		}

		// Check Schlick's approximation formula manually
		r0 := (1 - tt.refIdx) / (1 + tt.refIdx)
		r0 *= r0
		expected := r0 + (1-r0)*math.Pow((1-tt.cosine), 5)
		if math.Abs(result-expected) > 1e-10 {
			t.Errorf("Reflectance(%v, %v) = %v, expected %v", tt.cosine, tt.refIdx, result, expected)
		}
	}
}
