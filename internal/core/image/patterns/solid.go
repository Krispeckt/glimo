package patterns

import (
	"image/color"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// Solid represents a single uniform color pattern (no gradients or textures).
// It implements the Pattern and BlendedPattern interfaces, allowing it to be
// composited with blend modes and opacity like other pattern types.
type Solid struct {
	color   Color     // Base color of the pattern
	mode    BlendMode // Blend mode applied during rendering
	opacity float64   // Opacity factor in [0, 1]
}

// ColorAt returns the same solid color for any pixel coordinates (x, y).
// This makes Solid suitable for use as a background or overlay fill.
func (p *Solid) ColorAt(_, _ int) color.Color { return p.color }

// BlendMode returns the blend mode associated with this solid pattern.
func (p *Solid) BlendMode() BlendMode { return p.mode }

// Opacity returns the pattern’s opacity factor in the range [0, 1].
func (p *Solid) Opacity() float64 { return p.opacity }

// Constructors

// NewSolid creates a new solid pattern with the specified color,
// using the default blend mode (Normal) and full opacity.
func NewSolid(c Color) Pattern {
	return &Solid{color: c, mode: BlendPassThrough, opacity: 1}
}

// NewSolidWithBlend creates a new solid pattern with a specific blend mode and opacity.
// The opacity is clamped to the range [0, 1].
func NewSolidWithBlend(c Color, mode BlendMode, opacity float64) BlendedPattern {
	return &Solid{
		color:   c,
		mode:    mode,
		opacity: geom.ClampF64(opacity, 0, 1),
	}
}

// WithBlendMode sets the blend mode for the solid color and updates its internal color’s mode.
// Returns the same Solid instance for method chaining.
func (p *Solid) WithBlendMode(m BlendMode) *Solid {
	p.mode = m
	p.color.SetBlendMode(m)
	return p
}

// WithOpacity sets the opacity factor for the solid pattern.
// The value is clamped to [0, 1], and the instance is returned for chaining.
func (p *Solid) WithOpacity(a float64) *Solid {
	p.opacity = geom.ClampF64(a, 0, 1)
	return p
}
