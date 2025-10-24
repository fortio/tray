package ray

import (
	"image/color"
	"math"
	"testing"

	"fortio.org/sets"
)

func TestVec3Add(t *testing.T) {
	v := Vec3{1, 2, 3}
	u := XYZ(4, 5, 6)
	result := Add(v, u)
	expected := Vec3{5, 7, 9}
	if result != expected {
		t.Errorf("Add() = %v, want %v", result, expected)
	}
}

func TestVec3Sub(t *testing.T) {
	v := Vec3{5, 7, 9}
	u := Vec3{1, 2, 3}
	result := Sub(v, u)
	expected := Vec3{4, 5, 6}
	if result != expected {
		t.Errorf("Sub() = %v, want %v", result, expected)
	}
	u = Neg(u)
	result2 := Add(v, u)
	if result2 != expected {
		t.Errorf("Add with Neg() = %v, want %v", result2, expected)
	}
}

func TestAddMultiple(t *testing.T) {
	tests := []struct {
		name     string
		u        Vec3
		vs       []Vec3
		expected Vec3
	}{
		{"single vector", Vec3{1, 2, 3}, []Vec3{{4, 5, 6}}, Vec3{5, 7, 9}},
		{"two vectors", Vec3{1, 2, 3}, []Vec3{{4, 5, 6}, {7, 8, 9}}, Vec3{12, 15, 18}},
		{"three vectors", Vec3{1, 1, 1}, []Vec3{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}, Vec3{2, 2, 2}},
		{"no additional vectors", Vec3{5, 10, 15}, []Vec3{}, Vec3{5, 10, 15}},
		{"with negatives", Vec3{10, 10, 10}, []Vec3{{-5, 0, 5}, {3, -3, 0}}, Vec3{8, 7, 15}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddMultiple(tt.u, tt.vs...)
			if result != tt.expected {
				t.Errorf("AddMultiple() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSubMultiple(t *testing.T) {
	tests := []struct {
		name     string
		u        Vec3
		v0       Vec3
		vs       []Vec3
		expected Vec3
	}{
		{"single subtraction", Vec3{10, 10, 10}, Vec3{1, 2, 3}, []Vec3{}, Vec3{9, 8, 7}},
		{"two subtractions", Vec3{10, 10, 10}, Vec3{1, 2, 3}, []Vec3{{2, 3, 4}}, Vec3{7, 5, 3}},
		{"three subtractions", Vec3{20, 20, 20}, Vec3{5, 5, 5}, []Vec3{{3, 3, 3}, {2, 2, 2}}, Vec3{10, 10, 10}},
		{"with negatives", Vec3{10, 10, 10}, Vec3{5, 5, 5}, []Vec3{{-2, -2, -2}}, Vec3{7, 7, 7}},
		{"equivalent to nested Sub", Vec3{20, 30, 40}, Vec3{5, 10, 15}, []Vec3{{2, 3, 4}, {1, 1, 1}}, Vec3{12, 16, 20}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubMultiple(tt.u, tt.v0, tt.vs...)
			if result != tt.expected {
				t.Errorf("SubMultiple() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSubMultipleUsageExample(t *testing.T) {
	// Test the actual usage from ray.go
	camera := Vec3{0, 0, 0}
	horizontalHalf := Vec3{2, 0, 0}
	verticalHalf := Vec3{0, 1, 0}
	focal := Vec3{0, 0, 1}

	// SubMultiple(camera, horizontalHalf, verticalHalf, focal)
	// should equal camera - horizontalHalf - verticalHalf - focal
	result := SubMultiple(camera, horizontalHalf, verticalHalf, focal)
	expected := Vec3{-2, -1, -1}

	if result != expected {
		t.Errorf("SubMultiple() = %v, want %v", result, expected)
	}

	// Verify equivalence with nested Sub
	nested := Sub(Sub(Sub(camera, horizontalHalf), verticalHalf), focal)
	if result != nested {
		t.Errorf("SubMultiple() = %v, nested Sub() = %v, should be equal", result, nested)
	}
}

func TestMethodStyleAPI(t *testing.T) {
	// Test that method-style API produces same results as function-style
	v1 := Vec3{10, 20, 30}
	v2 := Vec3{1, 2, 3}
	v3 := Vec3{4, 5, 6}

	// Test Plus
	methodPlus := v1.Plus(v2, v3)
	funcPlus := AddMultiple(v1, v2, v3)
	if methodPlus != funcPlus {
		t.Errorf("Plus() = %v, AddMultiple() = %v, should be equal", methodPlus, funcPlus)
	}

	// Test Minus
	methodMinus := v1.Minus(v2, v3)
	funcMinus := SubMultiple(v1, v2, v3)
	if methodMinus != funcMinus {
		t.Errorf("Minus() = %v, SubMultiple() = %v, should be equal", methodMinus, funcMinus)
	}

	// Test Times
	methodTimes := v1.Times(2.5)
	funcTimes := SMul(v1, 2.5)
	if methodTimes != funcTimes {
		t.Errorf("Times() = %v, SMul() = %v, should be equal", methodTimes, funcTimes)
	}
}

func TestMethodStyleChaining(t *testing.T) {
	// Test realistic chaining as used in ray.go
	camera := XYZ(0, 0, 0)
	horizontal := XYZ(4, 0, 0)
	vertical := XYZ(0, 2, 0)
	focal := XYZ(0, 0, 1)

	// Method style (readable)
	upperLeft := camera.Minus(horizontal.Times(0.5), vertical.Times(0.5), focal)

	// Function style (equivalent)
	upperLeftFunc := SubMultiple(camera, SMul(horizontal, 0.5), SMul(vertical, 0.5), focal)

	if upperLeft != upperLeftFunc {
		t.Errorf("Method style = %v, function style = %v, should be equal", upperLeft, upperLeftFunc)
	}

	// Expected result
	expected := Vec3{-2, -1, -1}
	if upperLeft != expected {
		t.Errorf("upperLeft = %v, want %v", upperLeft, expected)
	}
}

func TestVec3SMul(t *testing.T) {
	v := Vec3{1, 2, 3}
	result := SMul(v, 2.5)
	expected := Vec3{2.5, 5.0, 7.5}
	if result != expected {
		t.Errorf("SMul() = %v, want %v", result, expected)
	}
}

func TestVec3SDiv(t *testing.T) {
	v := Vec3{10, 20, 30}
	result := SDiv(v, 10)
	expected := Vec3{1, 2, 3}
	if result != expected {
		t.Errorf("SDiv() = %v, want %v", result, expected)
	}
}

func TestVec3Length(t *testing.T) {
	tests := []struct {
		name     string
		v        Vec3
		expected float64
	}{
		{"unit vector", Vec3{1, 0, 0}, 1.0},
		{"3-4-5 triangle", Vec3{3, 4, 0}, 5.0},
		{"zero vector", Vec3{0, 0, 0}, 0.0},
		{"negative values", Vec3{-1, -2, -2}, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Length(tt.v)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Length() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVec3Unit(t *testing.T) {
	v := Vec3{3, 4, 0}
	result := Unit(v)
	expected := Vec3{0.6, 0.8, 0.0}

	for i := range 3 {
		if math.Abs(result[i]-expected[i]) > 1e-9 {
			t.Errorf("Unit()[%d] = %v, want %v", i, result[i], expected[i])
		}
	}

	// Check that unit vector has length 1
	length := Length(result)
	if math.Abs(length-1.0) > 1e-9 {
		t.Errorf("Unit().Length() = %v, want 1.0", length)
	}
}

func TestVec3Accessors(t *testing.T) {
	v := Vec3{1.5, 2.5, 3.5}

	if v.X() != 1.5 {
		t.Errorf("X() = %v, want 1.5", v.X())
	}
	if v.Y() != 2.5 {
		t.Errorf("Y() = %v, want 2.5", v.Y())
	}
	if v.Z() != 3.5 {
		t.Errorf("Z() = %v, want 3.5", v.Z())
	}
}

func TestColorF(t *testing.T) {
	c := ColorF{0.5, 0.75, 1.0}
	if c[0] != 0.5 || c[1] != 0.75 || c[2] != 1.0 {
		t.Errorf("ColorF() = %v, want [0.5 0.75 1.0]", c)
	}
}

func TestFloatColorToRGBA(t *testing.T) {
	tests := []struct {
		name     string
		c        ColorF
		expected color.RGBA
	}{
		{"black", ColorF{0, 0, 0}, color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		{"white", ColorF{1, 1, 1}, color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"red", ColorF{1, 0, 0}, color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"green", ColorF{0, 1, 0}, color.RGBA{R: 0, G: 255, B: 0, A: 255}},
		{"blue", ColorF{0, 0, 1}, color.RGBA{R: 0, G: 0, B: 255, A: 255}},
		{"mid gray", ColorF{0.5, 0.5, 0.5}, color.RGBA{R: 127, G: 127, B: 127, A: 255}},
		{"clamped above", ColorF{1.5, 2.0, 3.0}, color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"clamped below", ColorF{-1.0, -0.5, -2.0}, color.RGBA{R: 0, G: 0, B: 0, A: 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.c.ToRGBA()
			if result != tt.expected {
				t.Errorf("ToRGBA() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		x        float64
		min      float64
		max      float64
		expected float64
	}{
		{"within range", 0.5, 0.0, 1.0, 0.5},
		{"below min", -0.5, 0.0, 1.0, 0.0},
		{"above max", 1.5, 0.0, 1.0, 1.0},
		{"at min", 0.0, 0.0, 1.0, 0.0},
		{"at max", 1.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := Interval{Start: tt.min, End: tt.max}
			result := i.Clamp(tt.x)
			if result != tt.expected {
				t.Errorf("clamp(%v, %v, %v) = %v, want %v", tt.x, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestDot(t *testing.T) {
	v1 := Vec3{1, 2, 3}
	v2 := Vec3{4, 5, 6}
	result := Dot(v1, v2)
	expected := 32.0 // 1*4 + 2*5 + 3*6
	if result != expected {
		t.Errorf("Dot() = %v, want %v", result, expected)
	}
}

func TestIntervalLength(t *testing.T) {
	tests := []struct {
		name     string
		i        Interval
		expected float64
	}{
		{"positive range", Interval{Start: 0, End: 10}, 10},
		{"negative range", Interval{Start: -5, End: 5}, 10},
		{"zero length", Interval{Start: 5, End: 5}, 0},
		{"unit interval", ZeroOne, 1},
		{"negative length (empty)", Empty, math.Inf(-1)},
		{"infinite interval", Universe, math.Inf(1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.i.Length()
			if math.IsInf(tt.expected, 0) {
				if !math.IsInf(result, int(math.Copysign(1, tt.expected))) {
					t.Errorf("Length() = %v, want %v", result, tt.expected)
				}
			} else if result != tt.expected {
				t.Errorf("Length() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIntervalContains(t *testing.T) {
	tests := []struct {
		name     string
		i        Interval
		t        float64
		expected bool
	}{
		{"inside range", Interval{Start: 0, End: 10}, 5, true},
		{"at start", Interval{Start: 0, End: 10}, 0, true},
		{"at end", Interval{Start: 0, End: 10}, 10, true},
		{"below range", Interval{Start: 0, End: 10}, -1, false},
		{"above range", Interval{Start: 0, End: 10}, 11, false},
		{"zero in ZeroOne", ZeroOne, 0, true},
		{"one in ZeroOne", ZeroOne, 1, true},
		{"half in ZeroOne", ZeroOne, 0.5, true},
		{"negative in ZeroOne", ZeroOne, -0.1, false},
		{"above ZeroOne", ZeroOne, 1.1, false},
		{"zero in Empty", Empty, 0, false},
		{"anything in Universe", Universe, 999999, true},
		{"negative infinity in Universe", Universe, math.Inf(-1), true},
		{"positive infinity in Universe", Universe, math.Inf(1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.i.Contains(tt.t)
			if result != tt.expected {
				t.Errorf("Contains(%v) = %v, want %v", tt.t, result, tt.expected)
			}
		})
	}
}

func TestIntervalSurrounds(t *testing.T) {
	tests := []struct {
		name     string
		i        Interval
		t        float64
		expected bool
	}{
		{"inside range", Interval{Start: 0, End: 10}, 5, true},
		{"at start", Interval{Start: 0, End: 10}, 0, false},
		{"at end", Interval{Start: 0, End: 10}, 10, false},
		{"below range", Interval{Start: 0, End: 10}, -1, false},
		{"above range", Interval{Start: 0, End: 10}, 11, false},
		{"zero in ZeroOne", ZeroOne, 0, false},
		{"one in ZeroOne", ZeroOne, 1, false},
		{"half in ZeroOne", ZeroOne, 0.5, true},
		{"negative in ZeroOne", ZeroOne, -0.1, false},
		{"above ZeroOne", ZeroOne, 1.1, false},
		{"zero in Empty", Empty, 0, false},
		{"large value in Universe", Universe, 999999, true},
		{"negative infinity in Universe", Universe, math.Inf(-1), false},
		{"positive infinity in Universe", Universe, math.Inf(1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.i.Surrounds(tt.t)
			if result != tt.expected {
				t.Errorf("Surrounds(%v) = %v, want %v", tt.t, result, tt.expected)
			}
		})
	}
}

func TestIntervalClamp(t *testing.T) {
	tests := []struct {
		name     string
		i        Interval
		t        float64
		expected float64
	}{
		{"within range", Interval{Start: 0, End: 10}, 5, 5},
		{"at start", Interval{Start: 0, End: 10}, 0, 0},
		{"at end", Interval{Start: 0, End: 10}, 10, 10},
		{"below range", Interval{Start: 0, End: 10}, -5, 0},
		{"above range", Interval{Start: 0, End: 10}, 15, 10},
		{"negative to ZeroOne", ZeroOne, -0.5, 0},
		{"above ZeroOne", ZeroOne, 1.5, 1},
		{"within ZeroOne", ZeroOne, 0.75, 0.75},
		{"negative range below", Interval{Start: -10, End: -5}, -15, -10},
		{"negative range above", Interval{Start: -10, End: -5}, 0, -5},
		{"negative range within", Interval{Start: -10, End: -5}, -7, -7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.i.Clamp(tt.t)
			if result != tt.expected {
				t.Errorf("Clamp(%v) = %v, want %v", tt.t, result, tt.expected)
			}
		})
	}
}

func TestIntervalPredefinedConstants(t *testing.T) {
	// Test Empty interval
	if !math.IsInf(Empty.Start, 1) {
		t.Errorf("Empty.Start = %v, want +Inf", Empty.Start)
	}
	if !math.IsInf(Empty.End, -1) {
		t.Errorf("Empty.End = %v, want -Inf", Empty.End)
	}
	if Empty.Contains(0) {
		t.Error("Empty.Contains(0) = true, want false")
	}

	// Test Universe interval
	if !math.IsInf(Universe.Start, -1) {
		t.Errorf("Universe.Start = %v, want -Inf", Universe.Start)
	}
	if !math.IsInf(Universe.End, 1) {
		t.Errorf("Universe.End = %v, want +Inf", Universe.End)
	}
	if !Universe.Contains(0) {
		t.Error("Universe.Contains(0) = false, want true")
	}
	if !Universe.Contains(math.MaxFloat64) {
		t.Error("Universe.Contains(MaxFloat64) = false, want true")
	}

	// Test Front interval
	if Front.Start != 0 {
		t.Errorf("Front.Start = %v, want 0", Front.Start)
	}
	if !math.IsInf(Front.End, 1) {
		t.Errorf("Front.End = %v, want +Inf", Front.End)
	}
	if !Front.Contains(100) {
		t.Error("Front.Contains(100) = false, want true")
	}
	if Front.Contains(-1) {
		t.Error("Front.Contains(-1) = true, want false")
	}

	// Test ZeroOne interval
	if ZeroOne.Start != 0 {
		t.Errorf("ZeroOne.Start = %v, want 0", ZeroOne.Start)
	}
	if ZeroOne.End != 1 {
		t.Errorf("ZeroOne.End = %v, want 1", ZeroOne.End)
	}
	if ZeroOne.Length() != 1 {
		t.Errorf("ZeroOne.Length() = %v, want 1", ZeroOne.Length())
	}
}

// TestRandom just... exercises the Random function
// and that values are ... different.
func TestRandom(t *testing.T) {
	const samples = 10
	results := sets.New[Vec3]()
	expected := Interval{Start: 0.0, End: 1.0}
	for range samples {
		v := Random[Vec3]()
		// Check each component is in [0,1)
		for i := range 3 {
			if !expected.Contains(v[i]) {
				t.Errorf("Random() component %d = %v, want in [0,1)", i, v[i])
			}
		}
		// Collect unique samples
		results.Add(v)
	}
	if results.Len() != samples {
		t.Errorf("Random() produced %d unique samples, want %d", results.Len(), samples)
	}
}

// TestRandomUnitVectorCorrectness verifies that all three RandomUnitVector variants
// produce vectors of unit length.
func TestRandomUnitVectorCorrectness(t *testing.T) {
	tests := []struct {
		name string
		fn   func() Vec3
	}{
		{"RandomUnitVector", RandomUnitVector[Vec3]},
		{"RandomUnitVectorAngle", RandomUnitVectorAngle[Vec3]},
		{"RandomUnitVectorNorm", RandomUnitVector[Vec3]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const samples = 100
			const tolerance = 1e-9

			for i := range samples {
				v := tt.fn()
				length := Length(v)

				if math.Abs(length-1.0) > tolerance {
					t.Errorf("sample %d: Length() = %.15f, want 1.0 (diff: %.15e)",
						i, length, length-1.0)
				}
			}
		})
	}
}

// TestRandomUnitVectorDistribution checks that the generated vectors are
// uniformly distributed over the unit sphere by testing:
// 1. Mean of components approaches zero.
// 2. Standard deviation of each component approaches expected value.
// 3. Points cover all octants of the sphere.
func TestRandomUnitVectorDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping distribution test in short mode")
	}

	tests := []struct {
		name string
		fn   func() Vec3
	}{
		{"RandomUnitVectorRej", RandomUnitVectorRej[Vec3]},
		{"RandomUnitVectorAngle", RandomUnitVectorAngle[Vec3]},
		{"RandomUnitVectorNorm", RandomUnitVector[Vec3]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const samples = 100000

			// Track statistics
			var sumX, sumY, sumZ float64
			var sumX2, sumY2, sumZ2 float64
			octantCounts := make([]int, 8)

			for range samples {
				v := tt.fn()
				x, y, z := v[0], v[1], v[2]

				// Accumulate for mean and variance
				sumX += x
				sumY += y
				sumZ += z
				sumX2 += x * x
				sumY2 += y * y
				sumZ2 += z * z

				// Count octant
				octant := 0
				if x > 0 {
					octant |= 1
				}
				if y > 0 {
					octant |= 2
				}
				if z > 0 {
					octant |= 4
				}
				octantCounts[octant]++
			}

			// Check means are near zero
			meanX := sumX / samples
			meanY := sumY / samples
			meanZ := sumZ / samples

			// For uniform distribution on sphere, mean should be (0,0,0)
			// With 100k samples, standard error ≈ 1/sqrt(100000) ≈ 0.003
			// We use 5 sigma threshold for robustness: 5 * 0.003 ≈ 0.015
			const meanTolerance = 0.015
			if math.Abs(meanX) > meanTolerance {
				t.Errorf("mean X = %.6f, want ≈0 (within %.6f)", meanX, meanTolerance)
			}
			if math.Abs(meanY) > meanTolerance {
				t.Errorf("mean Y = %.6f, want ≈0 (within %.6f)", meanY, meanTolerance)
			}
			if math.Abs(meanZ) > meanTolerance {
				t.Errorf("mean Z = %.6f, want ≈0 (within %.6f)", meanZ, meanTolerance)
			}

			// Check variance for each component
			// For uniform distribution on unit sphere, variance of each component ≈ 1/3
			varX := sumX2/samples - meanX*meanX
			varY := sumY2/samples - meanY*meanY
			varZ := sumZ2/samples - meanZ*meanZ

			expectedVar := 1.0 / 3.0
			const varTolerance = 0.01 // Allow 1% deviation

			if math.Abs(varX-expectedVar) > varTolerance {
				t.Errorf("variance X = %.6f, want ≈%.6f (within %.6f)",
					varX, expectedVar, varTolerance)
			}
			if math.Abs(varY-expectedVar) > varTolerance {
				t.Errorf("variance Y = %.6f, want ≈%.6f (within %.6f)",
					varY, expectedVar, varTolerance)
			}
			if math.Abs(varZ-expectedVar) > varTolerance {
				t.Errorf("variance Z = %.6f, want ≈%.6f (within %.6f)",
					varZ, expectedVar, varTolerance)
			}

			// Check octant distribution
			// Each octant should contain approximately samples/8 points
			expectedPerOctant := samples / 8
			// Allow 15% deviation from expected
			octantTolerance := float64(expectedPerOctant) * 0.15

			for octant, count := range octantCounts {
				diff := math.Abs(float64(count) - float64(expectedPerOctant))
				if diff > octantTolerance {
					t.Errorf("octant %d: count = %d, want ≈%d (within %.0f)",
						octant, count, expectedPerOctant, octantTolerance)
				}
			}
		})
	}
}

// TestRandomUnitVectorNoBias specifically tests for known biases:
// - Angle method: check that z-coordinate distribution is uniform.
// - Rejection method: check rejection rate is reasonable.
func TestRandomUnitVectorNoBias(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping bias test in short mode")
	}

	t.Run("AngleMethod_ZDistribution", func(t *testing.T) {
		const samples = 10000
		const bins = 20
		histogram := make([]int, bins)

		for range samples {
			v := RandomUnitVectorAngle[Vec3]()
			// z is in [-1, 1], map to bin [0, bins-1]
			bin := int((v[2] + 1) / 2 * float64(bins))
			if bin >= bins {
				bin = bins - 1
			}
			if bin < 0 {
				bin = 0
			}
			histogram[bin]++
		}

		// Each bin should contain approximately samples/bins points
		expectedPerBin := samples / bins
		// Chi-square test would be more rigorous, but simple tolerance works
		tolerance := float64(expectedPerBin) * 0.20 // 20% tolerance

		for bin, count := range histogram {
			diff := math.Abs(float64(count) - float64(expectedPerBin))
			if diff > tolerance {
				t.Errorf("z-distribution bin %d: count = %d, want ≈%d (within %.0f)",
					bin, count, expectedPerBin, tolerance)
			}
		}
	})

	t.Run("RejectionMethod_ReasonableAcceptance", func(*testing.T) {
		// The rejection method should accept points in a sphere inscribed in a cube
		// Volume of sphere / volume of cube = (4/3)πr³ / (2r)³ = π/6 ≈ 0.524
		// So we expect roughly 52% acceptance rate
		// This is a smoke test, not a statistical test
		const samples = 1000
		for range samples {
			_ = RandomUnitVectorRej[Vec3]()
		}
		// If this hangs or takes too long, there's a problem with the rejection logic
		// The test passing means it completed in reasonable time
	})
}

// Benchmarks for comparing the three methods

func BenchmarkRandomUnitVectorRejection(b *testing.B) {
	for range b.N {
		_ = RandomUnitVectorRej[Vec3]()
	}
}

func BenchmarkRandomUnitVectorAngle(b *testing.B) {
	for range b.N {
		_ = RandomUnitVectorAngle[Vec3]()
	}
}

func BenchmarkRandomUnitVectorNorm(b *testing.B) {
	for range b.N {
		_ = RandomUnitVector[Vec3]()
	}
}
