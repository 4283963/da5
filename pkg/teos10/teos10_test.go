package teos10

import "testing"

func TestNoPanicOnInvalidInput(t *testing.T) {
	calc := NewTEOS10Calculator(30.0)

	testCases := []struct {
		name         string
		conductivity float64
		temperature  float64
		pressure     float64
	}{
		{"negative pressure", 42.914, 15.0, -100.0},
		{"zero pressure", 42.914, 15.0, 0.0},
		{"negative conductivity", -5.0, 15.0, 100.0},
		{"zero conductivity", 0.0, 15.0, 100.0},
		{"very low temperature", 42.914, -100.0, 100.0},
		{"extreme pressure", 42.914, 15.0, 100000.0},
		{"extreme pressure negative", 42.914, 15.0, -100000.0},
		{"all negative", -10.0, -50.0, -100.0},
		{"huge conductivity", 1e6, 15.0, 100.0},
		{"huge negative conductivity", -1e6, 15.0, 100.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic occurred for %s: %v", tc.name, r)
				}
			}()

			result := calc.Calibrate(tc.conductivity, tc.temperature, tc.pressure)

			if result.Salinity < 0 {
				t.Errorf("%s: salinity = %f, expected >= 0", tc.name, result.Salinity)
			}
			if result.Depth < 0 {
				t.Errorf("%s: depth = %f, expected >= 0", tc.name, result.Depth)
			}
			if result.Density < 0 {
				t.Errorf("%s: density = %f, expected >= 0", tc.name, result.Density)
			}
		})
	}
}

func TestValidInputStillWorks(t *testing.T) {
	calc := NewTEOS10Calculator(30.0)

	result := calc.Calibrate(42.914, 15.0, 0.0)

	if result.Salinity < 34.9 || result.Salinity > 35.1 {
		t.Errorf("standard seawater salinity should be ~35, got %f", result.Salinity)
	}

	t.Logf("Standard seawater test: salinity=%.6f, depth=%.6f, density=%.6f",
		result.Salinity, result.Depth, result.Density)
}
