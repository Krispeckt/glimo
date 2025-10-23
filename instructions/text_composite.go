package instructions

import (
	"image"

	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

// compositePatternWithMask composites a pattern into the overlay image
// using the given alpha mask, blending mode, and opacity settings.
// It merges color data from the provided pattern with base pixels,
// producing premultiplied alpha results in the overlay buffer.
//
// Parameters:
//   - base: the original background image (read-only during blending).
//   - overlay: the output image buffer receiving the composited result.
//   - mask: alpha mask controlling transparency at each pixel (A channel only).
//   - dstX, dstY: top-left target coordinates where the mask is placed.
//   - clip: optional clipping rectangle; if non-empty, restricts compositing area.
//   - p: the pattern providing colors (solid, gradient, or blended).
//
// Behavior:
//  1. Skips processing if any input is nil or if the intersection area is empty.
//  2. Extracts blend mode and opacity from the pattern when available.
//  3. Converts opacity from [0–1] range to an 8-bit multiplier.
//  4. For each pixel where mask alpha > 0:
//     - Fetches pattern color at absolute canvas coordinates.
//     - Blends pattern color over base color according to blend mode.
//     - Writes the premultiplied result to the overlay buffer.
//  5. Supports both fully opaque and partially blended writes depending on opacity.
//
// This function performs per-pixel alpha blending and is a core primitive
// for stroking and filling text or vector shapes with pattern-based colors.
func compositePatternWithMask(base, overlay, mask *image.RGBA, dstX, dstY int, clip image.Rectangle, p patterns.Pattern) {
	if base == nil || overlay == nil || mask == nil || p == nil {
		return
	}

	place := image.Rect(dstX, dstY, dstX+mask.Bounds().Dx(), dstY+mask.Bounds().Dy()).
		Intersect(base.Bounds()).
		Intersect(overlay.Bounds())
	if !clip.Empty() {
		place = place.Intersect(clip)
	}
	if place.Empty() {
		return
	}

	mode := patterns.BlendNormal
	opacity := 1.0
	if bp, ok := p.(patterns.BlendedPattern); ok {
		mode, opacity = bp.BlendMode(), bp.Opacity()
	}
	if opacity <= 0 {
		return
	}
	if opacity > 1 {
		opacity = 1
	}
	t := uint8(opacity*255 + 0.5)
	omt := 255 - t

	xOff := place.Min.X - dstX
	yOff := place.Min.Y - dstY
	mmin := mask.Bounds().Min
	w, h := place.Dx(), place.Dy()

	for y := 0; y < h; y++ {
		bRow := base.PixOffset(place.Min.X, place.Min.Y+y)
		oRow := overlay.PixOffset(place.Min.X, place.Min.Y+y)
		mRow := mask.PixOffset(mmin.X+xOff, mmin.Y+yOff+y)

		for x := 0; x < w; x++ {
			ma := mask.Pix[mRow+3]
			if ma != 0 {
				ax, ay := place.Min.X+x, place.Min.Y+y

				col := p.ColorAt(ax, ay)
				src, ok := col.(patterns.Color)
				if !ok {
					src = patterns.NewColorFromStd(col)
				}
				if src.BlendMode() == patterns.BlendPassThrough {
					src = src.SetBlendMode(mode)
				}

				bg := patterns.Color{
					R: base.Pix[bRow+0],
					G: base.Pix[bRow+1],
					B: base.Pix[bRow+2],
					A: base.Pix[bRow+3],
				}

				out := src.BlendOver(bg, float64(ma)/255.0)

				// premultiplied components
				prp := geom.Mul255(out.R, out.A)
				pgp := geom.Mul255(out.G, out.A)
				pbp := geom.Mul255(out.B, out.A)

				if t == 255 {
					overlay.Pix[oRow+0] = prp
					overlay.Pix[oRow+1] = pgp
					overlay.Pix[oRow+2] = pbp
					overlay.Pix[oRow+3] = out.A
				} else {
					brp := geom.Mul255(bg.R, bg.A)
					bgp := geom.Mul255(bg.G, bg.A)
					bbp := geom.Mul255(bg.B, bg.A)

					overlay.Pix[oRow+0] = uint8((uint32(brp)*uint32(omt) + uint32(prp)*uint32(t) + 127) / 255)
					overlay.Pix[oRow+1] = uint8((uint32(bgp)*uint32(omt) + uint32(pgp)*uint32(t) + 127) / 255)
					overlay.Pix[oRow+2] = uint8((uint32(bbp)*uint32(omt) + uint32(pbp)*uint32(t) + 127) / 255)
					overlay.Pix[oRow+3] = uint8((uint32(bg.A)*uint32(omt) + uint32(out.A)*uint32(t) + 127) / 255)
				}
			}

			bRow += 4
			oRow += 4
			mRow += 4
		}
	}
}
