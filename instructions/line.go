package instructions

import (
	"image"
	"image/draw"
	"math"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
	"github.com/Krispeckt/glimo/internal/render"
	"github.com/golang/freetype/raster"
	"golang.org/x/image/math/fixed"
)

// LineCap defines how the end of a stroked line is rendered.
type LineCap int

const (
	// LineCapRound renders rounded end caps.
	LineCapRound LineCap = iota
	// LineCapButt renders flat end caps at the exact end of the path.
	LineCapButt
	// LineCapSquare renders square end caps that extend by half the line width.
	LineCapSquare
)

// LineJoin defines how the junction between two line segments is rendered.
type LineJoin int

const (
	// LineJoinRound renders rounded joins between segments.
	LineJoinRound LineJoin = iota
	// LineJoinBevel renders a beveled (cut-off) corner at joins.
	LineJoinBevel
)

// FillRule defines how interior regions of a path are determined.
type FillRule int

const (
	// FillRuleWinding uses the non-zero winding rule.
	FillRuleWinding FillRule = iota
	// FillRuleEvenOdd uses the even-odd rule.
	FillRuleEvenOdd
)

// Line is the public facade that exposes a vector drawing API.
type Line struct{ eng *engine }

// NewLine creates a new Line with default styles and identity transform.
func NewLine() *Line {
	return &Line{
		eng: &engine{
			lineCap:       LineCapButt,
			lineJoin:      LineJoinRound,
			fillRule:      FillRuleWinding,
			lineWidth:     1,
			matrix:        geom.Identity(),
			fillPattern:   patterns.NewSolid(colors.Transparent),
			strokePattern: patterns.NewSolid(colors.Black),
		},
	}
}

// WithMatrix sets the current transform matrix.
func (l *Line) WithMatrix(m geom.Matrix) *Line { l.eng.matrix = m; return l }

// ResetMatrix resets the transform matrix to identity.
func (l *Line) ResetMatrix() *Line { l.eng.matrix = geom.Identity(); return l }

// SetLineCap sets the cap style for strokes.
func (l *Line) SetLineCap(c LineCap) *Line { l.eng.lineCap = c; return l }

// SetLineJoin sets the join style for strokes.
func (l *Line) SetLineJoin(j LineJoin) *Line { l.eng.lineJoin = j; return l }

// SetFillRule sets the fill rule used to determine interior.
func (l *Line) SetFillRule(r FillRule) *Line { l.eng.fillRule = r; return l }

// SetLineWidth sets the stroke width in pixels.
func (l *Line) SetLineWidth(w float64) *Line { l.eng.lineWidth = w; return l }

// SetDashes sets dash lengths alternating on/off along the stroke.
func (l *Line) SetDashes(d []float64) *Line { l.eng.dashes = d; return l }

// SetDashOffset sets the initial dash offset along the path.
func (l *Line) SetDashOffset(off float64) *Line { l.eng.dashOffset = off; return l }

// SetStrokePattern sets the pattern used to paint strokes.
func (l *Line) SetStrokePattern(p patterns.Pattern) *Line { l.eng.strokePattern = p; return l }

// SetFillPattern sets the pattern used to paint fills.
func (l *Line) SetFillPattern(p patterns.Pattern) *Line { l.eng.fillPattern = p; return l }

// ResetMask clears any active clip mask.
func (l *Line) ResetMask() *Line { l.eng.mask = nil; return l }

// polyStart opens a new polyline for stroke dash processing.
func (e *engine) polyStart(p *Point) {
	e.strokePolylines = append(e.strokePolylines, []*Point{p})
}

// polyAppend appends points to the current polyline.
func (e *engine) polyAppend(ps ...*Point) {
	if len(e.strokePolylines) == 0 {
		e.strokePolylines = append(e.strokePolylines, []*Point{})
	}
	last := len(e.strokePolylines) - 1
	e.strokePolylines[last] = append(e.strokePolylines[last], ps...)
}

// polyCloseTo closes the current polyline to the given point.
func (e *engine) polyCloseTo(p *Point) {
	if len(e.strokePolylines) == 0 {
		return
	}
	last := len(e.strokePolylines) - 1
	e.strokePolylines[last] = append(e.strokePolylines[last], p)
}

// copyPolylines deep-copies a slice of polylines.
func copyPolylines(in [][]*Point) [][]*Point {
	out := make([][]*Point, len(in))
	for i := range in {
		out[i] = make([]*Point, len(in[i]))
		copy(out[i], in[i])
	}
	return out
}

