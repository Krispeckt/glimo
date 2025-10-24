// Package effects provides post- and pre-render image effects that can be attached
// to drawable instructions. The effect in this file simulates Figma-like texture
// overlay (procedural roughness) by tinting destination pixels with a noisy mask.
//
// High-level algorithm
//
//  1. For each destination pixel (x, y) compute a smooth pseudo-random value n
//     in [0, 1] using a sum of sine products (see mathSinNoise). The evaluation
//     domain is scaled by the user-provided `scale` so bigger values produce
//     wider, lower-frequency waves; smaller values produce finer grain.
//  2. Clamp and scale n by `intensity` in [0, 1]. Intensity = 0 disables the
//     effect; 1 uses the full amplitude of the procedural noise.
//  3. Linearly interpolate the destination RGB channels toward the effect color
//     using t = n * opacity, while leaving the alpha channel unchanged.
//     Interpolation is performed in 8-bit sRGB space for performance:
//     C_out = Lerp(C_dst, C_tex, t)
//  4. The effect runs as a post effect (IsPre() == false). It expects that
//     the base content has already been drawn into dst.
//
// Notes and constraints
//
//   - Image format: dst must be *image.RGBA. The buffer stores premultiplied
//     alpha, but this effect does not modify A and blends only RGB in-place.
//     This matches a simple "overlay tint" and is fast. If you need accurate
//     alpha-aware compositing, convert to linear space and back.
//   - Determinism: mathSinNoise includes calls to math/rand for slight phase
//     jitter. Using the package-level RNG makes results depend on the global
//     seed. If you need reproducibility across runs, seed math/rand yourself
//     (e.g., rand.Seed(0)) before applying the effect, or rewrite the noise
//     to be hash-based and seedable per-effect.
//   - Performance: The implementation iterates pixels and writes directly into
//     dst.Pix via PixOffset to avoid allocations. For very large images consider
//     tiling or parallelization.
//   - Thread-safety: Package-level functions in math/rand are safe for concurrent
//     use, but modifying the same *image.RGBA from multiple goroutines is not.
//     Guard dst with your own synchronization if you parallelize.
package effects

import (
	"image"

	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

// TextureEffect simulates a texture/roughness overlay similar to Figma's "Texture" fill.
//
// The effect tints destination pixels toward a given color using a smooth,
// procedural, sine-based noise field. The alpha channel is preserved.
//
// Parameters:
//   - scale: Spatial scale of the noise in pixels. Larger values produce
//     broader features. Clamped to [1, 100].
//   - intensity: Strength of the noise modulation in [0, 1]. Acts like an
//     amplitude for the noise field before opacity is applied.
//   - opacity: Final opacity of the tint in [0, 1]. Multiplies the noise value.
//   - color: Target tint color in sRGB 8-bit channels.
//   - blendMode: Reserved for future compositing modes; currently unused.
//
// Implementation detail:
// The type conforms to the project’s Effect interface via Name, IsPre, and Apply.
type TextureEffect struct {
	// scale controls the spatial frequency of the noise, measured in pixels per wave.
	// Higher values => smoother, larger blobs. Lower values => finer grain.
	scale float64
	// intensity scales the raw noise value before opacity. 0 disables; 1 uses full amplitude.
	intensity float64
	// opacity multiplies the per-pixel modulation produced by the noise, acting as
	// the final blend factor for RGB channel interpolation.
	opacity float64
	// color is the target tint that the destination RGB will be lerped toward.
	color patterns.Color
}

// NewTextureEffect constructs a new TextureEffect with clamped parameters.
//
// scale is clamped to [1, 100]. intensity is clamped to [0, 1]. opacity defaults to 1.
//
// Example:
//
//	e := effects.NewTextureEffect(24, 0.6, colors.Silver).SetOpacity(0.5)
//	shape.AddEffect(e)
//
// The effect is post-applied (IsPre() == false).
func NewTextureEffect(scale, intensity float64, c patterns.Color) *TextureEffect {
	return &TextureEffect{
		scale:     geom.ClampF64(scale, 1, 100),
		intensity: geom.ClampF64(intensity, 0, 1),
		color:     c,
		opacity:   1.0,
	}
}

// SetScale sets the spatial scale of the noise and clamps it to [1, 100].
//
// Larger values yield lower-frequency textures. Returns the receiver for chaining.
func (e *TextureEffect) SetScale(v float64) *TextureEffect {
	e.scale = geom.ClampF64(v, 1, 100)
	return e
}

// SetIntensity sets the strength of the noise field in [0, 1].
//
// intensity scales the noise before opacity is applied. Returns the receiver for chaining.
func (e *TextureEffect) SetIntensity(v float64) *TextureEffect {
	e.intensity = geom.ClampF64(v, 0, 1)
	return e
}

// SetOpacity sets the final blend opacity in [0, 1].
//
// The effective per-pixel blend factor becomes noise * opacity. Returns the receiver for chaining.
func (e *TextureEffect) SetOpacity(v float64) *TextureEffect {
	e.opacity = geom.ClampF64(v, 0, 1)
	return e
}

// Name returns the human-readable effect name.
// Satisfies the Effect interface.
func (e *TextureEffect) Name() string {
	return "Texture"
}

// IsPre reports whether the effect should run before the main draw pass.
//
// TextureEffect is a post effect, so this method returns false.
// Satisfies the Effect interface.
func (e *TextureEffect) IsPre() bool {
	return false
}

// Apply runs the texture overlay over the entire destination image.
//
// Preconditions:
//   - dst must be non-nil and of type *image.RGBA.
//   - The base content should already be composited into dst.
//
// Behavior:
//   - If opacity == 0, the method is a no-op.
//   - For each pixel, compute a smooth noise value in [0, 1], scale it by
//     intensity, then linearly interpolate each RGB channel toward e.color
//     with t = noise * opacity. Alpha is left unchanged.
//
// Complexity: O(W×H) time, O(1) extra space.
//
// Warning: The interpolation is done in sRGB byte space for speed and may not
// match physically correct results in linear RGB.
func (e *TextureEffect) Apply(dst *image.RGBA) {
	if e.opacity == 0 {
		return
	}
	b := dst.Bounds()
	w, h := b.Dx(), b.Dy()

	// Precompute reciprocals to avoid repeated divisions in the hot loop.
	invScale := 1.0 / e.scale

	for y := 0; y < h; y++ {
		yy := float64(y) * invScale
		for x := 0; x < w; x++ {
			xx := float64(x) * invScale

			// Raw noise in [-1, 1] -> remap to [0, 1].
			n := (geom.SinNoise(xx, yy) + 1) * 0.5
			// Apply intensity and clamp to [0, 1].
			n = geom.ClampF64(n*e.intensity, 0, 1)
			// Final blend factor per pixel.
			t := n * e.opacity

			i := dst.PixOffset(b.Min.X+x, b.Min.Y+y)

			// Lerp each RGB channel toward the texture color.
			dst.Pix[i+0] = uint8(geom.Lerp(float64(dst.Pix[i+0]), float64(e.color.R), t))
			dst.Pix[i+1] = uint8(geom.Lerp(float64(dst.Pix[i+1]), float64(e.color.G), t))
			dst.Pix[i+2] = uint8(geom.Lerp(float64(dst.Pix[i+2]), float64(e.color.B), t))
			// Alpha channel (i+3) remains unchanged by design.
		}
	}
}
