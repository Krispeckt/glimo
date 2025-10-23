package instructions

import (
	"image"
	"sort"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

// Display describes the layout model used by a container.
// Currently only DisplayFlex is implemented, but the enum allows
// future extension (e.g., grid or block layout). todo
type Display int

const (
	// DisplayFlex enables Flexbox-style layout behavior.
	DisplayFlex Display = iota
)

// FlexDirection defines the orientation of the main axis in the flex container.
type FlexDirection int

const (
	// Row lays out items horizontally, left-to-right by default.
	Row FlexDirection = iota
	// Column lays out items vertically, top-to-bottom by default.
	Column
)

// JustifyContent defines how free space is distributed along the main axis.
type JustifyContent int

const (
	JustifyStart        JustifyContent = iota // Items packed at start (default)
	JustifyCenter                             // Items centered along main axis
	JustifyEnd                                // Items packed at end
	JustifySpaceBetween                       // Even spacing between items, none at ends
	JustifySpaceAround                        // Equal spacing around items, half-space at edges
	JustifySpaceEvenly                        // Equal spacing including container edges
)

// AlignItems defines alignment of items along the cross axis within each line.
type AlignItems int

const (
	AlignItemsStart   AlignItems = iota // Align items to start of cross axis
	AlignItemsCenter                    // Align items to center of cross axis
	AlignItemsEnd                       // Align items to end of cross axis
	AlignItemsStretch                   // Stretch items to fill cross axis
)

// PositionType indicates whether an item participates in normal layout flow.
type PositionType int

const (
	// PosRelative — participates in normal flow (default).
	PosRelative PositionType = iota
	// PosAbsolute — removed from flow and positioned relative to container padding box.
	PosAbsolute
)

// ContainerStyle defines CSS-like layout properties for an AutoLayout container.
//
// All numeric units are pixels. Width and Height values of 0 mean
// "auto-size to fit content" depending on the layout direction and children.
type ContainerStyle struct {
	Display       Display
	Direction     FlexDirection
	Wrap          bool
	Padding       [4]int  // top, right, bottom, left
	Gap           Vector2 // gap.X = horizontal spacing, gap.Y = vertical spacing
	Justify       JustifyContent
	AlignItems    AlignItems
	AlignContent  AlignItems // reserved for multi-line cross-axis packing
	Width, Height int        // container dimensions; 0 = auto by content
}

// ItemStyle defines the layout behavior of a single child within a flex container.
type ItemStyle struct {
	Margin     [4]int // top, right, bottom, left
	Width      int    // fixed width; 0 = auto
	Height     int    // fixed height; 0 = auto
	FlexGrow   float64
	FlexShrink float64 // defaults to 1 if 0 and container has negative free space
	FlexBasis  int     // preferred main size in px; 0 = auto → width/height/intrinsic
	AlignSelf  *AlignItems

	// Positioning properties for absolute items.
	Position PositionType
	Top      *int
	Right    *int
	Bottom   *int
	Left     *int

	// Painting order (higher values drawn later).
	ZIndex int
}

// node holds pre-computed layout and measurement data for a single child item.
// meas and pos both point to the same underlying object when it implements BoundedShape.
type node struct {
	shape Shape
	meas  BoundedShape // used to query intrinsic size
	pos   BoundedShape // used to update absolute coordinates
	st    ItemStyle
	x, y  int // computed top-left position
	w, h  int // computed width and height
}

// AutoLayout represents a flexible container that arranges child shapes
// according to flex layout rules and draws them to an overlay image.
//
// It never modifies the base layer directly.
type AutoLayout struct {
	x, y     int // container origin
	style    ContainerStyle
	children []*node
	w, h     int
}

// NewAutoLayout constructs a new flex container anchored at (x, y).
// If Display is not DisplayFlex, it is automatically set.
func NewAutoLayout(x, y int, style ContainerStyle) *AutoLayout {
	if style.Display != DisplayFlex {
		style.Display = DisplayFlex
	}
	return &AutoLayout{x: x, y: y, style: style}
}

// Add registers a child Shape with an optional layout style.
// If the shape implements BoundedShape, its size and position are
// queried and updated automatically.
func (al *AutoLayout) Add(s Shape, st ItemStyle) *AutoLayout {
	n := &node{shape: s, st: st}
	if bs, ok := s.(BoundedShape); ok {
		n.meas = bs
		n.pos = bs
	}
	al.children = append(al.children, n)
	return al
}

// Size returns the current outer dimensions of the AutoLayout container
// as a *geom.Size, including padding on all sides.
func (al *AutoLayout) Size() *geom.Size {
	if al.w == 0 && al.h == 0 {
		isRow := al.style.Direction == Row

		innerW, innerH, _, _, _, _ := al.computeInner(isRow)
		ptop, pright, pbottom, pleft := sum4(al.style.Padding)
		al.w = innerW + pleft + pright
		al.h = innerH + ptop + pbottom
	}
	return geom.NewSize(float64(al.w), float64(al.h))
}

// sum4 expands a 4-value margin or padding array into its components:
// top, right, bottom, left.
func sum4(a [4]int) (t, r, b, l int) { return a[0], a[1], a[2], a[3] }

// naturalSize returns a node's intrinsic width and height,
// respecting explicit Width/Height overrides in its ItemStyle.
func naturalSize(n *node) (w, h int) {
	if n.meas != nil {
		size := n.meas.Size()
		w, h = int(size.Width()), int(size.Height())
	}
	if n.st.Width > 0 {
		w = n.st.Width
	}
	if n.st.Height > 0 {
		h = n.st.Height
	}
	return
}

// baseMainCross computes an item’s base contribution along the main
// and cross axes, including margins. Used during line construction.
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

// resolveCrossSize computes an item’s effective size along the cross axis
// before any AlignItems stretching is applied.
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
	items []*node
	base  int // total main-axis length (including margins)
	cross int // maximum cross-axis length
}

