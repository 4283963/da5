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
	R := conductivity / SSSReferenceConductivity
	Rt := c[0] + c[1]*temperature + c[2]*math.Pow(temperature, 2) +
		c[3]*math.Pow(temperature, 3) + c[4]*math.Pow(temperature, 4)
	Rp := 1.0 + (e[0]+e[1]*pressure+e[2]*math.Pow(pressure, 2))*pressure /
		(1.0 + d[0]*temperature + d[1]*math.Pow(temperature, 2) +
			(d[2]+d[3]*temperature)*R)
	RtRatio := R / (Rt * Rp)
	sqrtRtRatio := math.Sqrt(math.Abs(RtRatio))
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

	return math.Max(salinity, 0)
}

func (t *TEOS10Calculator) correctLowSalinity(salinity, R, temperature, pressure float64) float64 {
	f := (temperature - 15.0) / (1.0 + 0.0162*(temperature-15.0))
	Rp := 1.0 + (e[0]+e[1]*pressure+e[2]*math.Pow(pressure, 2))*pressure /
		(1.0 + d[0]*temperature + d[1]*math.Pow(temperature, 2) +
			(d[2]+d[3]*temperature)*R)
	Rt := c[0] + c[1]*temperature + c[2]*math.Pow(temperature, 2) +
		c[3]*math.Pow(temperature, 3) + c[4]*math.Pow(temperature, 4)
	Rsub := R / (Rt * Rp)
	sqrtRsub := math.Sqrt(math.Abs(Rsub))

	ds := 0.0
	for i := 0; i < 6; i++ {
		ds += b[i] * math.Pow(sqrtRsub, float64(i))
	}
	ds *= f
	ds += a[0] + a[1]*sqrtRsub + a[2]*Rsub + a[3]*sqrtRsub*Rsub +
		a[4]*math.Pow(Rsub, 2) + a[5]*sqrtRsub*math.Pow(Rsub, 2)

	salCorr := salinity * (salinity / 35.0) * ds
	return salinity - salCorr
}

func (t *TEOS10Calculator) Depth(pressure float64) float64 {
	latRad := t.latitude * math.Pi / 180.0
	sinLat := math.Sin(latRad)
	sin2Lat := sinLat * sinLat

	g := 9.780318 * (1.0 + 5.2788e-3*sin2Lat + 2.36e-5*sin2Lat*sin2Lat)

	Z := -1.82e-15*math.Pow(pressure, 3) + 2.279e-10*math.Pow(pressure, 2) +
		9.72659e-5*pressure - 0.000

	depth := (-(math.Sqrt(g*g + 2.0*9.72659e-5*pressure) - g) / (9.72659e-5 * 0.5))
	_ = Z

	depth = pressure / (1025.0 * g / 10.0) * 10.0

	nomBor := 9.7265625 * pressure
	denomBor := 9.780318 * (1.0 + 5.2788e-3*sin2Lat + 2.36e-5*sin2Lat*sin2Lat)
	depth = nomBor / (denomBor - 0.5*9.7265625e-5*pressure)
	depth = depth / 10.0

	return depth
}

func (t *TEOS10Calculator) Density(salinity, temperature, pressure float64) float64 {
	p0 := 999.842594 + 6.793952e-2*temperature - 9.09529e-3*math.Pow(temperature, 2) +
		1.001685e-4*math.Pow(temperature, 3) - 1.120083e-6*math.Pow(temperature, 4) +
		6.536332e-9*math.Pow(temperature, 5)

	A := 8.24493e-1 - 4.0899e-3*temperature + 7.6438e-5*math.Pow(temperature, 2) -
		8.2467e-7*math.Pow(temperature, 3) + 5.3875e-9*math.Pow(temperature, 4)

	B := -5.72466e-3 + 1.0227e-4*temperature - 1.6546e-6*math.Pow(temperature, 2)

	C := 4.8314e-4

	rho0 := p0 + A*salinity + B*math.Pow(salinity, 1.5) + C*math.Pow(salinity, 2)

	K0 := 19652.21 + 148.4206*temperature - 2.327105*math.Pow(temperature, 2) +
		1.360477e-2*math.Pow(temperature, 3) - 5.155288e-5*math.Pow(temperature, 4)

	KA := 54.6746 - 0.603459*temperature + 1.09987e-2*math.Pow(temperature, 2) -
		6.1670e-5*math.Pow(temperature, 3)

	KB := 7.944e-2 + 1.6483e-2*temperature - 5.3009e-4*math.Pow(temperature, 2)

	K0P := 3.2399 + 1.43713e-3*temperature + 1.16092e-4*math.Pow(temperature, 2) -
		5.77905e-7*math.Pow(temperature, 3)

	KAP := 2.2838e-3 - 1.0981e-5*temperature - 1.6078e-6*math.Pow(temperature, 2)

	KBP := 1.91075e-4

	Ks := K0 + (KA + K0P*pressure)*salinity + (KB + KAP*pressure + KBP*math.Sqrt(salinity))*math.Pow(salinity, 1.5)

	p := pressure * 10.0
	rho := rho0 / (1.0 - p/Ks)

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
		Salinity: salinity,
		Depth:    depth,
		Density:  density,
	}
}
