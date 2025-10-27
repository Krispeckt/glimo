// Package instructions provides primitives for drawing raster images with fitting,
// flipping, rotation, mask support, effects, and global opacity.
package instructions

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/effects"
	"github.com/Krispeckt/glimo/internal/containers"
	"github.com/Krispeckt/glimo/internal/core/geom"
	imageUtil "github.com/Krispeckt/glimo/internal/core/image"
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

// FitMode defines how the source image is resized to the target width/height.
type FitMode int

const (
	// FitStretch stretches to exactly W×H. Aspect ratio is ignored.
	FitStretch FitMode = iota

	// FitContain preserves aspect ratio and fits fully inside W×H.
	// May leave empty space (letterbox/pillarbox). No cropping.
	FitContain

	// FitCover preserves aspect ratio and fills W×H completely.
	// Crops overflow. Good for thumbnails and covers.
	FitCover
)

// Image draws a raster with resize, flips, any-angle rotation,
// optional canvas expansion, effects, global opacity, and optional mask.
//
// Fields are intentionally unexported. Use setters to keep state consistent.
type Image struct {
	// src is the original input image to draw.
	src image.Image

	// mask is an optional per-pixel alpha mask in destination space.
	mask *image.RGBA

	// x,y are the destination top-left where the prepared layer is placed.
	x, y int

	// w,h are target dimensions before flips/rotation. Zero means use source.
	// Note: capital H avoids shadowing loop variables.
	w, h int

	// fit selects the resize policy.
	fit FitMode

	// flipH/flipV mirror horizontally / vertically.
	flipH, flipV bool

	// angleDeg is rotation in degrees. Positive is clockwise.
	angleDeg float64

	// expand grows the layer to avoid rotation cropping when true.
	expand bool

	// opacity is global alpha in [0..1].
	opacity float64

	// bg is the color used when rotation samples outside the source.
	bg patterns.Color

	// effects is an external pipeline that can run pre/post.
	effects *containers.Effects
}

// NewImage creates a new Image at (x, y) with safe defaults:
//   - FitContain
//   - opacity = 1
//   - bg = Transparent
//   - empty effects chain
func NewImage(src image.Image, x, y int) *Image {
	return &Image{
		x:       x,
		y:       y,
		src:     src,
		fit:     FitContain,
		opacity: 1,
		bg:      colors.Transparent,
		effects: &containers.Effects{},
	}
}

// SetSize sets target width/height. Zero keeps that axis from the source.
func (im *Image) SetSize(w, h int) *Image { im.w, im.h = w, h; return im }

// SetFit selects Stretch/Contain/Cover.
func (im *Image) SetFit(f FitMode) *Image { im.fit = f; return im }

// Mirror flips the image. h for horizontal, v for vertical.
func (im *Image) Mirror(h, v bool) *Image { im.flipH, im.flipV = h, v; return im }

// SetFlip is an alias of Mirror.
func (im *Image) SetFlip(h, v bool) *Image { im.flipH, im.flipV = h, v; return im }

// Rotate sets rotation angle in degrees.
func (im *Image) Rotate(deg float64) *Image { im.angleDeg = deg; return im }

// SetExpand controls whether rotation expands the canvas to avoid cropping.
func (im *Image) SetExpand(b bool) *Image { im.expand = b; return im }

// SetOpacity sets global alpha in [0..1]. Values are clamped.
func (im *Image) SetOpacity(o float64) *Image {
	im.opacity = geom.ClampF64(o, 0, 1)
	return im
}

// SetBackground sets the color sampled outside source bounds during rotation.
func (im *Image) SetBackground(c patterns.Color) *Image { im.bg = c; return im }

// SetMaskImage assigns a mask image in destination space.
func (im *Image) SetMaskImage(m *image.RGBA) *Image {
	im.mask = m
	return im
}

// SetMaskFromShape renders a Shape into an RGBA mask matching the target size.
// If target size is zero, uses source bounds.
func (im *Image) SetMaskFromShape(s Shape) *Image {
	if s == nil {
		return im
	}

	size := image.NewRGBA(image.Rect(0, 0, im.w, im.h))
	if im.w == 0 || im.h == 0 {
		b := im.src.Bounds()
		size = image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	}

	// The shape draws its alpha into the mask canvas.
	s.Draw(size, size)

	im.mask = size
	return im
}

