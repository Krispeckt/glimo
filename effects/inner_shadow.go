// Package effects provides post- and pre-render image effects that can be attached
// to drawable instructions and compositing stages. These effects modify pixel data
// to simulate optical phenomena such as shadows, lighting, texture overlays,
// or atmospheric depth. All operations are performed directly on CPU-based
// *image.RGBA buffers for deterministic, hardware-independent rendering.
//
// The effect implemented in this file simulates a CSS-like *inset box-shadow*
// (inner shadow). It darkens the interior edges of a rendered shape according
// to its alpha mask, creating the perception of depth or recessed surfaces.
//
// High-level algorithm
//
//  1. Extract the alpha channel of the destination image (dst). The alpha defines
//     the shape mask to which the inner shadow will be applied.
//  2. Invert the alpha mask so that the area outside the shape becomes opaque.
//     This inverted mask defines the region where the shadow will originate.
//  3. Offset the inverted mask in the opposite direction of the desired shadow
//     (negative offset) to simulate a light source casting from the given
//     direction. For example, positive offsetX and offsetY produce a lower-right
//     inner shadow.
//  4. Apply a box blur of radius `blur` to soften the shadow edges. The blur
//     kernel is separable and applied in-place for efficiency.
//  5. Multiply the blurred mask by the original alpha, ensuring the shadow
//     only affects pixels *inside* the shape’s visible area.
//  6. Tint destination RGB channels toward the shadow color with opacity-scaled
//     interpolation. The alpha channel remains unmodified to match CSS behavior.
//
// Notes and constraints
//
//   - Image format: dst must be *image.RGBA with premultiplied alpha. The effect
//     reads and writes pixel data directly through Pix/PixOffset. Alpha blending
//     is simulated via per-channel linear interpolation in sRGB space.
//   - Opacity: final opacity equals `effect.opacity * color.A/255`. Setting either
//     to 0 disables the shadow. Opacity values are clamped to [0, 1].
//   - Coordinate system: offsets use CSS-style semantics. Positive offsetX/offsetY
//     move the *light source* down and right, which pushes the shadow upward/left.
//   - Performance: the implementation is linear in the number of pixels. For large
//     images, consider parallelization or reusing intermediate buffers.
//   - Thread-safety: concurrent modification of the same *image.RGBA is unsafe.
//     Guard dst if the effect is applied concurrently across tiles or regions.
//
// Example usage:
//
//	shadow := effects.NewInnerShadowEffect(4, 4, 6, patterns.Color{R: 0, G: 0, B: 0, A: 128})
//	shadow.SetOpacity(0.8)
//	shadow.Apply(target)
//
// The resulting image will appear as if the drawn shape has a soft, recessed
// inner shadow consistent with Figma or CSS `box-shadow: inset ...` behavior.
package effects

import (
	"image"
	"math"

	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
	"golang.org/x/image/draw"
)

// InnerShadowEffect simulates a CSS-like "inset box-shadow" effect.
// It darkens the interior edges of a shape based on its alpha channel,
// producing the illusion of depth or an inner cavity.
//
// Parameters:
//   - offsetX / offsetY: shadow displacement in pixels (positive = right/bottom).
//   - blur: blur radius in pixels for soft edges.
//   - color: shadow tint color.
//   - opacity: overall transparency factor (0..1), multiplied by the color alpha.
type InnerShadowEffect struct {
	offsetX float64
	offsetY float64
	blur    float64
	color   patterns.Color
	opacity float64 // 0..1, multiplied by the color alpha
}

// NewInnerShadowEffect constructs a new inner shadow effect using the given
// offset, blur, and color. The initial opacity is derived from the color's alpha.
func NewInnerShadowEffect(offsetX, offsetY, blur float64, c patterns.Color) *InnerShadowEffect {
	return &InnerShadowEffect{
		offsetX: offsetX,
		offsetY: offsetY,
		blur:    blur,
		color:   c,
		opacity: float64(c.A) / 255.0,
	}
}

// SetOpacity sets a custom opacity factor (0..1) and returns the same effect
// for chaining. The final opacity is the product of this value and the color alpha.
func (e *InnerShadowEffect) SetOpacity(v float64) *InnerShadowEffect {
	e.opacity = geom.ClampF64(v, 0, 1)
	return e
}

// Name returns a human-readable identifier for this effect.
func (e *InnerShadowEffect) Name() string { return "InnerShadow" }

// IsPre reports whether this effect should be applied before drawing content.
// Inner shadows are post-render effects, so this always returns false.
func (e *InnerShadowEffect) IsPre() bool { return false }

// Apply renders the inner shadow into the destination RGBA image (dst).
//
// The effect is achieved in four stages:
//  1. Invert the alpha channel to create an outer mask.
//  2. Shift the mask in the opposite direction of the offset
//     (positive offset moves the light source down/right).
//  3. Blur the shifted mask for smooth transitions.
//  4. Multiply the blurred mask by the original alpha to limit the shadow
//     to inside the shape, tint it by the shadow color, and blend it back.
//
// The resulting appearance mimics CSS `box-shadow: inset ...`.
func (e *InnerShadowEffect) Apply(dst *image.RGBA) {
	if e.opacity <= 0 {
		return
	}

	b := dst.Bounds()

	// Step 1: Invert alpha — regions outside the shape become opaque.
	inv := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			i := dst.PixOffset(x, y)
			a := dst.Pix[i+3]
			inv.Pix[i+3] = 255 - a
		}
	}

	// Step 2: Offset mask (negative of the light offset).
	shiftX := -int(math.Round(e.offsetX))
	shiftY := -int(math.Round(e.offsetY))
	shifted := image.NewRGBA(b)
	draw.Copy(shifted, image.Point{X: b.Min.X + shiftX, Y: b.Min.Y + shiftY}, inv, b, draw.Src, nil)

	// Step 3: Apply blur.
	r := int(math.Round(e.blur))
	if r > 0 {
		boxBlur(shifted, shifted, r)
	}

	// Step 4: Mask intersection and color tinting.
	cr := float64(e.color.R)
	cg := float64(e.color.G)
	cb := float64(e.color.B)
	op := e.opacity

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			i := dst.PixOffset(x, y)
			sa := float64(shifted.Pix[i+3]) / 255.0 // shadow mask
			aa := float64(dst.Pix[i+3]) / 255.0     // original alpha
			a := sa * aa * op                       // final blend factor
			if a <= 0 {
				continue
			}
			// Blend toward shadow color without changing alpha.
			dst.Pix[i+0] = uint8(geom.Lerp(float64(dst.Pix[i+0]), cr, a))
			dst.Pix[i+1] = uint8(geom.Lerp(float64(dst.Pix[i+1]), cg, a))
			dst.Pix[i+2] = uint8(geom.Lerp(float64(dst.Pix[i+2]), cb, a))
		}
	}
}
