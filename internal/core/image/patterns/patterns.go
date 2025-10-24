package patterns

import (
	"image/color"
)

// Pattern defines a minimal interface for any drawable color pattern.
// It provides a method for sampling the color at a given pixel coordinate (x, y).
type Pattern interface {
	// ColorAt returns the color value at the specified pixel coordinate (x, y).
	ColorAt(x, y int) color.Color
}

// BlendedPattern extends Pattern with blending and opacity properties.
// This allows the pattern to define how it interacts visually with underlying content.
type BlendedPattern interface {
	Pattern
	// BlendMode returns the blending mode applied when the pattern is composited.
	BlendMode() BlendMode
	// Opacity returns the opacity factor in the range [0, 1].
	Opacity() float64
}

// GradientPattern defines a shared interface for all gradient-based patterns,
// such as linear, radial, or conic gradients. It supports adding color stops
// and includes blending and opacity control.
type GradientPattern interface {
	BlendedPattern
	// AddColorStop adds a color stop to the gradient at the given normalized position [0, 1].
	// Returns the same gradient instance for method chaining.
	AddColorStop(offset float64, c Color) GradientPattern
}
