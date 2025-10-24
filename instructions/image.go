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

// FitMode tells HOW we resize the source into the target width/height.
// Think of it as "fitting strategy".
type FitMode int

const (
	// FitStretch : brutally stretch to target W×h. Aspect ratio is ignored.
	// Use this when you do not care about distortion.
	FitStretch FitMode = iota

	// FitContain : keep aspect ratio and make the whole image fit INSIDE W×h.
	// You may get empty space around (letterboxing/pillarboxing). No cropping.
	FitContain

	// FitCover : keep aspect ratio and make the image FILL W×h completely.
	// This may crop the edges. Good for full-bleed thumbnails and covers.
	FitCover
)

// Image draws a raster image with resize, flips, any-angle rotation,
// optional canvas expansion, external effects, and global opacity.
//
// Fields are intentionally unexported. Use getters/setters. This keeps the
// internal state consistent and easy to reason about for newcomers.
type Image struct {
	// src is the original input picture you want to draw.
	src image.Image

	mask *image.RGBA

	// x,y are the top-left position where the prepared layer will be placed
	// on the destination canvas.
	x, y int

	// w,h is the target size BEFORE flips/rotation. Zero means "use source".
	// Note: h is capitalized to avoid shadowing "h" in loops. It is intentional.
	w, h int

	// fit defines how we squeeze the source into w×h (see FitMode).
	fit FitMode

	// flipH/flipV mirror the image left-right / top-bottom.
	flipH, flipV bool

	// angleDeg is rotation in degrees. Positive means clockwise.
	// You can pass ANY float (e.g., 13.37). Internally we normalize to [0..360).
	angleDeg float64

	// expand=true means we enlarge the layer so the rotated image is not cropped.
	// expand=false keeps the layer size and crops rotated corners.
	expand bool

	// opacity is global alpha in [0..1]. 1=opaque, 0=fully transparent.
	opacity float64

	// bg is the fill color used for pixels that fall OUTSIDE the source
	// during rotation sampling. Usually Transparent.
	bg patterns.Color

	// effects is your external effect pipeline. It can modify
	//  - the destination (PRE) and
	//  - the prepared layer (POST).
	effects *containers.Effects
}

// NewImage creates a new Image instruction located at (x, y) with safe defaults.
// Defaults are chosen to "just work":
//   - FitContain: no distortion.
//   - opacity=1: fully visible.
//   - bg=Transparent: clean rotation outside fill.
//   - empty effects chain: nothing runs unless you add effects.
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

// SetSize sets the target width/height in pixels.
// If you pass 0 for w or h, that specific dimension will use the source value.
// Example: SetSize(0, 300) keeps source width, scales height to 300.
func (im *Image) SetSize(w, h int) *Image { im.w, im.h = w, h; return im }

// SetFit selects the resize strategy (Stretch/Contain/Cover). See FitMode docs.
func (im *Image) SetFit(f FitMode) *Image { im.fit = f; return im }

// Mirror flips the image. h=true mirrors left↔right. v=true mirrors top↔bottom.
// You can combine both.
func (im *Image) Mirror(h, v bool) *Image { im.flipH, im.flipV = h, v; return im }

// SetFlip is the same as Mirror, provided for naming preference.
func (im *Image) SetFlip(h, v bool) *Image { im.flipH, im.flipV = h, v; return im }

// Rotate sets rotation in degrees. Any float is valid. Positive = clockwise.
func (im *Image) Rotate(deg float64) *Image { im.angleDeg = deg; return im }

// SetExpand controls whether the layer should grow to fully contain the rotated image.
// true = no crop at corners; false = keep original layer size and crop corners.
func (im *Image) SetExpand(b bool) *Image { im.expand = b; return im }

// SetOpacity sets global transparency in [0..1]. Values outside are clamped.
// 1 = fully visible. 0.5 = 50% see-through. 0 = invisible.
func (im *Image) SetOpacity(o float64) *Image {
	im.opacity = geom.ClampF64(o, 0, 1)
	return im
}

