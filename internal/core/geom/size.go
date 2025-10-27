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

// Union returns a new Size that is the axis-aligned union of s and other,
// assuming both are anchored at the same origin. The result width and height
// are the component-wise maxima. Nil receivers are treated as zero sizes.
func (s *Size) Union(other *Size) *Size {
	if s == nil && other == nil {
		return NewSize(0, 0)
	}
	if s == nil {
		return other.Copy()
	}
	if other == nil {
		return s.Copy()
	}
	w := s.width
	if other.width > w {
		w = other.width
	}
	h := s.height
	if other.height > h {
		h = other.height
	}
	return NewSize(w, h)
}

// UnionInPlace expands s to be the union with other. No-op if s is nil.
func (s *Size) UnionInPlace(other *Size) {
	if s == nil || other == nil {
		return
	}
	if other.width > s.width {
		s.width = other.width
	}
	if other.height > s.height {
		s.height = other.height
	}
	if s.width < 0 {
		s.width = 0
	}
	if s.height < 0 {
		s.height = 0
	}
}

// UnionAll returns the union of all provided sizes. Nil entries are treated as zero.
func UnionAll(sizes ...*Size) *Size {
	out := NewSize(0, 0)
	for _, sz := range sizes {
		out.UnionInPlace(sz)
	}
	return out
}
