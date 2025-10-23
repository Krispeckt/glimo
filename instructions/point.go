package instructions

import (
	"image"
	"math"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
	"golang.org/x/image/math/fixed"
)

// Vector2 is an alias for Point, used for semantic clarity when
// representing mathematical vectors instead of geometric coordinates.
type Vector2 = Point

// NewVec2 returns a new Vector2 (alias of Point) initialized with
// the given x and y values. The default color is black.
func NewVec2(x, y float64) *Vector2 {
	return &Vector2{X: x, Y: y, color: colors.Black}
}

// Point represents a 2D coordinate or vector using float64 precision.
// It also holds an optional color used for rendering operations.
type Point struct {
	X, Y  float64
	color patterns.Color
}

// NewPoint creates a new Point at (x, y) with the default color black.
func NewPoint(x, y float64) *Point {
	return &Point{X: x, Y: y, color: colors.Black}
}

// SetColor assigns a drawing color to the Point and returns itself
// for method chaining.
func (p *Point) SetColor(c patterns.Color) *Point {
	p.color = c
	return p
}

// Color returns the current color assigned to the Point.
func (p *Point) Color() patterns.Color {
	return p.color
}

// Fixed converts the Point to a fixed-point coordinate (26.6 format).
// Commonly used for subpixel-accurate rendering and rasterization.
func (p *Point) Fixed() fixed.Point26_6 {
	return fixed.Point26_6{
		X: fixed.Int26_6(p.X*64 + 0.5),
		Y: fixed.Int26_6(p.Y*64 + 0.5),
	}
}

// Draw renders a single pixel representing the Point’s location
// on the given overlay image, if it lies within bounds.
// The base image parameter is ignored for consistency with other Draw methods.
func (p *Point) Draw(_, overlay *image.RGBA) {
	x, y := int(math.Round(p.X)), int(math.Round(p.Y))
	if !image.Pt(x, y).In(overlay.Bounds()) {
		return
	}
	overlay.Set(x, y, p.color)
}

// Distance returns the Euclidean distance between the current Point (p)
// and another Point (q).
func (p *Point) Distance(q *Point) float64 {
	dx, dy := p.X-q.X, p.Y-q.Y
	return math.Hypot(dx, dy)
}

// Length returns the magnitude of the vector represented by the Point
// (distance from the origin to the point).
func (p *Point) Length() float64 {
	return math.Hypot(p.X, p.Y)
}

// Normalize returns a unit vector (length = 1) pointing in the same
// direction as p. If p is zero-length, it returns (0, 0).
func (p *Point) Normalize() Point {
	length := p.Length()
	if length == 0 {
		return Point{0, 0, p.color}
	}
	return Point{p.X / length, p.Y / length, p.color}
}

// Add returns the vector sum of p and q (p + q).
func (p *Point) Add(q Point) Point {
	return Point{p.X + q.X, p.Y + q.Y, p.color}
}

// Sub returns the vector difference between p and q (p − q).
func (p *Point) Sub(q Point) Point {
	return Point{p.X - q.X, p.Y - q.Y, p.color}
}

// Scale returns a new Point scaled by the specified factor.
// Both X and Y coordinates are multiplied by factor.
func (p *Point) Scale(factor float64) Point {
	return Point{p.X * factor, p.Y * factor, p.color}
}

// Midpoint returns the point exactly halfway between p and q.
func (p *Point) Midpoint(q Point) Point {
	return Point{
		X: (p.X + q.X) * 0.5,
		Y: (p.Y + q.Y) * 0.5,
	}
}

// Dot returns the dot product of p and q, commonly used for
// angle and projection calculations.
func (p *Point) Dot(q Point) float64 {
	return p.X*q.X + p.Y*q.Y
}

// Angle returns the angle in radians between vectors p and q,
// computed using the dot product. Returns 0 if either vector
// has zero length.
func (p *Point) Angle(q Point) float64 {
	den := p.Length() * q.Length()
	if den == 0 {
		return 0
	}
	cos := p.Dot(q) / den
	if cos > 1 {
		cos = 1
	} else if cos < -1 {
		cos = -1
	}
	return math.Acos(cos)
}

// Rotate returns a new Point rotated around the origin by the given
// angle in radians. Positive values rotate counterclockwise.
func (p *Point) Rotate(angle float64) *Point {
	s, c := math.Sincos(angle)
	return &Point{X: p.X*c - p.Y*s, Y: p.X*s + p.Y*c}
}

// Interpolate returns a linearly interpolated point between p and q.
// The parameter t (0–1) determines the position: t=0 → p, t=1 → q.
func (p *Point) Interpolate(q *Point, t float64) *Point {
	return &Point{
		X: p.X + (q.X-p.X)*t,
		Y: p.Y + (q.Y-p.Y)*t,
	}
}