// SetBackground sets the fallback color used by rotation sampling outside source bounds.
// For most cases you want colors.Transparent here.
func (im *Image) SetBackground(c patterns.Color) *Image { im.bg = c; return im }

func (im *Image) SetMaskImage(m *image.RGBA) *Image {
	im.mask = m
	return im
}

func (im *Image) SetMaskFromShape(s Shape) *Image {
	if s == nil {
		return im
	}

	size := image.NewRGBA(image.Rect(0, 0, im.w, im.h))
	if im.w == 0 || im.h == 0 {
		b := im.src.Bounds()
		size = image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	}

	s.Draw(size, size)

	im.mask = size
	return im
}

// ClearMask сбрасывает маску.
func (im *Image) ClearMask() *Image { im.mask = nil; return im }

// AddEffect attaches a visual effect to the text rendering pipeline.
func (im *Image) AddEffect(e effects.Effect) *Image {
	im.effects.Add(e)
	return im
}

// AddEffects attaches multiple visual effects to the pipeline.
func (im *Image) AddEffects(es ...effects.Effect) *Image {
	im.effects.AddList(es)
	return im
}

// SetPosition moves the layer to (x, y) on the destination canvas.
func (im *Image) SetPosition(x, y int) { im.x, im.y = x, y }

// Position returns the top-left coordinate where the layer is drawn.
func (im *Image) Position() (int, int) { return im.x, im.y }

// Size returns the target size. Zero value means "use source" for that axis.
func (im *Image) Size() *geom.Size { return geom.NewSize(float64(im.w), float64(im.h)) }

// Draw executes the full pipeline and paints onto overlay.
func (im *Image) Draw(_, overlay *image.RGBA) {
	if im.src == nil || im.opacity <= 0 {
		return
	}

	im.effects.PreApplyAll(overlay)

	img := im.src
	W, H := im.targetSize()
	if W > 0 && H > 0 {
		img = resizeWithFit(img, W, H, im.fit)
	}
	imgLayer := imageUtil.ToRGBA(img)

	if im.flipH || im.flipV {
		imgLayer = flipRGBA(imgLayer, im.flipH, im.flipV)
	}
	if a := math.Mod(im.angleDeg, 360); a != 0 {
		imgLayer = rotateAnyRGBA(imgLayer, a, im.bg, im.expand)
	}

	im.effects.PostApplyAll(imgLayer)

	dstPt := image.Pt(im.x, im.y)
	dstRect := image.Rectangle{Min: dstPt, Max: dstPt.Add(imgLayer.Bounds().Size())}

	// Клип по overlay.
	place := dstRect.Intersect(overlay.Bounds())
	if place.Empty() {
		return
	}
	// Смещение источника относительно обрезанного прямоугольника.
	srcPt := imgLayer.Bounds().Min.Add(place.Min.Sub(dstRect.Min))

	switch {
	case im.mask == nil && im.opacity >= 1:
		draw.Draw(overlay, place, imgLayer, srcPt, draw.Over)

	case im.mask != nil && im.opacity >= 1:
		// Маска читается в абсолютных координатах; клип — через place.
		draw.DrawMask(overlay, place, imgLayer, srcPt, im.mask, place.Min, draw.Over)

	default:
		alpha := uint8(geom.ClampF64(im.opacity*255, 0, 255))
		loc := multiplyMaskAlphaClipped(im.mask, place, alpha)
		draw.DrawMask(overlay, place, imgLayer, srcPt, loc, image.Point{}, draw.Over)
	}
}

// multiplyMaskAlphaClipped создаёт локальную маску размера place.
// A = mask.A(at absolute (x,y)) * globalAlpha / 255.
func multiplyMaskAlphaClipped(mask *image.RGBA, place image.Rectangle, globalAlpha uint8) *image.Alpha {
	out := image.NewAlpha(image.Rect(0, 0, place.Dx(), place.Dy()))
	if mask == nil || globalAlpha == 0 || place.Empty() {
		return out
	}
	for y := 0; y < place.Dy(); y++ {
		for x := 0; x < place.Dx(); x++ {
			ax := place.Min.X + x
			ay := place.Min.Y + y
			ma := mask.RGBAAt(ax, ay).A // вне bounds вернёт 0
			out.SetAlpha(x, y, color.Alpha{A: geom.Mul255(ma, globalAlpha)})
		}
	}
	return out
}

