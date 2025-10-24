package geom

import (
	"math"

	"golang.org/x/image/math/fixed"
)

// Norm maps a value x from range [a, b] to a normalized range [0, 1].
func Norm(x, a, b float64) float64 {
	return (x - a) / (b - a)
}

// Lerp performs linear interpolation between a and b using t in [0, 1].
func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// ClampF64 constrains x to stay within the range [lo, hi].
func ClampF64(x, lo, hi float64) float64 {
	return math.Min(math.Max(x, lo), hi)
}

// Dot3 computes the dot product of two 3D vectors (x0, y0, z0) and (x1, y1, z1).
func Dot3(x0, y0, z0, x1, y1, z1 float64) float64 {
	return x0*x1 + y0*y1 + z0*z1
}

// Quant64 rounds a floating-point coordinate to the nearest 1/64 pixel.
// Used to stabilize vector graphics rendering and eliminate subpixel jitter.
func Quant64(v float64) float64 {
	return math.Round(v*64.0) / 64.0
}

// MaxInt returns the greater of two integers.
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MaxF64 returns the greater of two doubles.
func MaxF64(a, b float64) float64 {
	return math.Max(a, b)
}

// NormalizeAngle normalizes an angle in degrees to the range [0, 360).
func NormalizeAngle(t float64) float64 {
	t = math.Mod(t, 360)
	if t < 0 {
		t += 360
	}
	return t
}

// Fixed-Point Arithmetic

// Unfix converts a fixed.Int26_6 value (1/64 fractional precision) to float64.
func Unfix(x fixed.Int26_6) float64 {
	const shift, mask = 6, 1<<6 - 1
	if x >= 0 {
		return float64(x>>shift) + float64(x&mask)/64
	}
	x = -x
	if x >= 0 {
		return -(float64(x>>shift) + float64(x&mask)/64)
	}
	return 0
}

// Fix converts a float64 value to fixed.Int26_6 (1/64 pixel precision).
func Fix(x float64) fixed.Int26_6 {
	return fixed.Int26_6(math.Round(x * 64))
}

// Integer and Byte Utilities

// Mul255 multiplies two 8-bit values (0–255), divides by 255, and rounds correctly.
// Commonly used for color blending and alpha compositing.
func Mul255(a, b uint8) uint8 {
	return uint8((int(a)*int(b) + 127) / 255)
}

// ClampInt constrains v to stay within the range [lo, hi].
func ClampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
