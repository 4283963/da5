package teos10

import "math"

const (
	StandardSeaPressure    = 101325.0
	SSSReferenceConductivity = 42.9140
	TemperatureReference    = 15.0
	PressureReference       = 0.0
	GravitationalAccel     = 9.7963
	LatitudeDefault        = 30.0
)

var (
	a = []float64{0.0080, -0.1692, 25.3851, 14.0941, -7.0261, 2.7081}
	b = []float64{0.0005, -0.0056, -0.0066, -0.0375, 0.0636, -0.0144}
	c = []float64{0.6766097, 0.0200564, 0.0001104259, -6.9698e-7, 1.0031e-9}
	d = []float64{0.03426, 0.0004464, 0.4215, -0.003107}
	e = []float64{2.070e-5, -6.370e-10, 3.989e-15}
)

type TEOS10Calculator struct {
	latitude float64
}

func NewTEOS10Calculator(latitude float64) *TEOS10Calculator {
	if latitude == 0 {
		latitude = LatitudeDefault
	}
	return &TEOS10Calculator{latitude: latitude}
}

func (t *TEOS10Calculator) PracticalSalinity(conductivity, temperature, pressure float64) float64 {
	if conductivity <= 0 {
		return 0
	}

	R := conductivity / SSSReferenceConductivity
	if R <= 0 {
		return 0
	}

	Rt := c[0] + c[1]*temperature + c[2]*math.Pow(temperature, 2) +
		c[3]*math.Pow(temperature, 3) + c[4]*math.Pow(temperature, 4)
	if Rt <= 0 {
		return 0
	}

	Rp := 1.0 + (e[0]+e[1]*pressure+e[2]*math.Pow(pressure, 2))*pressure /
		(1.0 + d[0]*temperature + d[1]*math.Pow(temperature, 2) +
			(d[2]+d[3]*temperature)*R)
	if Rp <= 0 {
		return 0
	}

	RtRatio := R / (Rt * Rp)
	if RtRatio <= 0 {
		return 0
	}

	sqrtRtRatio := math.Sqrt(RtRatio)
	delTemperature := temperature - 15.0
	delTerm := delTemperature / (1.0 + 0.0162*delTemperature)

	salinity := 0.0
	for i := 0; i < 6; i++ {
		salinity += a[i] * math.Pow(sqrtRtRatio, float64(i))
	}
	for i := 0; i < 6; i++ {
		salinity += delTerm * b[i] * math.Pow(sqrtRtRatio, float64(i))
	}

	if salinity < 2 {
		salinity = t.correctLowSalinity(salinity, R, temperature, pressure)
	}

	if salinity < 0 || math.IsNaN(salinity) || math.IsInf(salinity, 0) {
		return 0
	}

	return salinity
}

func (t *TEOS10Calculator) correctLowSalinity(salinity, R, temperature, pressure float64) float64 {
	if salinity <= 0 || R <= 0 {
		return 0
	}

	f := (temperature - 15.0) / (1.0 + 0.0162*(temperature-15.0))
	Rp := 1.0 + (e[0]+e[1]*pressure+e[2]*math.Pow(pressure, 2))*pressure /
		(1.0 + d[0]*temperature + d[1]*math.Pow(temperature, 2) +
			(d[2]+d[3]*temperature)*R)
	if Rp <= 0 {
		return salinity
	}

	Rt := c[0] + c[1]*temperature + c[2]*math.Pow(temperature, 2) +
		c[3]*math.Pow(temperature, 3) + c[4]*math.Pow(temperature, 4)
	if Rt <= 0 {
		return salinity
	}

	Rsub := R / (Rt * Rp)
	if Rsub <= 0 {
		return salinity
	}

	sqrtRsub := math.Sqrt(Rsub)

	ds := 0.0
	for i := 0; i < 6; i++ {
		ds += b[i] * math.Pow(sqrtRsub, float64(i))
	}
	ds *= f
	ds += a[0] + a[1]*sqrtRsub + a[2]*Rsub + a[3]*sqrtRsub*Rsub +
		a[4]*math.Pow(Rsub, 2) + a[5]*sqrtRsub*math.Pow(Rsub, 2)

	salCorr := salinity * (salinity / 35.0) * ds
	result := salinity - salCorr

	if result < 0 || math.IsNaN(result) || math.IsInf(result, 0) {
		return 0
	}

	return result
}