// targetSize computes the final size used for resizing.
// If either dimension is zero, we pull it from the source bounds.
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

// resizeWithFit resizes the image according to the selected FitMode.
// Result is either resized directly (Stretch/Contain) or resized+cropped (Cover).
func resizeWithFit(src image.Image, W, H int, mode FitMode) image.Image {
	switch mode {
	case FitStretch:
		// Directly resize to W×h, ignoring aspect ratio.
		return imageUtil.ResizeRGBA(src, W, H)

	case FitContain:
		// Keep aspect ratio. fit entire image inside W×h.
		sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
		if sw == 0 || sh == 0 {
			// Degenerate source; fall back to requested size.
			return imageUtil.ResizeRGBA(src, W, H)
		}
		r := math.Min(float64(W)/float64(sw), float64(H)/float64(sh))
		return imageUtil.ResizeRGBA(src,
			int(math.Round(float64(sw)*r)),
			int(math.Round(float64(sh)*r)),
		)

	case FitCover:
		// Keep aspect ratio. Fill W×h entirely. Crop the overflow at center.
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
		// Unknown mode: fall back to a simple resize.
		return imageUtil.ResizeRGBA(src, W, H)
	}
}

// flipRGBA returns a NEW image that is flipped horizontally and/or vertically.
// If neither flag is true, the original image is returned as is.
func flipRGBA(src *image.RGBA, hflip, vflip bool) *image.RGBA {
	if !hflip && !vflip {
		return src
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(b)

	// Simple index mapping. For each destination pixel we pull the corresponding
	// source pixel. This is easy to read and understand for beginners.
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

// rotateAnyRGBA rotates the image by ANY angle in DEGREES using bilinear sampling.
//   - Bilinear sampling mixes neighbor pixels, so rotated images look smooth.
//   - If expand=true: the output canvas becomes large enough to hold the full
//     rotated rectangle (no cropping).
//   - If expand=false: output keeps original size, so corners may be cut.
//   - Pixels that map outside the source bounds get the background color `bg`.
func rotateAnyRGBA(src *image.RGBA, angleDeg float64, bg patterns.Color, expand bool) *image.RGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw == 0 || sh == 0 {
		// Empty source: nothing to rotate.
		return src
	}

	// Convert degrees to radians and precompute sine/cosine.
	rad := geom.Deg2Rad(angleDeg)
	sinA, cosA := math.Sincos(rad)

	// Decide destination size depending on expand.
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

	// For every pixel in the DESTINATION, find where it came from in the SOURCE
	// using the inverse rotation. Then sample with bilinear interpolation.
	for y := 0; y < dh; y++ {
		fy := float64(y) - dcy
		for x := 0; x < dw; x++ {
			fx := float64(x) - dcx

			// Inverse mapping by −angle:
			//   [sx] = [ cosA  sinA] [fx] + [scx]
			//   [sy]   [-sinA  cosA] [fy]   [scy]
			sx := +fx*cosA + fy*sinA + scx
			sy := -fx*sinA + fy*cosA + scy

			// If the source coordinate is outside, fill with background.
			if sx < 0 || sx > float64(sw-1) || sy < 0 || sy > float64(sh-1) {
				dst.SetRGBA(x, y, bgRGBA)
				continue
			}

			// Otherwise, fetch a smooth color using bilinear sampling.
			// NOTE: we pass bg too, but here coords are inside so it is not used.
			c := geom.BilinearRGBAAt(src, sx, sy, bg)
			dst.SetRGBA(x, y, c)
		}
	}
	return dst
}