// computeInner calculates container content dimensions (inner width/height)
// and returns padding offsets and gap spacing. Auto-sizing is resolved here.
func (al *AutoLayout) computeInner(isRow bool) (innerW, innerH, pl, pt, gx, gy int) {
	cs := al.style

	// Estimate content size from children.
	natMain, natCross := 0, 0
	for _, n := range al.children {
		if n.meas == nil {
			continue
		}
		w, h := naturalSize(n)
		mt, mr, mb, ml := sum4(n.st.Margin)
		main := w + ml + mr
		cross := h + mt + mb
		if !isRow {
			main, cross = h+mt+mb, w+ml+mr
		}
		natMain += main
		natCross = geom.MaxInt(natCross, cross)
	}

	// Padding and gap.
	pt, pr, pb, pl := sum4(cs.Padding)
	gx, gy = int(cs.Gap.X), int(cs.Gap.Y)

	// Resolve inner content box dimensions.
	if cs.Width > 0 {
		innerW = cs.Width - pl - pr
	} else if isRow {
		innerW = natMain + gx*geom.MaxInt(0, len(al.children)-1)
	} else {
		innerW = natCross
	}
	if cs.Height > 0 {
		innerH = cs.Height - pt - pb
	} else if isRow {
		innerH = natCross
	} else {
		innerH = natMain + gy*geom.MaxInt(0, len(al.children)-1)
	}
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}
	return
}

// buildLines partitions children into flex lines depending on wrapping and main-axis limits.
func (al *AutoLayout) buildLines(isRow bool, mainLimit, gx, gy int) []line {
	var (
		lines []line
		cur   line
	)
	push := func() {
		if len(cur.items) > 0 {
			lines = append(lines, cur)
			cur = line{}
		}
	}

	for _, n := range al.children {
		if n.st.Position == PosAbsolute {
			continue
		}
		baseMain, baseCross := baseMainCross(n, isRow)

		itemWithGap := baseMain
		if len(cur.items) > 0 {
			if isRow {
				itemWithGap += gx
			} else {
				itemWithGap += gy
			}
		}
		if al.style.Wrap && len(cur.items) > 0 && cur.base+itemWithGap > mainLimit {
			push()
		}
		if len(cur.items) > 0 {
			if isRow {
				cur.base += gx
			} else {
				cur.base += gy
			}
		}
		cur.items = append(cur.items, n)
		cur.base += baseMain
		if baseCross > cur.cross {
			cur.cross = baseCross
		}
	}
	push()
	return lines
}

