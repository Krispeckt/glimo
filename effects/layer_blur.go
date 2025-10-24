// Package effects defines visual post-processing effects for 2D drawing instructions.
// These effects operate directly on image buffers (*image.RGBA) after or before
// the main rendering pass, allowing simulation of optical phenomena such as blur,
// shadows, glow, or texture overlays.
//
// This file implements a *layer blur* effect equivalent to Figma’s “Layer Blur”.
// Unlike a background blur, this effect acts solely on the content of the current
// layer — it diffuses the layer’s own pixels without sampling or blending the
// background underneath.
//
// High-level algorithm
//
//  1. Copy the current layer (dst) into a temporary buffer to preserve original data.
//  2. Apply one or more passes of separable box blur to spread pixel values.
//     - In constant mode, a uniform blur radius is used for the entire image.
//     - In progressive mode, blur radius is linearly interpolated along the Y-axis,
//     producing a gradient blur (useful for atmospheric or focus effects).
//  3. Optionally multiply all alpha values by `opacity` ∈ [0, 1] to attenuate
//     the blurred result.
//
// Implementation details
//
//   - Blur kernel: A three-pass box filter is used to approximate Gaussian blur.
//     Each pass performs horizontal and vertical averaging within ±radius.
//   - Progressive blur: Implemented by reusing `boxBlurLine` on each row with a
//     per-line radius computed via linear interpolation between `radiusStart`
//     and `radiusEnd`.
//   - Image format: Operates in sRGB 8-bit space directly. Alpha is blurred
//     along with RGB. The result remains premultiplied-compatible.
//   - Performance: Complexity is O(W*H*radius). For large images or high radii,
//     consider splitting into tiles or parallelizing.
//   - Thread-safety: Concurrent writes to the same *image.RGBA are unsafe.
//     Synchronize access when applying effects in parallel.
//
// Example usage:
//
//	blur := effects.NewLayerBlurEffect(8).SetOpacity(0.8)
//	blur.Apply(layer)
//
//	// Progressive blur from 4px (top) to 16px (bottom)
//	blur.SetProgressive(4, 16).Apply(layer)
//
// Both modes produce results visually similar to Figma’s layer blur.
package effects

