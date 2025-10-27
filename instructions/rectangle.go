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

// StrokePosition defines stroke alignment relative to the rectangle border.
type StrokePosition int

const (
	StrokeInside StrokePosition = iota
	StrokeCenter
	StrokeOutside
)

func (s StrokePosition) Outset(lineWidth float64) float64 {
	if lineWidth <= 0 {
		return 0
	}
	switch s {
	case StrokeInside:
		return 0
	case StrokeCenter:
		return lineWidth / 2
	case StrokeOutside:
		return lineWidth
	default:
		return 0
	}
}

// Rectangle represents a drawable rectangle with per-corner rounding, fill, and stroke.
type Rectangle struct {
	x, y          float64
	width, height float64

	radiusTL float64
	radiusTR float64
	radiusBR float64
	radiusBL float64

	fillPattern   patterns.Pattern
	strokePattern patterns.Pattern
	lineWidth     float64
	strokePos     StrokePosition
	roundSteps    int

	effects containers.Effects
}

// NewRectangle creates a new rectangle with the given position and size.
func NewRectangle(x, y, width, height float64) *Rectangle {
	return &Rectangle{
		x:             x,
		y:             y,
		width:         width,
		height:        height,
		fillPattern:   patterns.NewSolid(colors.Transparent),
		strokePattern: patterns.NewSolid(colors.Transparent),
		lineWidth:     1,
		strokePos:     StrokeInside,
		roundSteps:    8,
		effects:       containers.Effects{},
	}
}

// SetSize sets width and height.
func (r *Rectangle) SetSize(width, height float64) *Rectangle {
	r.width, r.height = width, height
	return r
}

// SetRadius sets uniform corner radius for all corners.
func (r *Rectangle) SetRadius(radius float64) *Rectangle {
	r.radiusTL = geom.ClampF64(radius, 0, math.MaxFloat64)
	r.radiusTR = geom.ClampF64(radius, 0, math.MaxFloat64)
	r.radiusBR = geom.ClampF64(radius, 0, math.MaxFloat64)
	r.radiusBL = geom.ClampF64(radius, 0, math.MaxFloat64)
	return r
}

// SetCornerRadii sets per-corner radii: top-left, top-right, bottom-right, bottom-left.
func (r *Rectangle) SetCornerRadii(tl, tr, br, bl float64) *Rectangle {
	r.radiusTL = geom.ClampF64(tl, 0, math.MaxFloat64)
	r.radiusTR = geom.ClampF64(tr, 0, math.MaxFloat64)
	r.radiusBR = geom.ClampF64(br, 0, math.MaxFloat64)
	r.radiusBL = geom.ClampF64(bl, 0, math.MaxFloat64)
	return r
}

// SetLineWidth sets stroke width.
func (r *Rectangle) SetLineWidth(width float64) *Rectangle {
	r.lineWidth = geom.ClampF64(width, 0, math.MaxFloat64)
	return r
}

// SetStrokePosition defines whether stroke is drawn inside, centered, or outside the border.
func (r *Rectangle) SetStrokePosition(pos StrokePosition) *Rectangle {
	r.strokePos = pos
	return r
}

// SetRoundedSteps sets resolution of rounded arcs.
func (r *Rectangle) SetRoundedSteps(steps int) *Rectangle {
	if steps < 1 {
		steps = 1
	}
	r.roundSteps = steps
	return r
}

// SetFillColor sets solid fill colorPattern.
func (r *Rectangle) SetFillColor(c patterns.Color) *Rectangle {
	r.fillPattern = c.MakeSolidPattern()
	return r
}

// SetFillPattern sets fill pattern.
func (r *Rectangle) SetFillPattern(p patterns.Pattern) *Rectangle {
	if p != nil {
		r.fillPattern = p
	}
	return r
}

// SetStrokeColor sets solid stroke colorPattern.
func (r *Rectangle) SetStrokeColor(c patterns.Color) *Rectangle {
	r.strokePattern = c.MakeSolidPattern()
	return r
}

// SetStrokePattern sets stroke pattern.
func (r *Rectangle) SetStrokePattern(p patterns.Pattern) *Rectangle {
	if p != nil {
		r.strokePattern = p
	}
	return r
}

// AddEffect attaches a visual effect to the rectangle rendering pipeline.
//
// The added effect will be stored inside the internal effect container `t.effects`
// and applied automatically during rendering:
//   - Pre-effects (e.IsPre() == true) execute before text drawing.
//   - Post-effects (e.IsPre() == false) execute after text drawing.
//
// This allows chaining multiple visual transformations such as
// drop shadows, layer blurs, and noise filters.
//
// Example:
//
//	rectangle.AddEffect(effects.NewDropShadow(0, 4, 4, 0, colors.Black, 0.25))
func (r *Rectangle) AddEffect(e effects.Effect) *Rectangle {
	r.effects.Add(e)
	return r
}

