package instructions

import (
	"image"

	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
	"github.com/golang/freetype/raster"
)

// engine is the rasterization backend that holds drawing state and buffers.
type engine struct {
	rasterizer *raster.Rasterizer

	lineCap       LineCap
	lineJoin      LineJoin
	fillRule      FillRule
	hasCurrent    bool
	strokePath    raster.Path
	fillPath      raster.Path
	start         *Point
	current       *Point
	dashes        []float64
	dashOffset    float64
	lineWidth     float64
	mask          *image.Alpha
	fillPattern   patterns.Pattern
	strokePattern patterns.Pattern

	strokePolylines [][]*Point

	matrix geom.Matrix

	base, overlay *image.RGBA
	width, height int

	pendingOps []func(e *engine)
}

// ensureRasterizer initializes or resizes the rasterizer to match the target image.
func (e *engine) ensureRasterizer() {
	if e.overlay == nil {
		return
	}
	w, h := e.overlay.Bounds().Dx(), e.overlay.Bounds().Dy()
	if e.rasterizer == nil || w != e.width || h != e.height {
		e.width, e.height = w, h
		e.rasterizer = raster.NewRasterizer(w, h)
		if e.mask != nil && e.mask.Bounds() != e.overlay.Bounds() {
			e.mask = nil
		}
	}
}

// capper returns the raster.Capper implementation for the current line cap.
func (e *engine) capper() raster.Capper {
	switch e.lineCap {
	case LineCapButt:
		return raster.ButtCapper
	case LineCapRound:
		return raster.RoundCapper
	case LineCapSquare:
		return raster.SquareCapper
	}
	return nil
}

// joiner returns the raster.Joiner implementation for the current line join.
func (e *engine) joiner() raster.Joiner {
	switch e.lineJoin {
	case LineJoinBevel:
		return raster.BevelJoiner
	case LineJoinRound:
		return raster.RoundJoiner
	}
	return nil
}

// stroke rasterizes the current stroke path using the provided painter.
func (e *engine) stroke(painter raster.Painter) {
	r := e.rasterizer
	r.UseNonZeroWinding = true
	r.Clear()
	r.AddStroke(e.strokePath, geom.Fix(e.lineWidth), e.capper(), e.joiner())
	r.Rasterize(painter)
}

// fill rasterizes the current fill path using the provided painter and fill rule.
func (e *engine) fill(painter raster.Painter) {
	path := e.fillPath
	if e.hasCurrent {
		path = make(raster.Path, len(e.fillPath))
		copy(path, e.fillPath)
		path.Add1(e.start.Fixed())
	}
	r := e.rasterizer
	r.UseNonZeroWinding = e.fillRule == FillRuleWinding
	r.Clear()
	r.AddPath(path)
	r.Rasterize(painter)
}