func (t *TEOS10Calculator) Depth(pressure float64) float64 {
	if pressure <= 0 {
		return 0
	}

	latRad := t.latitude * math.Pi / 180.0
	sinLat := math.Sin(latRad)
	sin2Lat := sinLat * sinLat

	g := 9.780318 * (1.0 + 5.2788e-3*sin2Lat + 2.36e-5*sin2Lat*sin2Lat)

	c := 9.7265625
	sqrtTerm := g*g + 2.0*g*c*pressure/1000.0
	if sqrtTerm <= 0 {
		return 0
	}

	depth := (-(math.Sqrt(sqrtTerm) - g) / (c * 0.5 / 1000.0))

	nomBor := c * pressure
	denomBor := g - 0.5*c*pressure/1000.0
	if denomBor <= 0 {
		return 0
	}

	depth = nomBor / denomBor
	depth = depth / 10.0

	if depth < 0 || math.IsNaN(depth) || math.IsInf(depth, 0) {
		return 0
	}

	return depth
}

func (t *TEOS10Calculator) Density(salinity, temperature, pressure float64) float64 {
	if salinity < 0 {
		salinity = 0
	}

	p0 := 999.842594 + 6.793952e-2*temperature - 9.09529e-3*math.Pow(temperature, 2) +
		1.001685e-4*math.Pow(temperature, 3) - 1.120083e-6*math.Pow(temperature, 4) +
		6.536332e-9*math.Pow(temperature, 5)

	A := 8.24493e-1 - 4.0899e-3*temperature + 7.6438e-5*math.Pow(temperature, 2) -
		8.2467e-7*math.Pow(temperature, 3) + 5.3875e-9*math.Pow(temperature, 4)

	B := -5.72466e-3 + 1.0227e-4*temperature - 1.6546e-6*math.Pow(temperature, 2)

	C := 4.8314e-4

	sqrtSal := math.Sqrt(salinity)
	sal15 := salinity * sqrtSal
	rho0 := p0 + A*salinity + B*sal15 + C*salinity*salinity

	K0 := 19652.21 + 148.4206*temperature - 2.327105*math.Pow(temperature, 2) +
		1.360477e-2*math.Pow(temperature, 3) - 5.155288e-5*math.Pow(temperature, 4)

	KA := 54.6746 - 0.603459*temperature + 1.09987e-2*math.Pow(temperature, 2) -
		6.1670e-5*math.Pow(temperature, 3)

	KB := 7.944e-2 + 1.6483e-2*temperature - 5.3009e-4*math.Pow(temperature, 2)

	K0P := 3.2399 + 1.43713e-3*temperature + 1.16092e-4*math.Pow(temperature, 2) -
		5.77905e-7*math.Pow(temperature, 3)

	KAP := 2.2838e-3 - 1.0981e-5*temperature - 1.6078e-6*math.Pow(temperature, 2)

	KBP := 1.91075e-4

	Ks := K0 + (KA + K0P*pressure)*salinity + (KB + KAP*pressure + KBP*sqrtSal)*sal15

	p := pressure * 10.0
	denom := 1.0 - p/Ks
	if math.Abs(denom) < 1e-10 {
		return 0
	}

	rho := rho0 / denom

	if rho < 0 || math.IsNaN(rho) || math.IsInf(rho, 0) {
		return 0
	}

	return rho
}

type CalibrationResult struct {
	Salinity float64
	Depth    float64
	Density  float64
}

func (t *TEOS10Calculator) Calibrate(conductivity, temperature, pressure float64) CalibrationResult {
	salinity := t.PracticalSalinity(conductivity, temperature, pressure)
	depth := t.Depth(pressure)
	density := t.Density(salinity, temperature, pressure)

	return CalibrationResult{
		Salinity: math.Max(salinity, 0),
		Depth:    math.Max(depth, 0),
		Density:  math.Max(density, 0),
	}
}
