package geom

import "image"

// Size represents a 2D dimension with width and height values as float64.
// It is commonly used to describe image or geometric dimensions with precision.
type Size struct {
	width  float64
	height float64
}

// NewSize creates a new Size instance with the specified width and height.
func NewSize(width, height float64) *Size {
	return &Size{width: width, height: height}
}

// NewSizeFromImage creates a new Size instance based on the dimensions
// of the provided image.Image. It reads the width and height from the image bounds.
func NewSizeFromImage(img image.Image) *Size {
	size := img.Bounds().Size()
	return NewSize(float64(size.X), float64(size.Y))
}

// Width returns the width value.
func (s *Size) Width() float64 {
	return s.width
}

// Height returns the height value.
func (s *Size) Height() float64 {
	return s.height
}

// Set updates the width and height values.
func (s *Size) Set(width, height float64) {
	s.width = width
	s.height = height
}

// AspectRatio returns the width-to-height ratio as a float64.
// Returns 0 if height is 0 to avoid division by zero.
func (s *Size) AspectRatio() float64 {
	if s.height == 0 {
		return 0
	}
	return s.width / s.height
}

// IsZero checks whether both width and height are zero.
func (s *Size) IsZero() bool {
	return s.width == 0 && s.height == 0
}

// Copy returns a deep copy of the Size.
func (s *Size) Copy() *Size {
	return &Size{width: s.width, height: s.height}
}