// MoveTo starts a new subpath at (x, y). Closes the previous fill subpath if needed.
func (l *Line) MoveTo(x, y float64) *Line {
	e := l.eng
	if e.hasCurrent {
		e.fillPath.Add1(e.start.Fixed())
	}
	x, y = e.matrix.TransformPoint(x, y)
	p := NewPoint(x, y)
	e.strokePath.Start(p.Fixed())
	e.fillPath.Start(p.Fixed())
	e.polyStart(p)
	e.start = p
	e.current = p
	e.hasCurrent = true
	return l
}

// LineTo adds a straight segment to (x, y). Starts a subpath if none exists.
func (l *Line) LineTo(x, y float64) *Line {
	e := l.eng
	if !e.hasCurrent {
		return l.MoveTo(x, y)
	}
	x, y = e.matrix.TransformPoint(x, y)
	p := NewPoint(x, y)
	e.strokePath.Add1(p.Fixed())
	e.fillPath.Add1(p.Fixed())
	e.polyAppend(p)
	e.current = p
	return l
}

// QuadraticTo adds a quadratic Bézier curve and discretizes it for dashing.
func (l *Line) QuadraticTo(x1, y1, x2, y2 float64) *Line {
	e := l.eng
	if !e.hasCurrent {
		l.MoveTo(x1, y1)
	}
	x0, y0 := e.current.X, e.current.Y
	x1, y1 = e.matrix.TransformPoint(x1, y1)
	x2, y2 = e.matrix.TransformPoint(x2, y2)
	p1 := NewPoint(x1, y1)
	p2 := NewPoint(x2, y2)

	// Keep the curve in strokePath/fillPath for high-fidelity rasterization.
	e.strokePath.Add2(p1.Fixed(), p2.Fixed())
	e.fillPath.Add2(p1.Fixed(), p2.Fixed())

	// Discretize to polyline for dash processing.
	points := QuadraticBezier(x0, y0, x1, y1, x2, y2)
	if len(points) > 1 {
		e.polyAppend(points[1:]...)
	}
	e.current = p2
	return l
}

// CubicTo adds a cubic Bézier curve and discretizes it to a polyline.
func (l *Line) CubicTo(x1, y1, x2, y2, x3, y3 float64) *Line {
	e := l.eng
	if !e.hasCurrent {
		l.MoveTo(x1, y1)
	}
	x0, y0 := e.current.X, e.current.Y
	x1, y1 = e.matrix.TransformPoint(x1, y1)
	x2, y2 = e.matrix.TransformPoint(x2, y2)
	x3, y3 = e.matrix.TransformPoint(x3, y3)

	points := CubicBezier(x0, y0, x1, y1, x2, y2, x3, y3)
	previous := e.current.Fixed()
	for _, p := range points[1:] {
		f := p.Fixed()
		if f == previous {
			continue
		}
		previous = f
		e.strokePath.Add1(f)
		e.fillPath.Add1(f)
	}
	e.polyAppend(points[1:]...)
	e.current = NewPoint(x3, y3)
	return l
}

// ClosePath closes the current subpath by drawing a segment to the start point.
func (l *Line) ClosePath() *Line {
	e := l.eng
	if e.hasCurrent {
		e.strokePath.Add1(e.start.Fixed())
		e.fillPath.Add1(e.start.Fixed())
		e.polyCloseTo(e.start)
		e.current = e.start
	}
	return l
}

// ClearPath clears all paths and resets subpath state.
func (l *Line) ClearPath() *Line {
	e := l.eng
	e.strokePath.Clear()
	e.fillPath.Clear()
	e.strokePolylines = nil
	e.hasCurrent = false
	return l
}

// NewSubPath ends the current fill subpath and prepares for a new one.
func (l *Line) NewSubPath() *Line {
	e := l.eng
	if e.hasCurrent {
		e.fillPath.Add1(e.start.Fixed())
	}
	e.hasCurrent = false
	return l
}

