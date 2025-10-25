package colors

import (
	"image"
	"strings"

	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

//
// Package Overview
//
// The `colors` package provides a high-level interface for constructing and composing
// color and pattern objects from the internal `patterns` package.
//
// It exposes the main gradient, surface, and blending abstractions directly under
// the `colors` namespace, allowing concise and readable use in client code.
//

// Transparent represents a fully transparent color (0, 0, 0, 0).
var Transparent = RGBA(0, 0, 0, 0)

//
// Type Aliases
//
// These aliases re-export key types from `patterns` for simplified access.
//

type (
	// Pattern defines any drawable color source capable of returning a color at (x, y).
	Pattern = patterns.Pattern
	// BlendedPattern extends Pattern with blending mode and opacity support.
	BlendedPattern = patterns.BlendedPattern
	// GradientPattern defines the interface common to all gradient types.
	GradientPattern = patterns.GradientPattern

	// ConicGradient represents a gradient that varies angularly around a center point.
	ConicGradient = patterns.ConicGradient
	// LinearGradient represents a gradient that transitions linearly between two points.
	LinearGradient = patterns.LinearGradient
	// RadialGradient represents a gradient that transitions between two circular regions.
	RadialGradient = patterns.RadialGradient
	// Solid represents a constant color fill.
	Solid = patterns.Solid
	// Surface represents an image-based pattern that can be repeated or clamped.
	Surface = patterns.Surface
)

//
// Gradient Constructors
//
// These functions provide direct access to the gradient creation logic
// from the `patterns` package.
//

// NewConicGradient creates a new conic gradient centered at (cx, cy),
// starting from the given angular offset in degrees.
func NewConicGradient(cx, cy, deg float64) *patterns.ConicGradient {
	return patterns.NewConicGradient(cx, cy, deg)
}

// NewConicGradientWithBlend creates a conic gradient with a blend mode and opacity.
func NewConicGradientWithBlend(cx, cy, deg float64, blend patterns.BlendMode, opacity float64) *patterns.ConicGradient {
	return patterns.NewConicGradientWithBlend(cx, cy, deg, blend, opacity)
}

// NewLinearGradient creates a new linear gradient between (x0, y0) and (x1, y1).
func NewLinearGradient(x0, y0, x1, y1 float64) *patterns.LinearGradient {
	return patterns.NewLinearGradient(x0, y0, x1, y1)
}

// NewLinearGradientWithBlend creates a linear gradient with a specific blend mode and opacity.
func NewLinearGradientWithBlend(x0, y0, x1, y1 float64, blend patterns.BlendMode, opacity float64) *patterns.LinearGradient {
	return patterns.NewLinearGradientWithBlend(x0, y0, x1, y1, blend, opacity)
}

// NewRadialGradient creates a new radial gradient between two circular regions.
func NewRadialGradient(cx0, cy0, r0, cx1, cy1, r1 float64) *patterns.RadialGradient {
	return patterns.NewRadialGradient(cx0, cy0, r0, cx1, cy1, r1)
}

// NewRadialGradientWithBlend creates a radial gradient with blending and opacity control.
func NewRadialGradientWithBlend(cx0, cy0, r0, cx1, cy1, r1 float64, blend patterns.BlendMode, opacity float64) *patterns.RadialGradient {
	return patterns.NewRadialGradientWithBlend(cx0, cy0, r0, cx1, cy1, r1, blend, opacity)
}

// NewSolid creates a solid fill using the given color.
func NewSolid(c patterns.Color) *patterns.Solid {
	return patterns.NewSolid(c)
}

// NewSolidWithBlend creates a solid fill with a specified blend mode and opacity.
func NewSolidWithBlend(c patterns.Color, blend patterns.BlendMode, opacity float64) *patterns.Solid {
	return patterns.NewSolidWithBlend(c, blend, opacity)
}

// NewSurface creates an image-based pattern with a specified repetition mode.
func NewSurface(img image.Image, repeat patterns.RepeatOp) *patterns.Surface {
	return patterns.NewSurface(img, repeat)
}

// NewSurfaceWithBlend creates an image-based pattern with repetition, blending, and opacity options.
func NewSurfaceWithBlend(img image.Image, repeat patterns.RepeatOp, blend patterns.BlendMode, opacity float64) *patterns.Surface {
	return patterns.NewSurfaceWithBlend(img, repeat, blend, opacity)
}

//
// Color Constructors
//
// These helper functions simplify the creation of Color objects
// compatible with the rest of the rendering system.
//

// RGBA constructs a Color from red, green, blue, and alpha components (0–255 each).
func RGBA(r, g, b, a uint8) patterns.Color {
	return patterns.Color{R: r, G: g, B: b, A: a}
}

// RGB constructs a fully opaque Color from red, green, and blue components (0–255 each).
func RGB(r, g, b uint8) patterns.Color {
	return patterns.Color{R: r, G: g, B: b, A: 255}
}

// HEX parses a hexadecimal color string (e.g. "#RRGGBB" or "#RRGGBBAA") into a Color object.
func HEX(hex string) (patterns.Color, error) {
	return patterns.ColorFromHex(hex)
}

// HSL constructs a Color from hue, saturation, and lightness values.
// The resulting color is fully opaque.
func HSL(h, s, l float64) patterns.Color {
	return patterns.ColorFromHSL(h, s, l, 255)
}

//
// Surface Repetition Modes
//
// These constants define how image-based patterns repeat along the X and Y axes.
//

type SurfaceRepeatOp = patterns.RepeatOp

const (
	// SurfaceRepeatBoth repeats the image in both directions.
	SurfaceRepeatBoth SurfaceRepeatOp = patterns.RepeatBoth
	// SurfaceRepeatX repeats the image only along the X axis.
	SurfaceRepeatX SurfaceRepeatOp = patterns.RepeatX
	// SurfaceRepeatY repeats the image only along the Y axis.
	SurfaceRepeatY SurfaceRepeatOp = patterns.RepeatY
	// SurfaceRepeatNone disables repetition and makes out-of-bounds pixels transparent.
	SurfaceRepeatNone SurfaceRepeatOp = patterns.RepeatNone
)

//
// Blend Modes
//
// Blend modes define how colors and patterns combine visually when composited.
// They follow conventions from digital compositing and CSS blending specifications.
//

type BlendMode = patterns.BlendMode

const (
	BlendPassThrough BlendMode = patterns.BlendPassThrough // disables blending; uses source color directly
	BlendNormal      BlendMode = patterns.BlendNormal      // standard source-over-destination compositing
	BlendDarken      BlendMode = patterns.BlendDarken      // selects darker pixels between layers
	BlendMultiply    BlendMode = patterns.BlendMultiply    // multiplies source and destination colors
	BlendPlusDarker  BlendMode = patterns.BlendPlusDarker  // linear burn effect
	BlendColorBurn   BlendMode = patterns.BlendColorBurn   // darkens by increasing contrast
	BlendLighten     BlendMode = patterns.BlendLighten     // selects lighter pixels between layers
	BlendScreen      BlendMode = patterns.BlendScreen      // inverse of multiply (lightens colors)
	BlendPlusLighter BlendMode = patterns.BlendPlusLighter // linear dodge effect
	BlendColorDodge  BlendMode = patterns.BlendColorDodge  // brightens by decreasing contrast
	BlendOverlay     BlendMode = patterns.BlendOverlay     // combines multiply and screen based on lightness
	BlendSoftLight   BlendMode = patterns.BlendSoftLight   // soft contrast modulation
	BlendHardLight   BlendMode = patterns.BlendHardLight   // strong contrast modulation
	BlendDifference  BlendMode = patterns.BlendDifference  // subtracts darker values
	BlendExclusion   BlendMode = patterns.BlendExclusion   // softens the difference effect
	BlendHue         BlendMode = patterns.BlendHue         // applies source hue; preserves destination luminance/saturation
	BlendSaturation  BlendMode = patterns.BlendSaturation  // applies source saturation; preserves destination hue/luminance
	BlendColor       BlendMode = patterns.BlendColor       // applies source hue/saturation; preserves destination luminance
	BlendLuminosity  BlendMode = patterns.BlendLuminosity  // applies source luminance; preserves destination hue/saturation
	BlendLinearBurn  BlendMode = patterns.BlendLinearBurn  // alias for PlusDarker
	BlendLinearDodge BlendMode = patterns.BlendLinearDodge // alias for PlusLighter
)

var BlendModeMap = map[string]BlendMode{
	"pass-through": BlendPassThrough,
	"normal":       BlendNormal,
	"darken":       BlendDarken,
	"multiply":     BlendMultiply,
	"plus-darker":  BlendPlusDarker,
	"color-burn":   BlendColorBurn,
	"lighten":      BlendLighten,
	"screen":       BlendScreen,
	"plus-lighter": BlendPlusLighter,
	"color-dodge":  BlendColorDodge,
	"overlay":      BlendOverlay,
	"soft-light":   BlendSoftLight,
	"hard-light":   BlendHardLight,
	"difference":   BlendDifference,
	"exclusion":    BlendExclusion,
	"hue":          BlendHue,
	"saturation":   BlendSaturation,
	"color":        BlendColor,
	"luminosity":   BlendLuminosity,
	"linear-burn":  BlendLinearBurn,
	"linear-dodge": BlendLinearDodge,
}

// ParseBlendMode converts a string into the corresponding BlendMode constant.
// Returns BlendNormal if not found.
func ParseBlendMode(s string) BlendMode {
	if mode, ok := BlendModeMap[strings.ToLower(s)]; ok {
		return mode
	}
	return BlendNormal
}
