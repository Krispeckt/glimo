package patterns

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// Color Representation and Utilities

// Color represents a simple 8-bit per channel RGBA color.
// It includes an optional BlendMode field used for compositing.
type Color struct {
	R, G, B, A uint8
	blendMode  BlendMode
}

// BlendMode returns the current blending mode assigned to the color.
func (c Color) BlendMode() BlendMode {
	return c.blendMode
}

// NewColorFromStd converts a standard color.Color into a Color type.
func NewColorFromStd(c color.Color) Color {
	r, g, b, a := c.RGBA()
	return Color{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

// RGBA returns 16-bit per channel alpha-premultiplied color components.
func (c Color) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = uint32(c.A)
	a |= a << 8
	return
}

// RGBA64 returns 64-bit (non-premultiplied) color channel values.
func (c Color) RGBA64() (r, g, b, a uint64) {
	r = uint64(c.R)
	r |= r << 8
	g = uint64(c.G)
	g |= g << 8
	b = uint64(c.B)
	b |= b << 8
	a = uint64(c.A)
	a |= a << 8
	return
}

// HEX Conversion

// ToHex returns the color as a hexadecimal string in #RRGGBB or #RRGGBBAA format.
func (c Color) ToHex() string {
	if c.A == 255 {
		return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
	}
	return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A)
}

// ColorFromHex parses a hexadecimal color string (#RGB, #RRGGBB, or #RRGGBBAA)
// and returns a corresponding Color value.
func ColorFromHex(hex string) (Color, error) {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b, a uint8 = 0, 0, 0, 255

	switch len(hex) {
	case 3:
		_, err := fmt.Sscanf(hex, "%1x%1x%1x", &r, &g, &b)
		if err != nil {
			return Color{}, err
		}
		r, g, b = r*17, g*17, b*17
	case 6:
		_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return Color{}, err
		}
	case 8:
		_, err := fmt.Sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a)
		if err != nil {
			return Color{}, err
		}
	default:
		return Color{}, fmt.Errorf("invalid hex color format: %s", hex)
	}
	return Color{R: r, G: g, B: b, A: a}, nil
}

// HSL Conversion

// ToHSL converts the color to HSL representation.
// Returns hue in [0–360], saturation in [0–1], and lightness in [0–1].
func (c Color) ToHSL() (h, s, l float64) {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255

	maximum := math.Max(r, math.Max(g, b))
	minimum := math.Min(r, math.Min(g, b))
	l = (maximum + minimum) / 2

	if maximum == minimum {
		h, s = 0, 0
	} else {
		d := maximum - minimum
		if l > 0.5 {
			s = d / (2 - maximum - minimum)
		} else {
			s = d / (maximum + minimum)
		}

		switch maximum {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6
			}
		case g:
			h = (b-r)/d + 2
		case b:
			h = (r-g)/d + 4
		}
		h *= 60
	}
	return
}

// ColorFromHSL creates a Color from HSL and alpha values.
func ColorFromHSL(h, s, l float64, a uint8) Color {
	h = math.Mod(h, 360)
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64
	switch {
	case 0 <= h && h < 60:
		r, g, b = c, x, 0
	case 60 <= h && h < 120:
		r, g, b = x, c, 0
	case 120 <= h && h < 180:
		r, g, b = 0, c, x
	case 180 <= h && h < 240:
		r, g, b = 0, x, c
	case 240 <= h && h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return Color{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: a,
	}
}

// Conversion and Pattern Helpers

// ToColor converts the custom Color type into a standard color.RGBA.
func (c Color) ToColor() color.RGBA {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

// MakeSolidPattern creates a solid fill pattern from this color,
// preserving its blend mode and full opacity.
func (c Color) MakeSolidPattern() BlendedPattern {
	return NewSolidWithBlend(c, c.blendMode, 1)
}

// MakeLinearGradient creates a linear gradient pattern from this color to another.
func (c Color) MakeLinearGradient(x0, y0, x1, y1 float64, to Color) GradientPattern {
	g := NewLinearGradient(x0, y0, x1, y1)
	g.AddColorStop(0, c)
	g.AddColorStop(1, to)
	return g
}

// MakeRadialGradient creates a radial gradient transitioning from this color to another.
func (c Color) MakeRadialGradient(cx0, cy0, r0, cx1, cy1, r1 float64, to Color) GradientPattern {
	g := NewRadialGradient(cx0, cy0, r0, cx1, cy1, r1)
	g.AddColorStop(0, c)
	g.AddColorStop(1, to)
	return g
}

// MakeConicGradient creates a conic gradient from this color to another,
// centered at (cx, cy) and sweeping `deg` degrees.
func (c Color) MakeConicGradient(cx, cy, deg float64, to Color) GradientPattern {
	g := NewConicGradient(cx, cy, deg)
	g.AddColorStop(0, c)
	g.AddColorStop(1, to)
	return g
}

// Blending and Compositing

// Mix performs gamma-corrected interpolation between two colors in sRGB space.
func (c Color) Mix(other Color, t float64) Color {
	t = geom.ClampF64(t, 0, 1)

	ra := geom.SrgbToLinear8(c.R)
	ga := geom.SrgbToLinear8(c.G)
	ba := geom.SrgbToLinear8(c.B)

	rb := geom.SrgbToLinear8(other.R)
	gb := geom.SrgbToLinear8(other.G)
	bb := geom.SrgbToLinear8(other.B)

	r := (1-t)*ra + t*rb
	g := (1-t)*ga + t*gb
	b := (1-t)*ba + t*bb

	return Color{
		R: geom.LinearToSRGB8(r),
		G: geom.LinearToSRGB8(g),
		B: geom.LinearToSRGB8(b),
		A: uint8(math.Round((1-t)*float64(c.A) + t*float64(other.A))),
	}
}

// Over performs Porter–Duff alpha compositing (source-over-destination).
// It blends the foreground color `c` over the background color `bg`.
func (c Color) Over(bg Color) Color {
	af := float64(c.A) / 255.0
	ab := float64(bg.A) / 255.0
	outA := af + ab*(1-af)
	if outA == 0 {
		return Color{}
	}

	rf := geom.SrgbToLinear8(c.R) * af
	gf := geom.SrgbToLinear8(c.G) * af
	bf := geom.SrgbToLinear8(c.B) * af

	rb := geom.SrgbToLinear8(bg.R) * ab
	gb := geom.SrgbToLinear8(bg.G) * ab
	bb := geom.SrgbToLinear8(bg.B) * ab

	r := (rf + rb*(1-af)) / outA
	g := (gf + gb*(1-af)) / outA
	b := (bf + bb*(1-af)) / outA

	return Color{
		R: geom.LinearToSRGB8(r),
		G: geom.LinearToSRGB8(g),
		B: geom.LinearToSRGB8(b),
		A: uint8(math.Round(outA * 255)),
	}
}

// SetBlendMode assigns a new blending mode to the color.
func (c Color) SetBlendMode(mode BlendMode) Color {
	c.blendMode = mode
	return c
}

// SetOpacity sets the color’s alpha channel from a normalized opacity value [0–1].
func (c Color) SetOpacity(opacity float64) Color {
	opacity = geom.ClampF64(opacity, 0, 1)
	c.A = uint8(math.Round(opacity * 255))
	return c
}