// StrokePreserve schedules a stroke rasterization of the current path without clearing it.
func (l *Line) StrokePreserve() *Line {
	e := l.eng
	// Snapshot geometry and styles.
	spath := make(raster.Path, len(e.strokePath))
	copy(spath, e.strokePath)
	spoly := copyPolylines(e.strokePolylines)
	dashes := append([]float64(nil), e.dashes...)
	dashOffset := e.dashOffset
	lineWidth := e.lineWidth
	capper := e.capper()
	joiner := e.joiner()
	strokePat := e.strokePattern

	l.eng.pendingOps = append(l.eng.pendingOps, func(e2 *engine) {
		var painter raster.Painter

		useFast := false
		if solid, ok := strokePat.(*patterns.Solid); ok && e2.mask == nil {
			if bp, ok := strokePat.(patterns.BlendedPattern); ok {
				useFast = (bp.BlendMode() == patterns.BlendPassThrough) && (bp.Opacity() == 1)
			} else {
				useFast = true
			}
			if useFast {
				p := raster.NewRGBAPainter(e2.overlay)
				p.SetColor(solid.ColorAt(0, 0))
				painter = p
			}
		}
		if painter == nil {
			painter = render.NewPatternPainter(e2.overlay, e2.base, e2.mask, strokePat)
		}

		path := spath
		if len(dashes) > 0 {
			path = rasterPath(dashPath(spoly, dashes, dashOffset))
		}
		r := e2.rasterizer
		r.UseNonZeroWinding = true
		r.Clear()
		r.AddStroke(path, geom.Fix(lineWidth), capper, joiner)
		r.Rasterize(painter)
	})
	return l
}

// Stroke strokes the current path and then clears it.
func (l *Line) Stroke() *Line {
	return l.StrokePreserve().ClearPath()
}

// FillPreserve schedules a fill rasterization of the current path without clearing it.
func (l *Line) FillPreserve() *Line {
	e := l.eng
	// Snapshot geometry and styles.
	fpath := make(raster.Path, len(e.fillPath))
	copy(fpath, e.fillPath)
	hasCurrent := e.hasCurrent
	var start fixed.Point26_6
	if e.start != nil {
		start = e.start.Fixed()
	}
	fillPat := e.fillPattern
	fillRule := e.fillRule

	l.eng.pendingOps = append(l.eng.pendingOps, func(e2 *engine) {
		var painter raster.Painter

		useFast := false
		if solid, ok := fillPat.(*patterns.Solid); ok && e2.mask == nil {
			if bp, ok := fillPat.(patterns.BlendedPattern); ok {
				useFast = (bp.BlendMode() == patterns.BlendPassThrough) && (bp.Opacity() == 1)
			} else {
				useFast = true
			}
			if useFast {
				p := raster.NewRGBAPainter(e2.overlay)
				p.SetColor(solid.ColorAt(0, 0))
				painter = p
			}
		}
		if painter == nil {
			painter = render.NewPatternPainter(e2.overlay, e2.base, e2.mask, fillPat)
		}

		path := make(raster.Path, len(fpath))
		copy(path, fpath)
		if hasCurrent {
			path.Add1(start)
		}
		r := e2.rasterizer
		r.UseNonZeroWinding = fillRule == FillRuleWinding
		r.Clear()
		r.AddPath(path)
		r.Rasterize(painter)
	})
	return l
}

// Fill fills the current path and then clears it.
func (l *Line) Fill() *Line {
	return l.FillPreserve().ClearPath()
}

// ClipPreserve updates the clip mask by rasterizing the current fill path.
func (l *Line) ClipPreserve() *Line {
	e := l.eng
	// Snapshot geometry.
	fpath := make(raster.Path, len(e.fillPath))
	copy(fpath, e.fillPath)
	hasCurrent := e.hasCurrent
	var start fixed.Point26_6
	if e.start != nil {
		start = e.start.Fixed()
	}
	fillRule := e.fillRule

	l.eng.pendingOps = append(l.eng.pendingOps, func(e2 *engine) {
		clip := image.NewAlpha(image.Rect(0, 0, e2.width, e2.height))
		path := make(raster.Path, len(fpath))
		copy(path, fpath)
		if hasCurrent {
			path.Add1(start)
		}
		r := e2.rasterizer
		r.UseNonZeroWinding = fillRule == FillRuleWinding
		r.Clear()
		r.AddPath(path)
		r.Rasterize(raster.NewAlphaOverPainter(clip))

		if e2.mask != nil && e2.mask.Bounds() != clip.Bounds() {
			e2.mask = nil
		}
		if e2.mask == nil {
			e2.mask = clip
		} else {
			mask := image.NewAlpha(image.Rect(0, 0, e2.width, e2.height))
			draw.DrawMask(mask, mask.Bounds(), clip, image.Point{}, e2.mask, image.Point{}, draw.Over)
			e2.mask = mask
		}
	})
	return l
}

// Draw executes all pending raster operations onto the provided RGBA image.
func (l *Line) Draw(base, overlay *image.RGBA) {
	e := l.eng
	e.overlay = overlay
	e.base = base
	e.ensureRasterizer()
	for _, op := range e.pendingOps {
		op(e)
	}
	e.pendingOps = e.pendingOps[:0]
}

