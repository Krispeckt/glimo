package instructions

import (
	"image"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// Shape defines the minimal contract for any drawable visual entity or
// rendering instruction in the system.
//
// Implementations of Shape are responsible for drawing visual output onto
// one or both of the provided RGBA image buffers. These buffers represent
// different layers in the rendering pipeline:
//
//   - base:    The read-only background image. Can be used for sampling or
//     blending reference data but should not be modified directly.
//   - overlay: The active target buffer where the shape should render its
//     visual output.
//
// The drawing function should confine its modifications to the intended
// area of the shape and must not alter unrelated pixels outside its bounds.
// Each Shape implementation defines its own compositing, blending, and
// color behavior.
type Shape interface {
	// Draw renders the shape’s visual representation into the given
	// base and overlay image buffers. Implementations decide how to
	// blend, composite, or otherwise use the provided layers.
	Draw(base, overlay *image.RGBA)
}

// BoundedShape extends Shape with positional and dimensional metadata.
//
// It introduces the concept of a rectangular bounding box, which defines
// both the size of the drawable object and its placement within a
// two-dimensional scene. This interface allows layout managers, compositors,
// and spatial algorithms to arrange and interact with multiple shapes
// consistently.
type BoundedShape interface {
	Shape

	// Size returns the intended width and height of the shape as a *geom.Size.
	// A zero value for either axis typically indicates that the shape should
	// determine its size automatically from its content or intrinsic geometry.
	Size() *geom.Size

	// Position returns the current top-left coordinate of the shape in
	// integer pixel units. This defines the anchor point used for rendering
	// and layout alignment.
	Position() (int, int)

	// SetPosition updates the shape’s anchor point to the given (x, y)
	// coordinate. This does not automatically trigger a redraw; the new
	// position is applied the next time Draw() is called.
	SetPosition(x, y int)
}
