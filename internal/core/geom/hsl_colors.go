package geom

import "math"

// Lum returns the perceived luminance of an RGB triple in [0,1].
func Lum(C [3]float64) float64 { return 0.3*C[0] + 0.59*C[1] + 0.11*C[2] }

// Sat returns the saturation (max - min channel) of an RGB triple.
func Sat(C [3]float64) float64 {
	maxc := math.Max(C[0], math.Max(C[1], C[2]))
	minc := math.Min(C[0], math.Min(C[1], C[2]))
	return maxc - minc
}

// ClipColor constrains RGB components to [0,1]
// while preserving overall luminance balance.
func ClipColor(C [3]float64) [3]float64 {
	L := Lum(C)
	minc := math.Min(C[0], math.Min(C[1], C[2]))
	maxc := math.Max(C[0], math.Max(C[1], C[2]))
	if minc < 0 {
		for i := 0; i < 3; i++ {
			C[i] = L + ((C[i]-L)*L)/(L-minc)
		}
	}
	if maxc > 1 {
		for i := 0; i < 3; i++ {
			C[i] = L + ((C[i]-L)*(1-L))/(maxc-L)
		}
	}
	return C
}

// SetLum sets target luminance for an RGB triple and clips the result.
func SetLum(C [3]float64, l float64) [3]float64 {
	d := l - Lum(C)
	C[0] += d
	C[1] += d
	C[2] += d
	return ClipColor(C)
}

// SetSat sets target saturation for an RGB triple, preserving hue.
func SetSat(C [3]float64, s float64) [3]float64 {
	type kv struct {
		v float64
		i int
	}
	a := []kv{{C[0], 0}, {C[1], 1}, {C[2], 2}}

	// Sort by ascending channel value
	if a[0].v > a[1].v {
		a[0], a[1] = a[1], a[0]
	}
	if a[1].v > a[2].v {
		a[1], a[2] = a[2], a[1]
	}
	if a[0].v > a[1].v {
		a[0], a[1] = a[1], a[0]
	}

	Cmin, Cmid, Cmax := a[0], a[1], a[2]
	out := C
	if Cmax.v > Cmin.v {
		out[Cmid.i] = ((Cmid.v - Cmin.v) * s) / (Cmax.v - Cmin.v)
		out[Cmax.i] = s
	} else {
		out[Cmid.i], out[Cmax.i] = 0, 0
	}
	out[Cmin.i] = 0
	return out
}

// U8 converts a normalized float [0,1] to 8-bit (0–255) with rounding.
func U8(v float64) uint8 { return uint8(math.Round(ClampF64(v, 0, 1) * 255)) }
