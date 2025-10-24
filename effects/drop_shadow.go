// Package effects defines post- and pre-render visual filters for 2D drawing
// operations. These filters operate directly on *image.RGBA buffers, modifying
// pixel data to simulate optical effects such as shadows, blurs, glow, and
// color transformations.
//
// This file implements a *drop shadow* effect similar to CSS’s `box-shadow`,
// applying a blurred, offset shadow below rendered content. The effect uses
// the existing alpha channel of the layer as a shape mask and tints it with
// a specified shadow color, spread, and blur radius.
//
// High-level algorithm
//
//  1. Extract the alpha channel from the destination buffer (dst) to form a mask
//     representing the visible shape of the layer.
//  2. Optionally expand (spread) the mask to enlarge the opaque region before blur.
//  3. Apply a separable box blur to the alpha mask to soften the edges.
//  4. Tint the blurred alpha mask with the shadow color and multiply it by
//     overall opacity. RGB values are premultiplied by alpha to maintain
//     correct compositing.
//  5. Render the resulting shadow offset by (x, y) *below* the original content,
//     then composite the source image back over it.
//
// Notes and constraints
//
//   - Image format: The function expects *image.RGBA with premultiplied alpha.
//     Alpha blending is handled manually by setting premultiplied RGB values.
//   - Spread: Expands alpha before blur, similar to CSS `box-shadow spread-radius`.
//   - Blur: Implemented as a separable two-pass box blur for performance.
//     The result approximates a Gaussian blur with radius `blur`.
//   - Opacity: The `opacity` parameter multiplies the color alpha and scales
//     the shadow’s final visibility. Range [0..1].
//   - Composition order: The shadow is drawn first, then the original layer
//     is overlaid using Porter-Duff `draw.Over` blending.
//   - Performance: O(W*H*blur) complexity. For large radii, consider downsampling.
//   - Thread-safety: Concurrent writes to the same RGBA buffer are unsafe.
//
// Example usage:
//
//	shadow := effects.NewDropShadow(4, 4, 6, 0, color.NRGBA{0, 0, 0, 255}, 0.6)
//	shadow.Apply(layer)
//
// This produces a smooth, semi-transparent shadow shifted 4px down and right,
// visually equivalent to CSS `box-shadow: 4px 4px 6px rgba(0,0,0,0.6);`.
package effects

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// DropShadowEffect adds a blurred, offset shadow below the rendered object.
// It works similarly to CSS `box-shadow` for rasterized layers, using
// the layer’s alpha channel as the shape mask.
type DropShadowEffect struct {
	x, y    float64     // Shadow offset in pixels
	blur    float64     // Blur radius
	spread  float64     // Alpha expansion before blur
	color   color.Color // Shadow tint color
	opacity float64     // Overall opacity [0..1]
}

// NewDropShadow creates a configured drop shadow effect using offset,
// blur radius, spread, color, and opacity.
//
// Arguments:
//   - x, y: shadow offset in pixels (positive = right/down).
//   - blur: blur radius for edge softness.
//   - spread: expansion radius before blurring.
//   - col: shadow color (any color.Color).
//   - opacity: final opacity factor [0..1].
func NewDropShadow(x, y, blur, spread float64, col color.Color, opacity float64) *DropShadowEffect {
	return &DropShadowEffect{
		x: x, y: y,
		blur: blur, spread: spread,
		color: col, opacity: geom.ClampF64(opacity, 0, 1),
	}
}

// Name returns the effect identifier.
func (e *DropShadowEffect) Name() string { return "DropShadow" }

// IsPre indicates whether this effect should be applied before drawing.
// Drop shadows are post effects, so it always returns false.
func (e *DropShadowEffect) IsPre() bool { return false }