// placeLines assigns coordinates and sizes to all nodes within the computed lines,
// handling justify, align, flex grow/shrink, and stretch logic.
func (al *AutoLayout) placeLines(lines []line, isRow bool, innerW, innerH, pl, pt, gx, gy int) {
	cs := al.style
	mainLimit := innerW
	if !isRow {
		mainLimit = innerH
	}

	crossOffset := 0
	for li := range lines {
		ln := &lines[li]
		free := mainLimit - ln.base

		// Compute grow/shrink totals for proportional space distribution.
		gapMain := gx
		if !isRow {
			gapMain = gy
		}
		totalGaps := 0
		if len(ln.items) > 1 {
			totalGaps = (len(ln.items) - 1) * gapMain
		}
		flexFree := free - totalGaps

		var totalGrow, totalShrink float64
		for _, n := range ln.items {
			totalGrow += n.st.FlexGrow
			if flexFree < 0 {
				if n.st.FlexShrink == 0 {
					totalShrink += 1
				} else {
					totalShrink += n.st.FlexShrink
				}
			} else {
				totalShrink += n.st.FlexShrink
			}
		}

		// JustifyContent offset and inter-item spacing.
		offset, space := 0, 0
		switch cs.Justify {
		case JustifyStart:
			offset = 0
		case JustifyCenter:
			if free > 0 {
				offset = free / 2
			}
		case JustifyEnd:
			if free > 0 {
				offset = free
			}
		case JustifySpaceBetween:
			if free > 0 && len(ln.items) > 1 {
				space = free / (len(ln.items) - 1)
			}
		case JustifySpaceAround:
			if free > 0 && len(ln.items) > 0 {
				space = free / len(ln.items)
				offset = space / 2
			}
		case JustifySpaceEvenly:
			if free > 0 && len(ln.items) > 0 {
				space = free / (len(ln.items) + 1)
				offset = space
			}
		}

		mainCursor := offset
		for idx, n := range ln.items {
			nw, nh := naturalSize(n)
			mt, mr, mb, ml := sum4(n.st.Margin)

			// Base size along main axis
			var baseMain int
			if isRow {
				if n.st.FlexBasis > 0 {
					baseMain = n.st.FlexBasis
				} else if n.st.Width > 0 {
					baseMain = n.st.Width
				} else {
					baseMain = nw
				}
			} else {
				if n.st.FlexBasis > 0 {
					baseMain = n.st.FlexBasis
				} else if n.st.Height > 0 {
					baseMain = n.st.Height
				} else {
					baseMain = nh
				}
			}

			// Apply flex grow/shrink adjustments.
			sizeMain := baseMain
			if flexFree > 0 && totalGrow > 0 {
				sizeMain += int(float64(flexFree) * (n.st.FlexGrow / totalGrow))
			} else if flexFree < 0 && totalShrink > 0 {
				sh := n.st.FlexShrink
				if sh == 0 {
					sh = 1
				}
				sizeMain += int(float64(flexFree) * (sh / totalShrink))
				if sizeMain < 0 {
					sizeMain = 0
				}
			}

			// Cross-axis sizing and alignment.
			sizeCross := resolveCrossSize(n, isRow, nw, nh)
			align := cs.AlignItems
			if n.st.AlignSelf != nil {
				align = *n.st.AlignSelf
			}
			crossPos := 0
			switch align {
			case AlignItemsStart:
				crossPos = mt
			case AlignItemsCenter:
				crossPos = (ln.cross-sizeCross-(mt+mb))/2 + mt
			case AlignItemsEnd:
				crossPos = ln.cross - sizeCross - mb
			case AlignItemsStretch:
				sizeCross = geom.MaxInt(1, ln.cross-(mt+mb))
				crossPos = mt
			}

			// Final coordinates in container space.
			if isRow {
				n.x = al.x + pl + mainCursor + ml
				n.y = al.y + pt + crossOffset + crossPos
				n.w = sizeMain
				n.h = sizeCross

				mainCursor += sizeMain + ml + mr
				if idx < len(ln.items)-1 {
					mainCursor += gx + space
				}
			} else {
				n.x = al.x + pl + crossOffset + crossPos
				n.y = al.y + pt + mainCursor + mt
				n.w = sizeCross
				n.h = sizeMain

				mainCursor += sizeMain + mt + mb
				if idx < len(ln.items)-1 {
					mainCursor += gy + space
				}
			}
		}

		// Move to next line.
		if isRow {
			crossOffset += ln.cross + gy
		} else {
			crossOffset += ln.cross + gx
		}
	}
}

// positionAbsolute sets coordinates for out-of-flow elements (PosAbsolute)
// relative to the container’s padding box.
func (al *AutoLayout) positionAbsolute(innerW, innerH, pl, pt int) {
	for _, n := range al.children {
		if n.st.Position != PosAbsolute {
			continue
		}
		w, h := naturalSize(n)
		n.w, n.h = w, h

		cx0 := al.x + pl
		cy0 := al.y + pt
		cx1 := cx0 + innerW
		cy1 := cy0 + innerH

		x, y := cx0, cy0
		if n.st.Left != nil {
			x = cx0 + *n.st.Left
		} else if n.st.Right != nil {
			x = cx1 - *n.st.Right - n.w
		}
		if n.st.Top != nil {
			y = cy0 + *n.st.Top
		} else if n.st.Bottom != nil {
			y = cy1 - *n.st.Bottom - n.h
		}
		n.x, n.y = x, y
	}
}

// layoutFlex executes the entire flex layout computation pipeline,
// returning the resolved inner content box dimensions.
func (al *AutoLayout) layoutFlex() (innerW, innerH int) {
	isRow := al.style.Direction == Row

	innerW, innerH, pl, pt, gx, gy := al.computeInner(isRow)

	mainLimit := innerW
	if !isRow {
		mainLimit = innerH
	}

	lines := al.buildLines(isRow, mainLimit, gx, gy)
	al.placeLines(lines, isRow, innerW, innerH, pl, pt, gx, gy)
	al.positionAbsolute(innerW, innerH, pl, pt)

	ptop, pright, pbottom, pleft := sum4(al.style.Padding)
	al.w = innerW + pleft + pright
	al.h = innerH + ptop + pbottom

	return innerW, innerH
}

// Draw performs layout, sorts children by ZIndex, and draws each one in order.
// Shapes that implement BoundedShape receive updated coordinates via SetPosition
// before drawing.
func (al *AutoLayout) Draw(base, overlay *image.RGBA) {
	al.layoutFlex()
	sort.SliceStable(al.children, func(i, j int) bool {
		return al.children[i].st.ZIndex < al.children[j].st.ZIndex
	})
	for _, n := range al.children {
		if n.pos != nil {
			n.pos.SetPosition(n.x, n.y)
		}
		n.shape.Draw(base, overlay)
	}
}
