// Package effects implements pre- and post-render effects for 2D drawing instructions.
// This file defines a noise overlay effect similar to Figma’s “Noise” fill.
// It modifies RGB values of the destination image using random sampling and blending.
//
// Algorithm summary
//
//  1. Iterate over all pixels of the target image.
//  2. For each pixel, with probability equal to `density`, apply a random noise color
//     according to the selected noise type (mono, duo, or multi).
//  3. Blend the noise color into the pixel’s RGB channels using linear interpolation
//     by `opacity`.
//
// The alpha channel remains unchanged. The effect is post-applied (IsPre() == false).
//
// Parameters:
//   - noiseType — type of noise pattern (mono/duo/multi).
//   - density — probability of applying noise at each pixel in [0, 1].
//   - opacity — blend factor in [0, 1].
//   - colorA, colorB — base colors for duo mode.
//   - blendMode — reserved for future blending options.
//
// Notes:
//   - Non-deterministic: Uses math/rand global RNG. Seed externally if deterministic output is needed.
//   - Complexity: O(W×H) with very low per-pixel cost.
//   - Safe for any *image.RGBA buffer. No allocations.
package effects

import (
	"image"
	"math/rand"

	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

// NoiseType defines the noise generation mode.
// It determines how pixel colors are chosen for the noise pattern.
type NoiseType int

const (
	// NoiseMono — uniform grayscale noise, each selected pixel takes a random
	// brightness value from 0 to 255, applied equally to R, G, and B channels.
	NoiseMono NoiseType = iota
	// NoiseDuo — binary color noise. Each pixel randomly picks between colorA and colorB.
	NoiseDuo
	// NoiseMulti — full RGB noise, each pixel gets a completely random color.
	NoiseMulti
)

// NoiseEffect represents a noise overlay applied as a post-processing step.
// The noise modifies RGB channels based on the selected NoiseType.
type NoiseEffect struct {
	noiseType NoiseType      // Noise generation mode (Mono, Duo, Multi)
	density   float64        // Fraction of pixels affected, range [0,1]
	opacity   float64        // Blend strength of noise, range [0,1]
	colorA    patterns.Color // Primary color for duo noise
	colorB    patterns.Color // Secondary color for duo noise
}

// NewNoiseEffect creates a new NoiseEffect with the given type and density.
// Density defines how frequently noise is applied to pixels (0=no noise, 1=full coverage).
// Opacity defaults to 1.0; colors default to white and black.
//
// Example:
//
//	e := effects.NewNoiseEffect(effects.NoiseDuo, 0.3).
//		SetColors(colors.White, colors.Black).
//		SetOpacity(0.5)
//
//	shape.AddEffect(e)
func NewNoiseEffect(nt NoiseType, density float64) *NoiseEffect {
	return &NoiseEffect{
		noiseType: nt,
		density:   geom.ClampF64(density, 0, 1),
		opacity:   1.0,
		colorA:    patterns.Color{R: 255, G: 255, B: 255, A: 255},
		colorB:    patterns.Color{A: 255},
	}
}

// SetColors sets the two colors used by NoiseDuo mode.
// Returns the receiver for chaining.
func (e *NoiseEffect) SetColors(a, b patterns.Color) *NoiseEffect {
	e.colorA, e.colorB = a, b
	return e
}

// SetOpacity sets the blending opacity for noise application in [0,1].
// 0 disables blending, 1 applies full strength.
// Returns the receiver for chaining.
func (e *NoiseEffect) SetOpacity(v float64) *NoiseEffect {
	e.opacity = geom.ClampF64(v, 0, 1)
	return e
}

// Name returns the effect identifier.
// Implements the Effect interface.
func (e *NoiseEffect) Name() string {
	return "Noise"
}

// IsPre indicates whether the effect should be applied before drawing.
// NoiseEffect is always post-applied, so this returns false.
func (e *NoiseEffect) IsPre() bool {
	return false
}

// Apply executes the noise effect on the given RGBA image.
//
// For each pixel within bounds:
//  1. With probability = density, generate noise according to the selected type.
//  2. Linearly interpolate destination RGB channels toward the noise color
//     using the specified opacity.
//
// Alpha channel remains unchanged.
func (e *NoiseEffect) Apply(dst *image.RGBA) {
	if e.opacity == 0 {
		return
	}

	b := dst.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			// Skip most pixels according to density
			if rand.Float64() > e.density {
				continue
			}

			i := dst.PixOffset(x, y)
			switch e.noiseType {

			case NoiseMono:
				// Grayscale noise
				v := uint8(rand.Intn(256))
				dst.Pix[i+0] = uint8(geom.Lerp(float64(dst.Pix[i+0]), float64(v), e.opacity))
				dst.Pix[i+1] = uint8(geom.Lerp(float64(dst.Pix[i+1]), float64(v), e.opacity))
				dst.Pix[i+2] = uint8(geom.Lerp(float64(dst.Pix[i+2]), float64(v), e.opacity))

			case NoiseDuo:
				// Two-color noise
				c := e.colorA
				if rand.Intn(2) == 1 {
					c = e.colorB
				}
				dst.Pix[i+0] = uint8(geom.Lerp(float64(dst.Pix[i+0]), float64(c.R), e.opacity))
				dst.Pix[i+1] = uint8(geom.Lerp(float64(dst.Pix[i+1]), float64(c.G), e.opacity))
				dst.Pix[i+2] = uint8(geom.Lerp(float64(dst.Pix[i+2]), float64(c.B), e.opacity))

			case NoiseMulti:
				// Full RGB random noise
				c := patterns.Color{
					R: uint8(rand.Intn(256)),
					G: uint8(rand.Intn(256)),
					B: uint8(rand.Intn(256)),
					A: 255,
				}
				dst.Pix[i+0] = uint8(geom.Lerp(float64(dst.Pix[i+0]), float64(c.R), e.opacity))
				dst.Pix[i+1] = uint8(geom.Lerp(float64(dst.Pix[i+1]), float64(c.G), e.opacity))
				dst.Pix[i+2] = uint8(geom.Lerp(float64(dst.Pix[i+2]), float64(c.B), e.opacity))
			}
		}
	}
}
