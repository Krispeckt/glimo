package instructions

import (
	"math"
)

// node stores pre-computed layout and measurement data for a single child.
// meas and pos point to the same object when Shape implements BoundedShape.
type node struct {
	shape Shape
	meas  BoundedShape // used to query intrinsic size
	pos   BoundedShape // used to update absolute coordinates
	st    ItemStyle
	x, y  int // computed top-left position
	w, h  int // computed width and height
}

// Resizable is an optional capability: layout passes resolved size to the shape.
type Resizable interface {
	SetSize(w, h int)
}

// Boundable is an optional capability: layout passes x,y,w,h in one call.
type Boundable interface {
	SetBounds(x, y, w, h int)
}

// sum4 expands [top,right,bottom,left].
func sum4(a [4]int) (t, r, b, l int) { return a[0], a[1], a[2], a[3] }

// naturalSize returns the shape’s intrinsic width/height,
// respecting explicit Width/Height overrides in ItemStyle.
func naturalSize(n *node) (w, h int) {
	if n.meas != nil {
		size := n.meas.Size()
		w = int(math.Round(size.Width()))
		h = int(math.Round(size.Height()))
	}
	if n.st.Width > 0 {
		w = n.st.Width
	}
	if n.st.Height > 0 {
		h = n.st.Height
	}
	return
}

// baseMainCross returns an item’s base contribution along main and cross axes,
// including margins. Used during line construction.
func baseMainCross(n *node, isRow bool) (baseMain, baseCross int) {
	w, h := naturalSize(n)
	mt, mr, mb, ml := sum4(n.st.Margin)

	if isRow {
		if n.st.FlexBasis > 0 {
			baseMain = n.st.FlexBasis + ml + mr
		} else if n.st.Width > 0 {
			baseMain = n.st.Width + ml + mr
		} else {
			baseMain = w + ml + mr
		}
		baseCross = h + mt + mb
		return
	}

	if n.st.FlexBasis > 0 {
		baseMain = n.st.FlexBasis + mt + mb
	} else if n.st.Height > 0 {
		baseMain = n.st.Height + mt + mb
	} else {
		baseMain = h + mt + mb
	}
	baseCross = w + ml + mr
	return
}

// resolveCrossSize computes an item’s effective cross-axis size
// before AlignItems stretching is applied.
func resolveCrossSize(n *node, isRow bool, nw, nh int) int {
	if isRow {
		if n.st.Height > 0 {
			return n.st.Height
		}
		return nh
	}
	if n.st.Width > 0 {
		return n.st.Width
	}
	return nw
}

// line groups items that belong to the same row or column when wrapping.
type line struct {
	items       []*node
	base        int // total main-axis length (including margins; fixed gaps added during build)
	cross       int // maximum cross-axis length (including margins)
	skippedGaps int // count of ignored container gaps in this line
}
