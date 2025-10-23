package patterns

import (
	"math"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// Color Blending and Porter–Duff Composition

// BlendMode defines a color blending operation between a source (foreground)
// and destination (background) color, following CSS and Porter–Duff conventions.
type BlendMode uint8

// Supported blending modes, mostly aligned with CSS Compositing and Blending Level 1.
// Non-standard aliases are provided for Photoshop-style linear modes.
const (
	BlendPassThrough BlendMode = iota
	BlendNormal
	BlendDarken
	BlendMultiply
	BlendPlusDarker // Alias of LinearBurn (non-CSS)
	BlendColorBurn
	BlendLighten
	BlendScreen
	BlendPlusLighter // Alias of LinearDodge (non-CSS)
	BlendColorDodge
	BlendOverlay
	BlendSoftLight
	BlendHardLight
	BlendDifference
	BlendExclusion
	BlendHue
	BlendSaturation
	BlendColor
	BlendLuminosity

	BlendLinearBurn  = BlendPlusDarker
	BlendLinearDodge = BlendPlusLighter
)

// String returns a string representation of the blending mode.
func (m BlendMode) String() string {
	switch m {
	case BlendPassThrough:
		return "PassThrough"
	case BlendNormal:
		return "Normal"
	case BlendDarken:
		return "Darken"
	case BlendMultiply:
		return "Multiply"
	case BlendPlusDarker:
		return "PlusDarker"
	case BlendColorBurn:
		return "ColorBurn"
	case BlendLighten:
		return "Lighten"
	case BlendScreen:
		return "Screen"
	case BlendPlusLighter:
		return "PlusLighter"
	case BlendColorDodge:
		return "ColorDodge"
	case BlendOverlay:
		return "Overlay"
	case BlendSoftLight:
		return "SoftLight"
	case BlendHardLight:
		return "HardLight"
	case BlendDifference:
		return "Difference"
	case BlendExclusion:
		return "Exclusion"
	case BlendHue:
		return "Hue"
	case BlendSaturation:
		return "Saturation"
	case BlendColor:
		return "Color"
	case BlendLuminosity:
		return "Luminosity"
	default:
		return "Unknown"
	}
}

// BlendOver: main blending function

// BlendOver blends the current color (`c`) over a background (`bg`) with a given opacity,
// using the selected blend mode.
//
// Internally, all blending is done in **linear RGB space** with proper sRGB <-> linear conversion.
// The output color is returned in sRGB space.
//
// Steps:
//  1. Decode sRGB (8-bit → float64 [0–1])
//  2. Convert to linear RGB
//  3. Apply the CSS blend mode
//  4. Combine using Porter–Duff SRC-OVER formula
//  5. Convert back to sRGB for output
func (c Color) BlendOver(bg Color, opacity float64) Color {
	mode := c.blendMode
	if mode == BlendPassThrough {
		mode = BlendNormal
	}
	opacity = geom.ClampF64(opacity, 0, 1)

	// Convert from sRGB [0–255] → [0–1]
	cbS := [3]float64{float64(bg.R) / 255, float64(bg.G) / 255, float64(bg.B) / 255}
	csS := [3]float64{float64(c.R) / 255, float64(c.G) / 255, float64(c.B) / 255}

	// Convert to linear space
	cb := [3]float64{geom.SrgbToLinear(cbS[0]), geom.SrgbToLinear(cbS[1]), geom.SrgbToLinear(cbS[2])}
	cs := [3]float64{geom.SrgbToLinear(csS[0]), geom.SrgbToLinear(csS[1]), geom.SrgbToLinear(csS[2])}

	ab := float64(bg.A) / 255.0
	as := float64(c.A) / 255.0 * opacity

	// Apply blending mode in sRGB domain, result converted back to linear
	var BcolL [3]float64
	switch mode {
	case BlendHue, BlendSaturation, BlendColor, BlendLuminosity:
		BcolL = nonSeparableB(cb, cs, mode) // Non-separable modes use HSL operations
	default:
		BcolL[0] = blendChannel(cb[0], cs[0], mode)
		BcolL[1] = blendChannel(cb[1], cs[1], mode)
		BcolL[2] = blendChannel(cb[2], cs[2], mode)
	}
	for i := 0; i < 3; i++ {
		BcolL[i] = geom.ClampF64(BcolL[i], 0, 1)
	}

	// Intermediate compositing:
	// Cs' = (1 - Ab) * Cs + Ab * B(Cs, Cb)
	csPrime := [3]float64{
		(1-ab)*cs[0] + ab*BcolL[0],
		(1-ab)*cs[1] + ab*BcolL[1],
		(1-ab)*cs[2] + ab*BcolL[2],
	}

	// Porter–Duff "Source Over" (SRC_OVER):
	// Co = As * Cs' + Ab * (1 - As) * Cb
	// Ao = As + Ab * (1 - As)
	co := [3]float64{
		as*csPrime[0] + ab*(1-as)*cb[0],
		as*csPrime[1] + ab*(1-as)*cb[1],
		as*csPrime[2] + ab*(1-as)*cb[2],
	}
	ao := as + ab*(1-as)
	if ao <= 0 {
		return Color{}
	}

	// Un-premultiply, convert back to sRGB
	Co := [3]float64{co[0] / ao, co[1] / ao, co[2] / ao}
	outS := [3]float64{
		geom.LinearToSrgb(Co[0]),
		geom.LinearToSrgb(Co[1]),
		geom.LinearToSrgb(Co[2]),
	}

	return Color{
		R: geom.U8(outS[0]),
		G: geom.U8(outS[1]),
		B: geom.U8(outS[2]),
		A: uint8(math.Round(ao * 255)),
	}
}

// Separable blending modes

// blendChannel applies a single-channel blend operation (Cb, Cs ∈ [0–1]).
//
// Each formula matches the CSS Compositing and Blending specification.
// See: https://www.w3.org/TR/compositing-1/
func blendChannel(Cb, Cs float64, mode BlendMode) float64 {
	switch mode {
	case BlendNormal:
		// Normal: B(Cb, Cs) = Cs
		return Cs

	case BlendMultiply:
		// Multiply: B(Cb, Cs) = Cb * Cs
		sb := geom.LinearToSrgb(Cb)
		ss := geom.LinearToSrgb(Cs)
		return geom.SrgbToLinear(sb * ss)

	case BlendOverlay:
		// Overlay: B(Cb, Cs) = 2*Cb*Cs if Cb < 0.5 else 1 - 2*(1-Cb)*(1-Cs)
		sb := geom.LinearToSrgb(Cb)
		ss := geom.LinearToSrgb(Cs)
		var resS float64
		if sb < 0.5 {
			resS = 2 * sb * ss
		} else {
			resS = 1 - 2*(1-sb)*(1-ss)
		}
		return geom.SrgbToLinear(resS)

	case BlendDarken:
		// Darken: B(Cb, Cs) = min(Cb, Cs)
		sb := geom.LinearToSrgb(Cb)
		ss := geom.LinearToSrgb(Cs)
		return geom.SrgbToLinear(math.Min(sb, ss))

	case BlendLighten:
		// Lighten: B(Cb, Cs) = max(Cb, Cs)
		sb := geom.LinearToSrgb(Cb)
		ss := geom.LinearToSrgb(Cs)
		return geom.SrgbToLinear(math.Max(sb, ss))

	case BlendColorDodge:
		// Color Dodge: B(Cb, Cs) = 1 if Cs >= 1 else min(1, Cb / (1 - Cs))
		sb := geom.LinearToSrgb(Cb)
		ss := geom.LinearToSrgb(Cs)
		if ss >= 1 {
			return 1
		}
		if sb <= 0 {
			return 0
		}
		return geom.SrgbToLinear(geom.ClampF64(sb/(1-ss), 0, 1))

	case BlendHardLight:
		// Hard Light: B(Cb, Cs) = 2*Cb*Cs if Cs < 0.5 else 1 - 2*(1-Cb)*(1-Cs)
		sb := geom.LinearToSrgb(Cb)
		ss := geom.LinearToSrgb(Cs)
		var resS float64
		if ss <= 0.5 {
			resS = 2 * sb * ss
		} else {
			resS = 1 - 2*(1-sb)*(1-ss)
		}
		return geom.SrgbToLinear(resS)

	case BlendSoftLight:
		// Soft Light (W3C formula):
		// if Cb ≤ 0.25: g = ((16*Cb - 12)*Cb + 4)*Cb
		// else: g = sqrt(Cb)
		// then B(Cb, Cs) = Cb + (2*Cs - 1)*(g - Cb)
		sb := geom.ClampF64(geom.LinearToSrgb(Cb), 0, 1)
		ss := geom.ClampF64(geom.LinearToSrgb(Cs), 0, 1)

		var g float64
		if sb <= 0.25 {
			g = ((16*sb-12)*sb + 4) * sb
		} else {
			g = math.Sqrt(sb)
		}

		var resS float64
		if ss <= 0.5 {
			resS = sb - (1-2*ss)*sb*(1-sb)
		} else {
			resS = sb + (2*ss-1)*(g-sb)
		}
		return geom.SrgbToLinear(geom.ClampF64(resS, 0, 1))

	case BlendDifference:
		// Difference: B(Cb, Cs) = |Cb - Cs|
		sb := geom.LinearToSrgb(Cb)
		ss := geom.LinearToSrgb(Cs)
		return geom.SrgbToLinear(math.Abs(sb - ss))

	case BlendExclusion:
		// Exclusion: B(Cb, Cs) = Cb + Cs - 2*Cb*Cs
		sb := geom.LinearToSrgb(Cb)
		ss := geom.LinearToSrgb(Cs)
		return geom.SrgbToLinear(sb + ss - 2*sb*ss)

	case BlendPlusLighter:
		// Linear Dodge (Plus Lighter): B(Cb, Cs) = min(1, Cb + Cs)
		sb := geom.LinearToSrgb(Cb)
		ss := geom.LinearToSrgb(Cs)
		return geom.SrgbToLinear(geom.ClampF64(sb+ss, 0, 1))

	case BlendPlusDarker:
		// Linear Burn (Plus Darker): B(Cb, Cs) = max(0, Cb + Cs - 1)
		sb := geom.LinearToSrgb(Cb)
		ss := geom.LinearToSrgb(Cs)
		return geom.SrgbToLinear(geom.ClampF64(sb+ss-1, 0, 1))

	default:
		return Cs
	}
}

// Non-separable modes (HSL-space)

// nonSeparableB implements CSS non-separable blend modes:
// Hue, Saturation, Color, and Luminosity.
// These operate in HSL space rather than per-channel RGB.
func nonSeparableB(Cb, Cs [3]float64, mode BlendMode) [3]float64 {
	CbS := [3]float64{geom.LinearToSrgb(Cb[0]), geom.LinearToSrgb(Cb[1]), geom.LinearToSrgb(Cb[2])}
	CsS := [3]float64{geom.LinearToSrgb(Cs[0]), geom.LinearToSrgb(Cs[1]), geom.LinearToSrgb(Cs[2])}

	var outS [3]float64
	switch mode {
	case BlendHue:
		// Hue: B(Cb, Cs) = setLum(setSat(Cs, sat(Cb)), lum(Cb))
		outS = geom.SetLum(geom.SetSat(CsS, geom.Sat(CbS)), geom.Lum(CbS))
	case BlendSaturation:
		// Saturation: B(Cb, Cs) = setLum(setSat(Cb, sat(Cs)), lum(Cb))
		outS = geom.SetLum(geom.SetSat(CbS, geom.Sat(CsS)), geom.Lum(CbS))
	case BlendColor:
		// Color: B(Cb, Cs) = setLum(Cs, lum(Cb))
		outS = geom.SetLum(CsS, geom.Lum(CbS))
	case BlendLuminosity:
		// Luminosity: B(Cb, Cs) = setLum(Cb, lum(Cs))
		outS = geom.SetLum(CbS, geom.Lum(CsS))
	default:
		outS = CsS
	}

	return [3]float64{geom.SrgbToLinear(outS[0]), geom.SrgbToLinear(outS[1]), geom.SrgbToLinear(outS[2])}
}
