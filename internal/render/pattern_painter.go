package render

import (
	"image"

	"github.com/Krispeckt/glimo/internal/core/image/patterns"
	"github.com/golang/freetype/raster"
)

// PatternPainter implements the freetype/raster.Painter interface.
// It renders spans into an RGBA overlay using a pattern fill, blending
// with a base image and an optional alpha mask.
//
// It supports different blending modes and opacity control when the
// pattern implements the `patterns.BlendedPattern` interface.
type PatternPainter struct {
	overlay, base *image.RGBA  // Target overlay and base layers
	mask          *image.Alpha // Optional alpha mask for coverage control
	pattern       patterns.Pattern
}

// Paint renders a list of raster spans (`ss`) onto the overlay image.
// Each span is filled using the current pattern, blended with the base
// image according to the pattern's blend mode and opacity.
//
// If a mask is provided, it modulates the per-pixel alpha coverage.
// This function is typically called by a rasterizer during vector path filling.
func (r *PatternPainter) Paint(ss []raster.Span, _ bool) {
	b := r.overlay.Bounds()

	// Default blending mode and opacity
	mode := patterns.BlendNormal
	opacity := 1.0

	// If the pattern supports blending, extract its mode and opacity
	if bp, ok := r.pattern.(patterns.BlendedPattern); ok {
		mode = bp.BlendMode()
		opacity = bp.Opacity()
	}

	const m = 1<<16 - 1 // Maximum alpha value used by raster.Span

	for _, s := range ss {
		// Skip spans outside vertical bounds
		if s.Y < b.Min.Y {
			continue
		}
		if s.Y >= b.Max.Y {
			return
		}

		// Clamp horizontal span to image bounds
		if s.X0 < b.Min.X {
			s.X0 = b.Min.X
		}
		if s.X1 > b.Max.X {
			s.X1 = b.Max.X
		}
		if s.X0 >= s.X1 {
			continue
		}

		y := s.Y - r.overlay.Rect.Min.Y
		x0 := s.X0 - r.overlay.Rect.Min.X
		i0 := (s.Y-r.overlay.Rect.Min.Y)*r.overlay.Stride + (s.X0-r.overlay.Rect.Min.X)*4
		i1 := i0 + (s.X1-s.X0)*4

		// Process each pixel in the span
		for i, x := i0, x0; i < i1; i, x = i+4, x+1 {
			ma := s.Alpha

			// Apply mask if present
			if r.mask != nil {
				ma = ma * uint32(r.mask.AlphaAt(x, y).A) / 255
				if ma == 0 {
					continue
				}
			}

			// Sample pattern color
			col := r.pattern.ColorAt(x, y)
			src, ok := col.(patterns.Color)
			if !ok {
				src = patterns.NewColorFromStd(col)
			}

			// Apply inherited blend mode if "PassThrough"
			if src.BlendMode().String() == "PassThrough" {
				src = src.SetBlendMode(mode)
			}

			// Read background color
			bg := patterns.Color{
				R: r.base.Pix[i+0],
				G: r.base.Pix[i+1],
				B: r.base.Pix[i+2],
				A: r.base.Pix[i+3],
			}

			coverage := float64(ma) / float64(m)
			blended := src.BlendOver(bg, coverage)

			// Apply opacity blending to the overlay
			if opacity >= 1.0 {
				r.overlay.Pix[i+0] = blended.R
				r.overlay.Pix[i+1] = blended.G
				r.overlay.Pix[i+2] = blended.B
				r.overlay.Pix[i+3] = blended.A
			} else {
				inv := 1.0 - opacity
				r.overlay.Pix[i+0] = uint8(inv*float64(bg.R) + opacity*float64(blended.R) + 0.5)
				r.overlay.Pix[i+1] = uint8(inv*float64(bg.G) + opacity*float64(blended.G) + 0.5)
				r.overlay.Pix[i+2] = uint8(inv*float64(bg.B) + opacity*float64(blended.B) + 0.5)
				r.overlay.Pix[i+3] = uint8(inv*float64(bg.A) + opacity*float64(blended.A) + 0.5)
			}
		}
	}
}

// NewPatternPainter creates and returns a new PatternPainter.
//
// Parameters:
//   - overlay: destination RGBA layer where the pattern will be painted
//   - base: background RGBA layer to blend over
//   - mask: optional alpha mask (can be nil)
//   - p: pattern implementing the patterns.Pattern interface
func NewPatternPainter(overlay, base *image.RGBA, mask *image.Alpha, p patterns.Pattern) *PatternPainter {
	return &PatternPainter{overlay, base, mask, p}
}