// ClearMask removes the current mask.
func (im *Image) ClearMask() *Image { im.mask = nil; return im }

// AddEffect appends a single effect to the pipeline.
func (im *Image) AddEffect(e effects.Effect) *Image {
	im.effects.Add(e)
	return im
}

// AddEffects appends multiple effects.
func (im *Image) AddEffects(es ...effects.Effect) *Image {
	im.effects.AddList(es)
	return im
}

// SetPosition moves the layer to (x, y).
func (im *Image) SetPosition(x, y int) { im.x, im.y = x, y }

// Position returns the destination top-left coordinate.
func (im *Image) Position() (int, int) { return im.x, im.y }

// Size returns the target size. Zero values mean "use source" for that axis.
func (im *Image) Size() *geom.Size { return geom.NewSize(float64(im.w), float64(im.h)) }

// Draw runs the pipeline and composites onto overlay.
func (im *Image) Draw(_, overlay *image.RGBA) {
	if im.src == nil || im.opacity <= 0 {
		return
	}

	im.effects.PreApplyAll(overlay)

	// 1) Resize according to FitMode.
	img := im.src
	W, H := im.targetSize()
	if W > 0 && H > 0 {
		img = resizeWithFit(img, W, H, im.fit)
	}
	imgLayer := imageUtil.ToRGBA(img)

	// 2) Flips and rotation.
	if im.flipH || im.flipV {
		imgLayer = flipRGBA(imgLayer, im.flipH, im.flipV)
	}
	if a := math.Mod(im.angleDeg, 360); a != 0 {
		imgLayer = rotateAnyRGBA(imgLayer, a, im.bg, im.expand)
	}

	// 3) Post effects operate on the prepared layer.
	im.effects.PostApplyAll(imgLayer)

	// 4) Compute destination placement and clip region.
	dstPt := image.Pt(im.x, im.y)
	dstRect := image.Rectangle{Min: dstPt, Max: dstPt.Add(imgLayer.Bounds().Size())}

	place := dstRect.Intersect(overlay.Bounds())
	if place.Empty() {
		return
	}

	// Source origin adjusted by the amount clipped.
	srcPt := imgLayer.Bounds().Min.Add(place.Min.Sub(dstRect.Min))

	// 5) Align mask to follow the image. Mask origin is tied to dstRect.Min.
	maskOffset := image.Pt(0, 0)
	if im.mask != nil {
		// mp in DrawMask corresponds to place.Min. To move mask with the image,
		// set mp = place.Min - dstRect.Min.
		maskOffset = place.Min.Sub(dstRect.Min)
	}

	// 6) Composite with or without mask and global opacity.
	switch {
	case im.mask == nil && im.opacity >= 1:
		draw.Draw(overlay, place, imgLayer, srcPt, draw.Over)

	case im.mask != nil && im.opacity >= 1:
		draw.DrawMask(overlay, place, imgLayer, srcPt, im.mask, maskOffset, draw.Over)

	default:
		alpha := uint8(geom.ClampF64(im.opacity*255, 0, 255))
		loc := multiplyMaskAlphaClippedOffset(im.mask, place, maskOffset, alpha)
		draw.DrawMask(overlay, place, imgLayer, srcPt, loc, image.Point{}, draw.Over)
	}
}

// multiplyMaskAlphaClippedOffset creates a local Alpha mask sized to `place`.
// Each pixel is: mask.A at (mp + local) multiplied by globalAlpha / 255.
func multiplyMaskAlphaClippedOffset(mask *image.RGBA, place image.Rectangle, maskOffset image.Point, globalAlpha uint8) *image.Alpha {
	out := image.NewAlpha(image.Rect(0, 0, place.Dx(), place.Dy()))
	if mask == nil || globalAlpha == 0 || place.Empty() {
		return out
	}

	mb := mask.Bounds()
	baseX := mb.Min.X + maskOffset.X
	baseY := mb.Min.Y + maskOffset.Y

	for y := 0; y < place.Dy(); y++ {
		for x := 0; x < place.Dx(); x++ {
			mx := baseX + x
			my := baseY + y
			ma := uint8(0)
			if mx >= mb.Min.X && mx < mb.Max.X && my >= mb.Min.Y && my < mb.Max.Y {
				ma = mask.RGBAAt(mx, my).A
			}
			out.SetAlpha(x, y, color.Alpha{A: geom.Mul255(ma, globalAlpha)})
		}
	}
	return out
}

