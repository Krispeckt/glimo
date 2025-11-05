package instructions

import (
	"image"
	"sort"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// AutoLayout is a flexible container that arranges child shapes according to
// Flexbox-like rules and draws them to an overlay image. It never modifies the base layer.
type AutoLayout struct {
	x, y     int // container origin
	style    ContainerStyle
	children []*node
	w, h     int
	dirty    bool // marks layout as invalidated
}

// NewAutoLayout constructs a new flex container anchored at (x, y).
// If Display is not DisplayFlex, it is forced to DisplayFlex.
func NewAutoLayout(x, y int, style ContainerStyle) *AutoLayout {
	if style.Display != DisplayFlex {
		style.Display = DisplayFlex
	}
	return &AutoLayout{x: x, y: y, style: style, dirty: true}
}

// Add registers a child Shape with an optional ItemStyle.
// If the shape implements BoundedShape, its size and position are queried/updated automatically.
func (al *AutoLayout) Add(s Shape, st ItemStyle) *AutoLayout {
	n := &node{shape: s, st: st}
	if bs, ok := s.(BoundedShape); ok {
		n.meas = bs
		n.pos = bs
	}
	al.children = append(al.children, n)
	al.w, al.h = 0, 0
	al.dirty = true
	return al
}

// SetStyle replaces the container style and invalidates the current layout.
func (al *AutoLayout) SetStyle(style ContainerStyle) {
	if style.Display != DisplayFlex {
		style.Display = DisplayFlex
	}
	al.style = style
	al.w, al.h = 0, 0
	al.dirty = true
}

// Size returns the outer dimensions of the container including padding.
// Triggers layout if needed.
func (al *AutoLayout) Size() *geom.Size {
	al.ensureLayout()
	return geom.NewSize(float64(al.w), float64(al.h))
}

// Draw performs layout, sorts children by ZIndex, and draws each one in order.
// Shapes implementing Boundable receive SetBounds; else Position/Size are propagated if available.
func (al *AutoLayout) Draw(base, overlay *image.RGBA) {
	al.ensureLayout()
	sort.SliceStable(al.children, func(i, j int) bool {
		return al.children[i].st.ZIndex < al.children[j].st.ZIndex
	})
	for _, n := range al.children {
		// Propagate resolved bounds to the shape if supported.
		if b, ok := n.shape.(Boundable); ok {
			b.SetBounds(n.x, n.y, n.w, n.h)
		} else {
			if n.pos != nil {
				n.pos.SetPosition(n.x, n.y)
			}
			if rs, ok := n.shape.(Resizable); ok {
				rs.SetSize(n.w, n.h)
			}
		}
		n.shape.Draw(base, overlay)
	}
}

// ensureLayout computes a fresh layout if dirty or empty.
func (al *AutoLayout) ensureLayout() {
	if al.dirty || (al.w == 0 && al.h == 0) {
		al.layoutFlex()
		al.dirty = false
	}
}
