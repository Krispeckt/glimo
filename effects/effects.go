// Package effects defines interfaces and utilities for 2D image effects that can
// be applied to rendered layers before or after drawing operations. These effects
// modify *image.RGBA buffers directly to simulate phenomena like shadows, blur,
// glow, or color overlays.
//
// This file contains the base Effect interface shared by all post-processing
// filters, as well as helper functions for working with alpha masks such as
// extracting or softening alpha channels.
//
// Overview
//
//   - The Effect interface provides a unified contract for any visual filter
//     that modifies an RGBA image, including its name and processing phase.
//   - The helper functions (`extractAlpha`, `featherMask`) operate on alpha
//     masks to facilitate effects that depend on object silhouettes or smooth
//     transparency gradients.
//
// Notes:
//   - Effects may be either pre-render (IsPre == true) or post-render (IsPre == false).
//     Pre effects modify the layer before content is drawn (e.g., tint, blur base),
//     while post effects process the result afterward (e.g., shadow, glow).
//   - All effects work in-place on the provided image buffer and must ensure
//     bounds safety themselves.
//   - Helper functions are CPU-based and not optimized for real-time pipelines,
//     but they are suitable for procedural rendering or offline composition.
package effects

import (
	"image"
	"image/color"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// Effect defines the universal interface for any image-based visual effect.
//
// An Effect operates directly on an *image.RGBA buffer, optionally modifying
// pixels in-place to produce a visual transformation such as blur, shadow,
// glow, or color grading.
type Effect interface {
	// Apply performs the effect’s operation on the provided RGBA image.
	// The modification is in-place; implementations must handle bounds safety.
	Apply(dst *image.RGBA)

	// Name returns a short, human-readable identifier for this effect,
	// typically used for debugging or logging.
	Name() string

	// IsPre reports whether this effect should be applied before the layer’s
	// content is drawn. Pre-effects modify the base buffer; post-effects apply
	// after the main drawing pass.
	IsPre() bool
}

// featherMask softens or expands an alpha mask by averaging neighboring pixels.
//
// This function applies a simple box blur over the alpha channel of the source
// image (`src`) with the given radius. The result is a smoother or "feathered"
// mask that can be used to produce soft edges for shadows, glows, or fades.
//
// Parameters:
//   - src: input alpha image.
//   - radius: number of pixels around each sample to average. Values <= 0 disable the effect.
//
// Returns a new *image.Alpha containing the softened alpha mask.
func featherMask(src *image.Alpha, radius int) *image.Alpha {
	if radius <= 0 {
		return src
	}
	b := src.Bounds()
	dst := image.NewAlpha(b)
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var sum, count int
			for ky := -radius; ky <= radius; ky++ {
				yy := geom.ClampF64(float64(y+ky), 0, float64(h-1))
				for kx := -radius; kx <= radius; kx++ {
					xx := geom.ClampF64(float64(x+kx), 0, float64(w-1))
					sum += int(src.AlphaAt(b.Min.X+int(xx), b.Min.Y+int(yy)).A)
					count++
				}
			}
			dst.SetAlpha(b.Min.X+x, b.Min.Y+y, color.Alpha{A: uint8(sum / count)})
		}
	}
	return dst
}

// extractAlpha extracts the alpha channel from an RGBA image and returns it as
// a standalone *image.Alpha mask.
//
// This is commonly used by effects such as shadows or glows to identify the
// visible shape of an object based on its transparency.
//
// Parameters:
//   - src: input *image.RGBA image.
//
// Returns:
//   - A new *image.Alpha containing the alpha channel values from `src`.
//
// The resulting mask shares the same bounds as the source image.
func extractAlpha(src *image.RGBA) *image.Alpha {
	b := src.Bounds()
	mask := image.NewAlpha(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			i := src.PixOffset(x, y)
			mask.Pix[(y-b.Min.Y)*b.Dx()+(x-b.Min.X)] = src.Pix[i+3]
		}
	}
	return mask
}
