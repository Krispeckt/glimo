package patterns

import (
	"image/color"
	"sort"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// LinearGradient represents a linear color gradient between two points (x0, y0) and (x1, y1).
// Colors are interpolated along the line connecting these points according to color stops.
type LinearGradient struct {
	x0, y0, x1, y1 float64    // Start and end points of the gradient
	stops          geom.Stops // Sorted list of color stops

	mode    BlendMode // Blending mode applied during rendering
	opacity float64   // Opacity value in [0, 1]
}

// BlendMode returns the blending mode associated with the gradient.
func (g *LinearGradient) BlendMode() BlendMode { return g.mode }

// Opacity returns the opacity level of the gradient (0–1).
func (g *LinearGradient) Opacity() float64 { return g.opacity }

// Constructors

// NewLinearGradient creates a new linear gradient from (x0, y0) to (x1, y1)
// with default blend mode (Normal) and full opacity.
func NewLinearGradient(x0, y0, x1, y1 float64) GradientPattern {
	return &LinearGradient{
		x0: x0, y0: y0, x1: x1, y1: y1,
		mode: BlendPassThrough, opacity: 1,
	}
}

// NewLinearGradientWithBlend creates a linear gradient from (x0, y0) to (x1, y1)
// with a specified blend mode and opacity.
func NewLinearGradientWithBlend(x0, y0, x1, y1 float64, mode BlendMode, opacity float64) GradientPattern {
	return &LinearGradient{
		x0: x0, y0: y0, x1: x1, y1: y1,
		mode: mode, opacity: geom.ClampF64(opacity, 0, 1),
	}
}

// WithBlendMode sets the gradient’s blending mode and returns the gradient itself for chaining.
func (g *LinearGradient) WithBlendMode(m BlendMode) *LinearGradient {
	g.mode = m
	return g
}

// WithOpacity sets the opacity level of the gradient and returns it for chaining.
func (g *LinearGradient) WithOpacity(a float64) *LinearGradient {
	g.opacity = geom.ClampF64(a, 0, 1)
	return g
}

// Color Stops

// AddColorStop adds a new color stop at the specified offset [0–1].
// Stops define how colors transition along the gradient axis.
func (g *LinearGradient) AddColorStop(offset float64, c Color) GradientPattern {
	offset = geom.ClampF64(offset, 0, 1)
	g.stops = append(g.stops, geom.NewStop(offset, c))
	sort.Sort(g.stops)
	return g
}

// Sampling

// ColorAt returns the interpolated color at a specific pixel coordinate (x, y).
// The function projects the point onto the gradient vector and computes
// the normalized interpolation factor t ∈ [0, 1]:
//
//	t = ((x - x0)*dx + (y - y0)*dy) / (dx² + dy²)
//
// Depending on the gradient’s orientation, the function automatically
// handles horizontal, vertical, and diagonal gradients.
func (g *LinearGradient) ColorAt(x, y int) color.Color {
	if len(g.stops) == 0 {
		return color.Transparent
	}

	fx, fy := float64(x), float64(y)
	dx, dy := g.x1-g.x0, g.y1-g.y0

	switch {
	case dy == 0 && dx != 0:
		// Horizontal gradient
		return geom.GetColor(geom.ClampF64((fx-g.x0)/dx, 0, 1), g.stops)
	case dx == 0 && dy != 0:
		// Vertical gradient
		return geom.GetColor(geom.ClampF64((fy-g.y0)/dy, 0, 1), g.stops)
	default:
		// General linear gradient
		den := dx*dx + dy*dy
		if den == 0 {
			return g.stops[0].Color()
		}
		t := ((fx-g.x0)*dx + (fy-g.y0)*dy) / den
		return geom.GetColor(geom.ClampF64(t, 0, 1), g.stops)
	}
}