// AddEffects attaches multiple visual effects to the pipeline.
func (r *Rectangle) AddEffects(es ...effects.Effect) *Rectangle {
	r.effects.AddList(es)
	return r
}

// SetPosition sets the rectangle’s top-left corner.
func (r *Rectangle) SetPosition(x, y int) {
	r.x, r.y = float64(x), float64(y)
}

// Size returns rectangle size.
func (r *Rectangle) Size() *geom.Size {
	o := r.strokePos.Outset(r.lineWidth)
	return geom.NewSize(r.width+o, r.height+o)
}

// Position returns the top-left coordinate where the layer is drawn.
func (r *Rectangle) Position() (int, int) { return int(r.x), int(r.y) }

// Draw renders the rectangle with stroke alignment (inside, center, outside).
func (r *Rectangle) Draw(base, overlay *image.RGBA) {
	if r.width <= 0 || r.height <= 0 {
		return
	}

	offset := 0.0
	switch r.strokePos {
	case StrokeInside:
		offset = r.lineWidth / 2
	case StrokeOutside:
		offset = -r.lineWidth / 2
	default:
		offset = 0
	}

	r.effects.PreApplyAll(overlay)

	line := NewLine().
		SetLineWidth(r.lineWidth).
		SetStrokePattern(r.strokePattern).
		SetFillPattern(r.fillPattern)

	addRoundedRectCorners(
		line,
		r.x+offset,
		r.y+offset,
		r.width-2*offset,
		r.height-2*offset,
		r.radiusTL-offset,
		r.radiusTR-offset,
		r.radiusBR-offset,
		r.radiusBL-offset,
		r.roundSteps,
	)

	line.FillPreserve()
	if r.lineWidth > 0 && r.strokePattern != nil {
		line.StrokePreserve()
	}
	line.Draw(base, overlay)

	r.effects.PostApplyAll(overlay)
}

// addRoundedRectCorners draws rectangle with per-corner radii.
func addRoundedRectCorners(line *Line, x, y, w, h, rtl, rtr, rbr, rbl float64, steps int) {
	clamp := func(v float64) float64 { return geom.ClampF64(v, 0, math.MaxFloat64) }

	rtl = clamp(rtl)
	rtr = clamp(rtr)
	rbr = clamp(rbr)
	rbl = clamp(rbl)

	maxR := func(r, max float64) float64 {
		if r > max {
			return max
		}
		return r
	}
	rtl = maxR(rtl, math.Min(w/2, h/2))
	rtr = maxR(rtr, math.Min(w/2, h/2))
	rbr = maxR(rbr, math.Min(w/2, h/2))
	rbl = maxR(rbl, math.Min(w/2, h/2))

	cxTL, cyTL := x+rtl, y+rtl
	cxTR, cyTR := x+w-rtr, y+rtr
	cxBR, cyBR := x+w-rbr, y+h-rbr
	cxBL, cyBL := x+rbl, y+h-rbl

	line.MoveTo(x+rtl, y)
	line.LineTo(x+w-rtr, y)
	if rtr > 0 {
		addQuarterArc(line, cxTR, cyTR, rtr, 270, 360, steps)
	}
	line.LineTo(x+w, y+h-rbr)
	if rbr > 0 {
		addQuarterArc(line, cxBR, cyBR, rbr, 0, 90, steps)
	}
	line.LineTo(x+rbl, y+h)
	if rbl > 0 {
		addQuarterArc(line, cxBL, cyBL, rbl, 90, 180, steps)
	}
	line.LineTo(x, y+rtl)
	if rtl > 0 {
		addQuarterArc(line, cxTL, cyTL, rtl, 180, 270, steps)
	}
	line.ClosePath()
}

// addQuarterArc approximates a quarter circle with line segments.
func addQuarterArc(line *Line, cx, cy, r float64, degStart, degEnd, steps int) {
	if r <= 0 || steps < 1 {
		return
	}
	ds := float64(degStart)
	de := float64(degEnd)
	step := (de - ds) / float64(steps)
	for i := 1; i <= steps; i++ {
		a := (ds + float64(i)*step) * math.Pi / 180
		x := cx + r*math.Cos(a)
		y := cy + r*math.Sin(a)
		line.LineTo(x, y)
	}
}