// Apply draws the drop shadow under the current image content.
//
// The shadow is composited below existing pixels based on their alpha,
// blurred, tinted, and offset by (x, y). The destination buffer (dst)
// is modified in place.
func (e *DropShadowEffect) Apply(dst *image.RGBA) {
	if e.opacity <= 0 {
		return
	}

	b := dst.Bounds()
	mask := extractAlpha(dst)

	// Expand alpha before blur (spread effect)
	if e.spread > 0 {
		mask = featherMask(mask, int(math.Round(e.spread)))
	}

	// Apply blur for smooth falloff
	if e.blur > 0 {
		mask = fastBlurAlpha(mask, int(math.Round(e.blur)))
	}

	const m = 1<<16 - 1
	r16, g16, b16, a16 := e.color.RGBA()
	cr := float64(r16) / m
	cg := float64(g16) / m
	cb := float64(b16) / m
	ca := float64(a16) / m * e.opacity

	// Build tinted shadow image with premultiplied RGB
	shadow := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			m := float64(mask.AlphaAt(x, y).A) / 255
			if m == 0 {
				continue
			}
			a := geom.ClampF64(m*ca, 0, 1)

			pr := uint8(math.Round(cr * a * 255))
			pg := uint8(math.Round(cg * a * 255))
			pb := uint8(math.Round(cb * a * 255))
			pa := uint8(math.Round(a * 255))

			i := shadow.PixOffset(x, y)
			shadow.Pix[i+0] = pr
			shadow.Pix[i+1] = pg
			shadow.Pix[i+2] = pb
			shadow.Pix[i+3] = pa
		}
	}

	offset := image.Point{X: int(math.Round(e.x)), Y: int(math.Round(e.y))}

	// Preserve original content
	srcCopy := image.NewRGBA(b)
	draw.Draw(srcCopy, b, dst, b.Min, draw.Src)

	// Clear destination and composite
	clearAlpha(dst)
	draw.Draw(dst, b, shadow, b.Min.Sub(offset), draw.Over)
	draw.Draw(dst, b, srcCopy, b.Min, draw.Over)
}

// clearAlpha sets all pixels in an RGBA image to transparent (zeroed).
func clearAlpha(img *image.RGBA) {
	for i := range img.Pix {
		img.Pix[i] = 0
	}
}

// fastBlurAlpha performs a separable box blur on the alpha channel of an image.
//
// The algorithm runs two passes (horizontal and vertical) over the alpha values,
// averaging neighboring pixels within ±radius. The result approximates a
// Gaussian blur efficiently without additional allocations.
func fastBlurAlpha(src *image.Alpha, radius int) *image.Alpha {
	if radius <= 0 {
		return src
	}
	b := src.Bounds()
	dst := image.NewAlpha(b)
	tmp := image.NewAlpha(b)

	w, h := b.Dx(), b.Dy()
	r := radius
	size := 2*r + 1

	// Horizontal blur pass
	for y := 0; y < h; y++ {
		sum := 0
		for x := -r; x <= r; x++ {
			xx := int(geom.ClampF64(float64(x), 0, float64(w-1)))
			sum += int(src.AlphaAt(b.Min.X+xx, b.Min.Y+y).A)
		}
		for x := 0; x < w; x++ {
			tmp.SetAlpha(b.Min.X+x, b.Min.Y+y, color.Alpha{A: uint8(sum / size)})
			left := int(geom.ClampF64(float64(x-r), 0, float64(w-1)))
			right := int(geom.ClampF64(float64(x+r+1), 0, float64(w-1)))
			sum += int(src.AlphaAt(b.Min.X+right, b.Min.Y+y).A)
			sum -= int(src.AlphaAt(b.Min.X+left, b.Min.Y+y).A)
		}
	}

	// Vertical blur pass
	for x := 0; x < w; x++ {
		sum := 0
		for y := -r; y <= r; y++ {
			yy := int(geom.ClampF64(float64(y), 0, float64(h-1)))
			sum += int(tmp.AlphaAt(b.Min.X+x, b.Min.Y+yy).A)
		}
		for y := 0; y < h; y++ {
			dst.SetAlpha(b.Min.X+x, b.Min.Y+y, color.Alpha{A: uint8(sum / size)})
			top := int(geom.ClampF64(float64(y-r), 0, float64(h-1)))
			bot := int(geom.ClampF64(float64(y+r+1), 0, float64(h-1)))
			sum += int(tmp.AlphaAt(b.Min.X+x, b.Min.Y+bot).A)
			sum -= int(tmp.AlphaAt(b.Min.X+x, b.Min.Y+top).A)
		}
	}
	return dst
}