// quadratic evaluates a quadratic Bézier at parameter t in [0,1].
func quadratic(x0, y0, x1, y1, x2, y2, t float64) (x, y float64) {
	u := 1 - t
	a := u * u
	b := 2 * u * t
	c := t * t
	x = a*x0 + b*x1 + c*x2
	y = a*y0 + b*y1 + c*y2
	return
}

// QuadraticBezier discretizes a quadratic Bézier into a sequence of points.
func QuadraticBezier(x0, y0, x1, y1, x2, y2 float64) []*Point {
	l := math.Hypot(x1-x0, y1-y0) + math.Hypot(x2-x1, y2-y1)
	n := int(l + 0.5)
	if n < 4 {
		n = 4
	}
	d := float64(n) - 1
	result := make([]*Point, n)
	for i := 0; i < n; i++ {
		t := float64(i) / d
		x, y := quadratic(x0, y0, x1, y1, x2, y2, t)
		result[i] = NewPoint(x, y)
	}
	return result
}

// cubic evaluates a cubic Bézier at parameter t in [0,1].
func cubic(x0, y0, x1, y1, x2, y2, x3, y3, t float64) (x, y float64) {
	u := 1 - t
	a := u * u * u
	b := 3 * u * u * t
	c := 3 * u * t * t
	d := t * t * t
	x = a*x0 + b*x1 + c*x2 + d*x3
	y = a*y0 + b*y1 + c*y2 + d*y3
	return
}

// CubicBezier discretizes a cubic Bézier into a sequence of points.
func CubicBezier(x0, y0, x1, y1, x2, y2, x3, y3 float64) []*Point {
	l := math.Hypot(x1-x0, y1-y0) +
		math.Hypot(x2-x1, y2-y1) +
		math.Hypot(x3-x2, y3-y2)
	n := int(l + 0.5)
	if n < 4 {
		n = 4
	}
	d := float64(n) - 1
	result := make([]*Point, n)
	for i := 0; i < n; i++ {
		t := float64(i) / d
		x, y := cubic(x0, y0, x1, y1, x2, y2, x3, y3, t)
		result[i] = NewPoint(x, y)
	}
	return result
}

// dashPath applies dash pattern to polylines and returns visible segments.
func dashPath(paths [][]*Point, dashes []float64, offset float64) [][]*Point {
	var result [][]*Point
	if len(dashes) == 0 {
		return paths
	}
	if len(dashes) == 1 {
		dashes = append(dashes, dashes[0])
	}
	for _, path := range paths {
		if len(path) < 2 {
			continue
		}
		previous := path[0]
		pathIndex := 1
		dashIndex := 0
		segmentLength := 0.0

		// Normalize and apply initial offset into the dash cycle.
		if offset != 0 {
			var totalLength float64
			for _, dashLength := range dashes {
				totalLength += dashLength
			}
			offset = math.Mod(offset, totalLength)
			if offset < 0 {
				offset += totalLength
			}
			for i, dashLength := range dashes {
				offset -= dashLength
				if offset < 0 {
					dashIndex = i
					segmentLength = dashLength + offset
					break
				}
			}
		}

		segment := []*Point{previous}
		for pathIndex < len(path) {
			dashLength := dashes[dashIndex]
			point := path[pathIndex]
			d := previous.Distance(point)
			maxd := dashLength - segmentLength
			if d > maxd {
				t := maxd / d
				p := previous.Interpolate(point, t)
				segment = append(segment, p)
				if dashIndex%2 == 0 && len(segment) > 1 {
					result = append(result, segment)
				}
				segment = []*Point{p}
				segmentLength = 0
				previous = p
				dashIndex = (dashIndex + 1) % len(dashes)
			} else {
				segment = append(segment, point)
				previous = point
				segmentLength += d
				pathIndex++
			}
		}
		if dashIndex%2 == 0 && len(segment) > 1 {
			result = append(result, segment)
		}
	}
	return result
}

// rasterPath converts polylines to a raster.Path while dropping near-duplicate points.
func rasterPath(paths [][]*Point) raster.Path {
	var result raster.Path
	for _, path := range paths {
		var previous fixed.Point26_6
		for i, point := range path {
			f := point.Fixed()
			if i == 0 {
				result.Start(f)
			} else {
				dx := f.X - previous.X
				dy := f.Y - previous.Y
				if dx < 0 {
					dx = -dx
				}
				if dy < 0 {
					dy = -dy
				}
				// Avoid adding points that are too close in fixed space.
				if dx+dy > 8 {
					result.Add1(f)
				}
			}
			previous = f
		}
	}
	return result
}
