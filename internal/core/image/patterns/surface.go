package patterns

import (
	"image"
	"image/color"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// RepeatOp defines how an image-based pattern repeats along the X and Y axes.
type RepeatOp int

const (
	// RepeatBoth repeats the image in both X and Y directions.
	RepeatBoth RepeatOp = iota
	// RepeatX repeats only along the X axis.
	RepeatX
	// RepeatY repeats only along the Y axis.
	RepeatY
	// RepeatNone disables repetition; pixels outside the image bounds are transparent.
	RepeatNone
)

// Surface represents an image-based pattern (a texture) that can be repeated
// or clamped along one or both axes. It implements the Pattern and BlendedPattern
// interfaces and supports blending and opacity adjustments.
type Surface struct {
	im image.Image // Source image (texture)
	op RepeatOp    // Repetition mode

	mode    BlendMode // Blending mode
	opacity float64   // Opacity factor [0, 1]
}

// BlendMode returns the blending mode associated with this surface pattern.
func (s *Surface) BlendMode() BlendMode { return s.mode }

// Opacity returns the pattern’s opacity factor in [0, 1].
func (s *Surface) Opacity() float64 { return s.opacity }

// Sampling

// ColorAt returns the color of the surface at pixel coordinate (x, y).
// Depending on the RepeatOp setting, the image can repeat along one or both axes.
//
// Behavior:
//   - RepeatBoth — tiles in both X and Y directions.
//   - RepeatX — repeats horizontally; outside vertical bounds is transparent.
//   - RepeatY — repeats vertically; outside horizontal bounds is transparent.
//   - RepeatNone — outside both bounds is transparent.
//
// Example use cases:
//   - Seamless textures for fills or brush patterns.
//   - Image-based masks with blending applied per pixel.
func (s *Surface) ColorAt(x, y int) color.Color {
	b := s.im.Bounds()

	switch s.op {
	case RepeatX:
		if y >= b.Dy() {
			return color.Transparent
		}
	case RepeatY:
		if x >= b.Dx() {
			return color.Transparent
		}
	case RepeatNone:
		if x >= b.Dx() || y >= b.Dy() {
			return color.Transparent
		}
	}

	// Wrap coordinates within image bounds if repeating
	x = x%b.Dx() + b.Min.X
	y = y%b.Dy() + b.Min.Y

	// Convert to internal Color and apply blend mode
	return NewColorFromStd(s.im.At(x, y)).SetBlendMode(s.mode)
}

// Constructors

// NewSurface creates a new Surface pattern from an image with a repetition mode.
// By default, it uses normal blending and full opacity.
func NewSurface(im image.Image, op RepeatOp) *Surface {
	return &Surface{im: im, op: op}
}

// NewSurfaceWithBlend creates a new Surface pattern with custom blending and opacity.
// The opacity value is clamped to [0, 1].
func NewSurfaceWithBlend(im image.Image, op RepeatOp, mode BlendMode, opacity float64) *Surface {
	return &Surface{
		im:      im,
		op:      op,
		mode:    mode,
		opacity: geom.ClampF64(opacity, 0, 1),
	}
}

// Modifiers

// WithBlendMode sets the blend mode and returns the same Surface instance for chaining.
func (s *Surface) WithBlendMode(m BlendMode) *Surface {
	s.mode = m
	return s
}

// WithOpacity sets the opacity (clamped to [0, 1]) and returns the same Surface instance for chaining.
func (s *Surface) WithOpacity(a float64) *Surface {
	s.opacity = geom.ClampF64(a, 0, 1)
	return s
}
