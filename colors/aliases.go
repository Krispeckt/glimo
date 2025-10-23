package colors

import (
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

//
// Package Overview
//
// The `colors` package provides a simplified, user-friendly interface for creating and
// manipulating color and pattern objects from the internal `patterns` package.
//
// It re-exports gradient constructors, solid fills, surfaces, and blending modes,
// allowing developers to use them directly under `colors.*` without referencing subpackages.
//

// Transparent represents a fully transparent color (0, 0, 0, 0).
var Transparent = RGBA(0, 0, 0, 0)

//
// Type Aliases
//
// These aliases expose key types from the `patterns` package directly under the `colors` namespace.
//

type (
	// Pattern defines a drawable color source with a ColorAt(x, y) method.
	Pattern = patterns.Pattern
	// BlendedPattern extends Pattern with blend mode and opacity support.
	BlendedPattern = patterns.BlendedPattern
	// GradientPattern defines a common interface for gradient types (linear, radial, conic).
	GradientPattern = patterns.GradientPattern

	// ConicGradient represents an angular gradient centered at (cx, cy).
	ConicGradient = patterns.ConicGradient
	// LinearGradient represents a linear gradient between two points.
	LinearGradient = patterns.LinearGradient
	// RadialGradient represents a radial gradient between two circles.
	RadialGradient = patterns.RadialGradient
	// Solid represents a solid color fill.
	Solid = patterns.Solid
	// Surface represents an image-based pattern (texture) with repetition options.
	Surface = patterns.Surface
)

//
// Gradient Constructors
//
// These are direct aliases to the gradient creation functions from `patterns`.
//

var (
	NewConicGradient          = patterns.NewConicGradient
	NewConicGradientWithBlend = patterns.NewConicGradientWithBlend

	NewLinearGradient          = patterns.NewLinearGradient
	NewLinearGradientWithBlend = patterns.NewLinearGradientWithBlend

	NewRadialGradient          = patterns.NewRadialGradient
	NewRadialGradientWithBlend = patterns.NewRadialGradientWithBlend

	NewSolid          = patterns.NewSolid
	NewSolidWithBlend = patterns.NewSolidWithBlend

	NewSurface          = patterns.NewSurface
	NewSurfaceWithBlend = patterns.NewSurfaceWithBlend
)

//
// Color Constructors
//
// Provides convenient color creation utilities compatible with the rest of the API.
//

// RGBA creates a new color from red, green, blue, and alpha components (0–255 each).
func RGBA(r, g, b, a uint8) patterns.Color {
	return patterns.Color{R: r, G: g, B: b, A: a}
}

// RGB creates a new fully opaque color from red, green, and blue components (0–255 each).
func RGB(r, g, b uint8) patterns.Color {
	return patterns.Color{R: r, G: g, B: b, A: 255}
}

// HEX parses a hexadecimal color string (e.g. "#RRGGBB" or "#RRGGBBAA") into a Color.
var HEX = patterns.ColorFromHex

// HSL constructs a Color from hue, saturation, and lightness values, plus an alpha channel.
var HSL = patterns.ColorFromHSL

//
// Surface Repetition Modes
//
// Defines how image-based patterns (Surface) repeat or clamp along the X and Y axes.
//

type SurfaceRepeatOp = patterns.RepeatOp

var (
	// SurfaceRepeatBoth repeats the image in both directions.
	SurfaceRepeatBoth = patterns.RepeatBoth
	// SurfaceRepeatX repeats only along the X axis.
	SurfaceRepeatX = patterns.RepeatX
	// SurfaceRepeatY repeats only along the Y axis.
	SurfaceRepeatY = patterns.RepeatY
	// SurfaceRepeatNone disables repetition; pixels outside the image bounds are transparent.
	SurfaceRepeatNone = patterns.RepeatNone
)

//
// Blend Modes
//
// Defines how colors and patterns interact visually when composited.
// These correspond to standard CSS and digital compositing blend modes.
//

type BlendMode = patterns.BlendMode

var (
	BlendPassThrough = patterns.BlendPassThrough // disables blending, passes color as-is
	BlendNormal      = patterns.BlendNormal      // normal (source over destination)
	BlendDarken      = patterns.BlendDarken      // selects darker pixels
	BlendMultiply    = patterns.BlendMultiply    // multiplies source and destination
	BlendPlusDarker  = patterns.BlendPlusDarker  // linear burn
	BlendColorBurn   = patterns.BlendColorBurn   // darkens by increasing contrast
	BlendLighten     = patterns.BlendLighten     // selects lighter pixels
	BlendScreen      = patterns.BlendScreen      // inverse of multiply
	BlendPlusLighter = patterns.BlendPlusLighter // linear dodge
	BlendColorDodge  = patterns.BlendColorDodge  // brightens by decreasing contrast
	BlendOverlay     = patterns.BlendOverlay     // combines multiply/screen based on lightness
	BlendSoftLight   = patterns.BlendSoftLight   // soft contrast-based blend
	BlendHardLight   = patterns.BlendHardLight   // strong contrast-based blend
	BlendDifference  = patterns.BlendDifference  // subtracts darker values
	BlendExclusion   = patterns.BlendExclusion   // softer version of difference
	BlendHue         = patterns.BlendHue         // applies source hue, keeps destination luminance/saturation
	BlendSaturation  = patterns.BlendSaturation  // applies source saturation, keeps destination hue/luminance
	BlendColor       = patterns.BlendColor       // applies source hue/saturation, keeps destination luminance
	BlendLuminosity  = patterns.BlendLuminosity  // applies source luminance, keeps destination hue/saturation
	BlendLinearBurn  = patterns.BlendLinearBurn  // alias of PlusDarker
	BlendLinearDodge = patterns.BlendLinearDodge // alias of PlusLighter
)
