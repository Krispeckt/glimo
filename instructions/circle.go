package instructions

import (
	"image"
	"math"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/effects"
	"github.com/Krispeckt/glimo/internal/containers"
	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

// Circle represents a drawable circle with fill, stroke, and optional effects.
type Circle struct {
	x, y      float64 // top-left corner
	radius    float64
	fill      patterns.Pattern
	stroke    patterns.Pattern
	lineWidth float64
	strokePos StrokePosition
	steps     int
	effects   containers.Effects
}

// NewCircle creates a new circle with given top-left and radius.
func NewCircle(x, y, radius float64) *Circle {
	return &Circle{
		x:         x,
		y:         y,
		radius:    geom.ClampF64(radius, 0, math.MaxFloat64),
		fill:      patterns.NewSolid(colors.Transparent),
		stroke:    patterns.NewSolid(colors.Transparent),
		lineWidth: 1,
		strokePos: StrokeInside,
		steps:     32,
		effects:   containers.Effects{},
	}
}

// SetRadius sets circle radius.
func (c *Circle) SetRadius(r float64) *Circle {
	c.radius = geom.ClampF64(r, 0, math.MaxFloat64)
	return c
}

// SetLineWidth sets stroke width.
func (c *Circle) SetLineWidth(width float64) *Circle {
	c.lineWidth = geom.ClampF64(width, 0, math.MaxFloat64)
	return c
}

// SetStrokePosition defines stroke alignment: inside, center, outside.
func (c *Circle) SetStrokePosition(pos StrokePosition) *Circle {
	c.strokePos = pos
	return c
}

// SetSteps sets resolution for circle approximation.
func (c *Circle) SetSteps(steps int) *Circle {
	if steps < 3 {
		steps = 3
	}
	c.steps = steps
	return c
}

// SetFillColor sets solid fill color.
func (c *Circle) SetFillColor(col patterns.Color) *Circle {
	c.fill = col.MakeSolidPattern()
	return c
}

// SetFillPattern sets custom fill pattern.
func (c *Circle) SetFillPattern(p patterns.Pattern) *Circle {
	if p != nil {
		c.fill = p
	}
	return c
}

// SetStrokeColor sets solid stroke color.
func (c *Circle) SetStrokeColor(col patterns.Color) *Circle {
	c.stroke = col.MakeSolidPattern()
	return c
}

// SetStrokePattern sets custom stroke pattern.
func (c *Circle) SetStrokePattern(p patterns.Pattern) *Circle {
	if p != nil {
		c.stroke = p
	}
	return c
}

// AddEffect adds a visual effect (blur, shadow, etc.) to the circle.
func (c *Circle) AddEffect(e effects.Effect) *Circle {
	c.effects.Add(e)
	return c
}

// AddEffects adds multiple visual effects.
func (c *Circle) AddEffects(es ...effects.Effect) *Circle {
	c.effects.AddList(es)
	return c
}

// SetPosition sets top-left position of circle bounding box.
func (c *Circle) SetPosition(x, y int) {
	c.x, c.y = float64(x), float64(y)
}

// Size returns circle diameter as Size.
func (c *Circle) Size() *geom.Size {
	d := c.radius * 2
	o := c.strokePos.Outset(c.lineWidth)
	return geom.NewSize(d+o, d+o)
}

// Position returns top-left of bounding box.
func (c *Circle) Position() (int, int) {
	return int(c.x), int(c.y)
}

// Draw renders the circle to the overlay.
func (c *Circle) Draw(base, overlay *image.RGBA) {
	if c.radius <= 0 {
		return
	}

	offset := 0.0
	switch c.strokePos {
	case StrokeInside:
		offset = c.lineWidth / 2
	case StrokeOutside:
		offset = -c.lineWidth / 2
	default:
		offset = 0
	}

	c.effects.PreApplyAll(overlay)

	r := geom.MaxF64(c.radius-offset, 0)
	// convert top-left to center
	cx := c.x + c.radius
	cy := c.y + c.radius

	line := NewLine().
		SetLineWidth(c.lineWidth).
		SetStrokePattern(c.stroke).
		SetFillPattern(c.fill)

	addCirclePath(line, cx, cy, r, c.steps)

	line.FillPreserve()
	if c.lineWidth > 0 && c.stroke != nil {
		line.StrokePreserve()
	}
	line.Draw(base, overlay)

	c.effects.PostApplyAll(overlay)
}

// addCirclePath approximates circle using polygonal segments.
func addCirclePath(line *Line, cx, cy, r float64, steps int) {
	if r <= 0 || steps < 3 {
		return
	}
	step := 2 * math.Pi / float64(steps)
	line.MoveTo(cx+r, cy)
	for i := 1; i <= steps; i++ {
		a := float64(i) * step
		x := cx + r*math.Cos(a)
		y := cy + r*math.Sin(a)
		line.LineTo(x, y)
	}
	line.ClosePath()
}
