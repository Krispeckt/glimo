package patterns

import (
	"image/color"
	"math"
	"sort"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// ConicGradient represents a conic (angular) gradient.
// The gradient is centered at (cx, cy), and colors transition
// around the center based on the angle of each pixel relative to it.
type ConicGradient struct {
	cx, cy   float64    // Center coordinates
	rotation float64    // Rotation offset in turns (0–1)
	stops    geom.Stops // Sorted list of color stops

	mode    BlendMode // Blending mode applied during rendering
	opacity float64   // Opacity factor in [0, 1]
}

// BlendMode returns the blend mode of the gradient.
func (g *ConicGradient) BlendMode() BlendMode { return g.mode }

// Opacity returns the opacity factor of the gradient (0–1).
func (g *ConicGradient) Opacity() float64 { return g.opacity }

// Constructors

// NewConicGradient creates a new conic gradient centered at (cx, cy),
// rotating `deg` degrees clockwise.
// Rotation is normalized to [0, 1) turns, where 1.0 = 360°.
func NewConicGradient(cx, cy, deg float64) GradientPattern {
	return &ConicGradient{
		cx:       cx,
		cy:       cy,
		rotation: geom.NormalizeAngle(deg) / 360,
		mode:     BlendPassThrough,
		opacity:  1,
	}
}

// NewConicGradientWithBlend creates a conic gradient with a specified
// blending mode and opacity.
func NewConicGradientWithBlend(cx, cy, deg float64, mode BlendMode, opacity float64) GradientPattern {
	return &ConicGradient{
		cx:       cx,
		cy:       cy,
		rotation: geom.NormalizeAngle(deg) / 360,
		mode:     mode,
		opacity:  geom.ClampF64(opacity, 0, 1),
	}
}

// WithBlendMode sets the blending mode and returns the gradient.
func (g *ConicGradient) WithBlendMode(m BlendMode) *ConicGradient {
	g.mode = m
	return g
}

// WithOpacity sets the opacity (0–1) and returns the gradient.
func (g *ConicGradient) WithOpacity(a float64) *ConicGradient {
	g.opacity = geom.ClampF64(a, 0, 1)
	return g
}

// Color Stops

// AddColorStop adds a color stop to the gradient at the given offset [0–1].
// Stops are automatically sorted by position.
func (g *ConicGradient) AddColorStop(offset float64, c Color) GradientPattern {
	offset = geom.ClampF64(offset, 0, 1)
	g.stops = append(g.stops, geom.NewStop(offset, c))
	sort.Sort(g.stops)
	return g
}

// StopsCount returns the total number of color stops.
func (g *ConicGradient) StopsCount() int { return len(g.stops) }

// Sampling

// ColorAt computes the interpolated color for the given pixel (x, y).
// The function calculates the angular position of the point relative to
// the center, maps it to a normalized offset [0–1), and interpolates between
// the nearest color stops.
func (g *ConicGradient) ColorAt(x, y int) color.Color {
	if len(g.stops) == 0 {
		return color.Transparent
	}
	angle := g.angleAt(float64(x), float64(y))
	t := g.angleToOffset(angle)
	return geom.GetColor(t, g.stops)
}

// Geometry

// Center returns the coordinates of the gradient’s center.
func (g *ConicGradient) Center() (float64, float64) { return g.cx, g.cy }

// Rotation returns the gradient rotation in degrees (0–360).
func (g *ConicGradient) Rotation() float64 { return g.rotation * 360 }

// SetCenter updates the gradient’s center coordinates.
func (g *ConicGradient) SetCenter(cx, cy float64) { g.cx, g.cy = cx, cy }

// SetRotation sets the rotation angle in degrees and normalizes it to [0, 360).
func (g *ConicGradient) SetRotation(deg float64) {
	g.rotation = geom.NormalizeAngle(deg) / 360
}

// Internal math

// angleAt computes the polar angle (in radians) of a point relative to the gradient’s center.
// Range: [-π, π].
func (g *ConicGradient) angleAt(x, y float64) float64 {
	return math.Atan2(y-g.cy, x-g.cx)
}

// angleToOffset converts an angle (radians) to a normalized offset [0, 1)
// within the gradient’s angular range, applying rotation offset.
func (g *ConicGradient) angleToOffset(a float64) float64 {
	t := geom.Norm(a, -math.Pi, math.Pi) - g.rotation
	if t < 0 {
		t += 1
	}
	return t
}
