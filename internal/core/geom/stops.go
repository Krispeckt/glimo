package geom

import "image/color"

// Gradient Color Stop

// Stop represents a single gradient stop with a color and its normalized position.
//
// Position `pos` is typically in the range [0, 1], where:
//   - 0 corresponds to the start of the gradient,
//   - 1 corresponds to the end of the gradient.
type Stop struct {
	pos   float64
	color color.Color
}

// Color returns the color associated with this stop.
func (s *Stop) Color() color.Color {
	return s.color
}

// Position returns the normalized position of this stop along the gradient.
func (s *Stop) Position() float64 {
	return s.pos
}

// NewStop creates and returns a new gradient stop with the given position and color.
func NewStop(pos float64, color color.Color) *Stop {
	return &Stop{
		pos:   pos,
		color: color,
	}
}

// Gradient Stop Collection

// Stops represents a slice of gradient stops.
//
// It implements sort.Interface, allowing stops to be sorted by their position.
type Stops []*Stop

// Len returns the number of stops in the collection.
func (s Stops) Len() int { return len(s) }

// Swap exchanges two stops in the collection.
func (s Stops) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Less reports whether the stop at index i appears before the stop at index j.
func (s Stops) Less(i, j int) bool { return s[i].pos < s[j].pos }