import (
	"image"
	"math"

	"golang.org/x/image/draw"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// LayerBlurEffect applies a blur to the pixels of the current layer itself.
//
// It can operate either with a constant blur radius or progressively —
// increasing or decreasing blur intensity along the Y-axis.
//
// The implementation uses a simple multi-pass box blur approximation
// of a Gaussian blur, applied directly to RGBA pixel data.
//
// Parameters:
//   - radiusStart, radiusEnd — starting and ending radius of blur (for progressive mode).
//   - progressive — when true, interpolates radius vertically across image height.
//   - opacity — final opacity multiplier applied to the blurred layer.
//
// Notes:
//   - Works as a post-effect (IsPre() == false).
//   - Computational cost grows linearly with radius and pixel count.
//   - Not thread-safe when used on the same *image.RGBA concurrently.
type LayerBlurEffect struct {
	radiusStart float64
	radiusEnd   float64
	progressive bool
	opacity     float64
}

// NewLayerBlurEffect creates a new layer blur with constant radius.
//
// Example:
//
//	e := effects.NewLayerBlurEffect(8).SetOpacity(0.8)
//	layer.AddEffect(e)
func NewLayerBlurEffect(radius float64) *LayerBlurEffect {
	return &LayerBlurEffect{
		radiusStart: radius,
		radiusEnd:   radius,
		opacity:     1.0,
	}
}

// SetProgressive enables progressive blur and sets start/end radii.
//
// When active, blur strength varies linearly from start at the top to end at the bottom.
// Returns the receiver for chaining.
func (e *LayerBlurEffect) SetProgressive(start, end float64) *LayerBlurEffect {
	e.progressive = true
	e.radiusStart = start
	e.radiusEnd = end
	return e
}

// SetOpacity sets the layer’s alpha scaling factor in [0,1].
// This reduces the overall strength of the blurred image.
func (e *LayerBlurEffect) SetOpacity(v float64) *LayerBlurEffect {
	e.opacity = geom.ClampF64(v, 0, 1)
	return e
}

// Name returns the human-readable identifier of this effect.
func (e *LayerBlurEffect) Name() string {
	return "LayerBlur"
}

// IsPre indicates whether this effect runs before drawing.
// LayerBlurEffect is post-applied, so this returns false.
func (e *LayerBlurEffect) IsPre() bool {
	return false
}

// Apply executes the blur operation on the destination image.
//
// Steps:
//  1. Copy the current layer to a temporary source buffer.
//  2. Apply either a uniform blur (boxBlur) or progressive per-line blur (boxBlurLine).
//  3. If opacity < 1, scale the alpha channel accordingly.
//
// The blur uses three passes of box filtering (horizontal + vertical)
// to approximate Gaussian blur.
func (e *LayerBlurEffect) Apply(dst *image.RGBA) {
	b := dst.Bounds()
	src := image.NewRGBA(b)
	draw.Copy(src, image.Point{}, dst, b, draw.Src, nil)

	radius := e.radiusStart
	if e.progressive {
		h := b.Dy()
		for y := 0; y < h; y++ {
			t := float64(y) / float64(h-1)
			r := geom.Lerp(e.radiusStart, e.radiusEnd, t)
			boxBlurLine(dst, src, int(math.Max(1, r)), y)
		}
	} else {
		boxBlur(dst, src, int(math.Max(1, radius)))
	}
	if e.opacity < 1.0 {
		applyOpacity(dst, e.opacity)
	}
}

// boxBlur applies a three-pass box blur approximation over the entire image.
//
// The function alternates horizontal and vertical passes to diffuse
// color and alpha values evenly. Each pass averages neighboring pixels
// within ±radius in both directions.
func boxBlur(dst, src *image.RGBA, radius int) {
	if radius < 1 {
		return
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	tmp := image.NewRGBA(src.Bounds())

	for pass := 0; pass < 3; pass++ {
		// horizontal pass
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				var rs, gs, bs, as, count int
				for k := -radius; k <= radius; k++ {
					xk := geom.ClampF64(float64(x+k), 0, float64(w-1))
					i := int(xk) + y*w
					r, g, b, a := src.Pix[i*4+0], src.Pix[i*4+1], src.Pix[i*4+2], src.Pix[i*4+3]
					rs += int(r)
					gs += int(g)
					bs += int(b)
					as += int(a)
					count++
				}
				idx := (y*w + x) * 4
				tmp.Pix[idx+0] = uint8(rs / count)
				tmp.Pix[idx+1] = uint8(gs / count)
				tmp.Pix[idx+2] = uint8(bs / count)
				tmp.Pix[idx+3] = uint8(as / count)
			}
		}
		// vertical pass
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				var rs, gs, bs, as, count int
				for k := -radius; k <= radius; k++ {
					yk := geom.ClampF64(float64(y+k), 0, float64(h-1))
					i := x + int(yk)*w
					r, g, b, a := tmp.Pix[i*4+0], tmp.Pix[i*4+1], tmp.Pix[i*4+2], tmp.Pix[i*4+3]
					rs += int(r)
					gs += int(g)
					bs += int(b)
					as += int(a)
					count++
				}
				idx := (y*w + x) * 4
				dst.Pix[idx+0] = uint8(rs / count)
				dst.Pix[idx+1] = uint8(gs / count)
				dst.Pix[idx+2] = uint8(bs / count)
				dst.Pix[idx+3] = uint8(as / count)
			}
		}
	}
}

// boxBlurLine applies a one-dimensional blur along a single horizontal line.
// Useful for progressive blur where radius changes with Y.
func boxBlurLine(dst, src *image.RGBA, radius, lineY int) {
	b := src.Bounds()
	w := b.Dx()
	if radius < 1 || lineY < 0 || lineY >= b.Dy() {
		return
	}
	for x := 0; x < w; x++ {
		var rs, gs, bs, as, count int
		for k := -radius; k <= radius; k++ {
			yk := geom.ClampF64(float64(lineY+k), 0, float64(b.Dy()-1))
			i := x + int(yk)*w
			r, g, b, a := src.Pix[i*4+0], src.Pix[i*4+1], src.Pix[i*4+2], src.Pix[i*4+3]
			rs += int(r)
			gs += int(g)
			bs += int(b)
			as += int(a)
			count++
		}
		idx := (lineY*w + x) * 4
		dst.Pix[idx+0] = uint8(rs / count)
		dst.Pix[idx+1] = uint8(gs / count)
		dst.Pix[idx+2] = uint8(bs / count)
		dst.Pix[idx+3] = uint8(as / count)
	}
}

// applyOpacity scales all alpha channel values by the given factor.
//
// This provides a simple way to fade the entire blurred layer without
// re-rendering. Alpha is multiplied by factor ∈ [0,1].
func applyOpacity(img *image.RGBA, alpha float64) {
	if alpha >= 1.0 {
		return
	}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			i := img.PixOffset(x, y)
			img.Pix[i+3] = uint8(float64(img.Pix[i+3]) * alpha)
		}
	}
}
