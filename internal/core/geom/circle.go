package geom

import "math"

// Circle represents a 2D circle with a center position and radius as float64.
// It is used for geometric operations, bounds calculations, and collision checks.
type Circle struct {
	x, y   float64
	radius float64
}

// NewCircle creates a new Circle with the given center coordinates and radius.
func NewCircle(x, y, radius float64) *Circle {
	return &Circle{x: x, y: y, radius: radius}
}

// X returns the X coordinate of the circle center.
func (c *Circle) X() float64 {
	return c.x
}

// Y returns the Y coordinate of the circle center.
func (c *Circle) Y() float64 {
	return c.y
}

// Radius returns the circle radius.
func (c *Circle) Radius() float64 {
	return c.radius
}

// Set updates the circle center and radius.
func (c *Circle) Set(x, y, radius float64) {
	c.x = x
	c.y = y
	c.radius = radius
}

// Area returns the circle area using πr².
func (c *Circle) Area() float64 {
	return math.Pi * c.radius * c.radius
}

// Circumference returns the circle circumference using 2πr.
func (c *Circle) Circumference() float64 {
	return 2 * math.Pi * c.radius
}

// Diameter returns the circle diameter (2r).
func (c *Circle) Diameter() float64 {
	return 2 * c.radius
}

// Contains checks whether a given point (px, py) lies inside the circle.
func (c *Circle) Contains(px, py float64) bool {
	dx := px - c.x
	dy := py - c.y
	return dx*dx+dy*dy <= c.radius*c.radius
}

// IsZero checks whether the circle has a zero radius.
func (c *Circle) IsZero() bool {
	return c.radius == 0
}

// Copy returns a deep copy of the circle.
func (c *Circle) Copy() *Circle {
	return &Circle{x: c.x, y: c.y, radius: c.radius}
}
