package geom

import (
	"image"
	"image/color"
	"math"
)

// Color Space Conversion

// SrgbToLinear8 converts an 8-bit sRGB channel value (0–255) to a linear float64 value (0–1).
func SrgbToLinear8(u uint8) float64 {
	c := float64(u) / 255.0
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// LinearToSRGB8 converts a linear float64 value (0–1) to an 8-bit sRGB channel value (0–255).
func LinearToSRGB8(f float64) uint8 {
	if f <= 0 {
		return 0
	}
	if f >= 1 {
		return 255
	}
	var c float64
	if f <= 0.0031308 {
		c = 12.92 * f
	} else {
		c = 1.055*math.Pow(f, 1.0/2.4) - 0.055
	}
	return uint8(math.Round(c * 255))
}

// SrgbToLinear converts an sRGB channel value (0–1) to linear light space (0–1).
func SrgbToLinear(c float64) float64 {
	c = ClampF64(c, 0, 1)
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// LinearToSrgb converts a linear light channel value (0–1) to sRGB (0–1).
func LinearToSrgb(c float64) float64 {
	c = ClampF64(c, 0, 1)
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1.0/2.4) - 0.055
}

// Color Interpolation

// LerpColor performs linear interpolation between two colors c1 and c2.
// Parameter t defines the interpolation amount, where 0 = c1 and 1 = c2.
func LerpColor(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	r := uint8(Lerp(float64(r1>>8), float64(r2>>8), t))
	g := uint8(Lerp(float64(g1>>8), float64(g2>>8), t))
	b := uint8(Lerp(float64(b1>>8), float64(b2>>8), t))
	a := uint8(Lerp(float64(a1>>8), float64(a2>>8), t))

	return color.NRGBA{R: r, G: g, B: b, A: a}
}

// GetColor returns a color interpolated from a set of color stops.
// It finds the interval containing t and interpolates between its two surrounding colors.
func GetColor(t float64, stops Stops) color.Color {
	if len(stops) == 1 {
		return stops[0].Color()
	}

	for i := 1; i < len(stops); i++ {
		if t <= stops[i].Position() {
			p0, p1 := stops[i-1], stops[i]
			f := Norm(t, p0.Position(), p1.Position())
			return LerpColor(p0.Color(), p1.Color(), f)
		}
	}

	return stops[len(stops)-1].Color()
}

// Processing

// BilinearRGBAAt performs bilinear interpolation on an RGBA image at floating-point coordinates (fx, fy).
// If the coordinates fall outside the image bounds, it returns the provided background color.
func BilinearRGBAAt(src *image.RGBA, fx, fy float64, bg color.Color) color.RGBA {
	b := src.Bounds()
	minX, minY := b.Min.X, b.Min.Y
	maxX, maxY := b.Max.X-1, b.Max.Y-1

	if fx < float64(minX) || fx > float64(maxX) || fy < float64(minY) || fy > float64(maxY) {
		r, g, b, a := bg.RGBA()
		return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
	}

	x0 := int(math.Floor(fx))
	y0 := int(math.Floor(fy))
	x1 := ClampInt(x0+1, minX, maxX)
	y1 := ClampInt(y0+1, minY, maxY)

	wx := fx - float64(x0)
	wy := fy - float64(y0)

	c00 := src.RGBAAt(x0, y0)
	c10 := src.RGBAAt(x1, y0)
	c01 := src.RGBAAt(x0, y1)
	c11 := src.RGBAAt(x1, y1)

	w00 := (1 - wx) * (1 - wy)
	w10 := wx * (1 - wy)
	w01 := (1 - wx) * wy
	w11 := wx * wy

	r := float64(c00.R)*w00 + float64(c10.R)*w10 + float64(c01.R)*w01 + float64(c11.R)*w11
	g := float64(c00.G)*w00 + float64(c10.G)*w10 + float64(c01.G)*w01 + float64(c11.G)*w11
	bl := float64(c00.B)*w00 + float64(c10.B)*w10 + float64(c01.B)*w01 + float64(c11.B)*w11
	a := float64(c00.A)*w00 + float64(c10.A)*w10 + float64(c01.A)*w01 + float64(c11.A)*w11

	return color.RGBA{
		R: uint8(math.Round(r)),
		G: uint8(math.Round(g)),
		B: uint8(math.Round(bl)),
		A: uint8(math.Round(a)),
	}
}
