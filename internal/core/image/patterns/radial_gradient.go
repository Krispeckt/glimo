package patterns

import (
	"image/color"
	"math"
	"sort"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// RadialGradient represents a two-point radial gradient defined by
// two circles (c0, c1) and their relative displacement (cd).
// The color transition occurs along the radii from the first circle to the second.
type RadialGradient struct {
	c0, c1, cd *geom.Circle // Start, end, and delta circles
	a, inva    float64      // Precomputed coefficients for intersection solving
	mindr      float64      // Minimum allowed distance between circles
	stops      geom.Stops   // Sorted list of color stops

	mode    BlendMode // Blend mode applied during rendering
	opacity float64   // Opacity in range [0, 1]
}

// BlendMode returns the blend mode associated with the gradient.
func (g *RadialGradient) BlendMode() BlendMode { return g.mode }

// Opacity returns the gradient’s opacity factor in [0, 1].
func (g *RadialGradient) Opacity() float64 { return g.opacity }

// Constructors

// NewRadialGradient creates a new radial gradient defined by two circles:
// the start circle (x0, y0, r0) and the end circle (x1, y1, r1).
// Colors transition radially from the inner circle to the outer one.
func NewRadialGradient(x0, y0, r0, x1, y1, r1 float64) *RadialGradient {
	c0 := geom.NewCircle(x0, y0, r0)
	c1 := geom.NewCircle(x1, y1, r1)
	cd := geom.NewCircle(x1-x0, y1-y0, r1-r0)
	a := geom.Dot3(cd.X(), cd.Y(), -cd.Radius(), cd.X(), cd.Y(), cd.Radius())

	var inva float64
	if a != 0 {
		inva = 1.0 / a
	}

	return &RadialGradient{
		c0:      c0,
		c1:      c1,
		cd:      cd,
		a:       a,
		inva:    inva,
		mindr:   -c0.Radius(),
		mode:    BlendPassThrough,
		opacity: 1,
	}
}

// NewRadialGradientWithBlend creates a new radial gradient with a specified
// blend mode and opacity, using the same geometry definition as above.
func NewRadialGradientWithBlend(x0, y0, r0, x1, y1, r1 float64, mode BlendMode, opacity float64) *RadialGradient {
	c0 := geom.NewCircle(x0, y0, r0)
	c1 := geom.NewCircle(x1, y1, r1)
	cd := geom.NewCircle(x1-x0, y1-y0, r1-r0)
	a := geom.Dot3(cd.X(), cd.Y(), -cd.Radius(), cd.X(), cd.Y(), cd.Radius())

	var inva float64
	if a != 0 {
		inva = 1.0 / a
	}

	return &RadialGradient{
		c0:      c0,
		c1:      c1,
		cd:      cd,
		a:       a,
		inva:    inva,
		mindr:   -c0.Radius(),
		mode:    mode,
		opacity: geom.ClampF64(opacity, 0, 1),
	}
}

// WithBlendMode sets the blending mode and returns the gradient for chaining.
func (g *RadialGradient) WithBlendMode(m BlendMode) *RadialGradient {
	g.mode = m
	return g
}

// WithOpacity sets the gradient opacity and returns the gradient for chaining.
func (g *RadialGradient) WithOpacity(a float64) *RadialGradient {
	g.opacity = geom.ClampF64(a, 0, 1)
	return g
}

// Color Stops

// AddColorStop adds a color stop to the gradient at a specified offset [0, 1].
// Stops define how colors transition along the gradient.
func (g *RadialGradient) AddColorStop(offset float64, c Color) GradientPattern {
	offset = geom.ClampF64(offset, 0, 1)
	g.stops = append(g.stops, geom.NewStop(offset, c))
	sort.Sort(g.stops)
	return g
}

// Sampling

// ColorAt computes the interpolated color for a given pixel (x, y).
// The algorithm calculates the position of the pixel relative to the two circles
// and solves for the intersection point of the pixel ray with the gradient’s radius transition.
//
// This follows the SVG radial gradient equation:
//
//	(x - x0)² + (y - y0)² = (r0 + t*(r1 - r0))²
//
// where `t` ∈ [0,1] represents the normalized interpolation factor between both circles.
//
// Steps:
//  1. Compute vector from first circle to pixel (dx, dy).
//  2. Solve quadratic equation a*t² + 2b*t + c = 0 for t.
//  3. Pick the smallest valid t within [0, 1] and sample the color stops.
//  4. Return transparent if the pixel lies outside the gradient bounds.
func (g *RadialGradient) ColorAt(x, y int) color.Color {
	if len(g.stops) == 0 {
		return color.Transparent
	}

	dx, dy := float64(x)+0.5-g.c0.X(), float64(y)+0.5-g.c0.Y()
	b := geom.Dot3(dx, dy, g.c0.Radius(), g.cd.X(), g.cd.Y(), g.cd.Radius())
	c := geom.Dot3(dx, dy, -g.c0.Radius(), dx, dy, g.c0.Radius())

	if g.a == 0 {
		// Degenerate case: linear relationship between circles.
		if b == 0 {
			return color.Transparent
		}
		t := 0.5 * c / b
		if t*g.cd.Radius() >= g.mindr {
			return geom.GetColor(t, g.stops)
		}
		return color.Transparent
	}

	// Quadratic discriminant: b² - a*c
	discr := geom.Dot3(b, g.a, 0, b, -c, 0)
	if discr < 0 {
		return color.Transparent
	}

	sqrtd := math.Sqrt(discr)
	t0 := (b + sqrtd) * g.inva
	t1 := (b - sqrtd) * g.inva

	switch {
	case t0*g.cd.Radius() >= g.mindr:
		return geom.GetColor(t0, g.stops)
	case t1*g.cd.Radius() >= g.mindr:
		return geom.GetColor(t1, g.stops)
	default:
		return color.Transparent
	}
}