// targetSize returns the final resize dimensions, substituting source
// dimensions for any axis that is zero.
func (im *Image) targetSize() (int, int) {
	w, h := im.w, im.h
	if w <= 0 || h <= 0 {
		sb := im.src.Bounds()
		if w <= 0 {
			w = sb.Dx()
		}
		if h <= 0 {
			h = sb.Dy()
		}
	}
	return w, h
}

// resizeWithFit applies the selected FitMode.
// Stretch: direct resize. Contain: aspect-fit. Cover: aspect-fill + center crop.
func resizeWithFit(src image.Image, W, H int, mode FitMode) image.Image {
	switch mode {
	case FitStretch:
		return imageUtil.ResizeRGBA(src, W, H)

	case FitContain:
		sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
		if sw == 0 || sh == 0 {
			return imageUtil.ResizeRGBA(src, W, H)
		}
		r := math.Min(float64(W)/float64(sw), float64(H)/float64(sh))
		return imageUtil.ResizeRGBA(src,
			int(math.Round(float64(sw)*r)),
			int(math.Round(float64(sh)*r)),
		)

	case FitCover:
		sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
		if sw == 0 || sh == 0 {
			return imageUtil.ResizeRGBA(src, W, H)
		}
		r := math.Max(float64(W)/float64(sw), float64(H)/float64(sh))
		tw := int(math.Ceil(float64(sw) * r))
		th := int(math.Ceil(float64(sh) * r))

		scaled := imageUtil.ResizeRGBA(src, tw, th)
		cx := (tw - W) / 2
		cy := (th - H) / 2
		return imageUtil.CropRGBA(scaled, image.Rect(cx, cy, cx+W, cy+H))

	default:
		return imageUtil.ResizeRGBA(src, W, H)
	}
}

// flipRGBA returns a new image flipped horizontally and/or vertically.
// If both flags are false, returns the original reference.
func flipRGBA(src *image.RGBA, hflip, vflip bool) *image.RGBA {
	if !hflip && !vflip {
		return src
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(b)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			tx, ty := x, y
			if hflip {
				tx = w - 1 - x
			}
			if vflip {
				ty = h - 1 - y
			}
			dst.Set(tx+b.Min.X, ty+b.Min.Y, src.At(x+b.Min.X, y+b.Min.Y))
		}
	}
	return dst
}

// rotateAnyRGBA rotates by an arbitrary angle (degrees) using bilinear sampling.
// - If expand=true, output bounds are enlarged to fit the rotated rect.
// - If expand=false, output size equals input size and corners may be cropped.
// - Pixels mapped outside the source use bg.
func rotateAnyRGBA(src *image.RGBA, angleDeg float64, bg patterns.Color, expand bool) *image.RGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw == 0 || sh == 0 {
		return src
	}

	rad := geom.Deg2Rad(angleDeg)
	sinA, cosA := math.Sincos(rad)

	dw, dh := sw, sh
	if expand {
		dw, dh = geom.RotatedBounds(sw, sh, rad)
	}
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))

	scx := float64(sw-1) / 2
	scy := float64(sh-1) / 2
	dcx := float64(dw-1) / 2
	dcy := float64(dh-1) / 2

	bgRGBA := bg.ToColor()

	// Inverse mapping for each destination pixel followed by bilinear sampling.
	for y := 0; y < dh; y++ {
		fy := float64(y) - dcy
		for x := 0; x < dw; x++ {
			fx := float64(x) - dcx

			// Inverse rotation by -angle.
			// [sx] = [ cosA  sinA] [fx] + [scx]
			// [sy]   [-sinA  cosA] [fy]   [scy]
			sx := +fx*cosA + fy*sinA + scx
			sy := -fx*sinA + fy*cosA + scy

			if sx < 0 || sx > float64(sw-1) || sy < 0 || sy > float64(sh-1) {
				dst.SetRGBA(x, y, bgRGBA)
				continue
			}

			c := geom.BilinearRGBAAt(src, sx, sy, bg)
			dst.SetRGBA(x, y, c)
		}
	}
	return dst
}
