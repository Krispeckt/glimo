package instructions

import (
	"image"
	"math"
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
	AlignContent  AlignItems // cross-axis packing for multiple lines: Start/Center/End/Stretch
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

	// IgnoreGapBefore skips the container gap directly before this item.
	// This affects line construction, wrapping, and final positioning.
	IgnoreGapBefore bool
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

// Resizable is an optional capability. If implemented by a shape,
// AutoLayout will pass the resolved width and height to the shape.
type Resizable interface {
	SetSize(w, h int)
}

// Boundable is an optional capability. If implemented by a shape,
// AutoLayout will pass the resolved position and size in one call.
type Boundable interface {
	SetBounds(x, y, w, h int)
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
	dirty    bool // marks layout as invalidated
}

// NewAutoLayout constructs a new flex container anchored at (x, y).
// If Display is not DisplayFlex, it is automatically set.
func NewAutoLayout(x, y int, style ContainerStyle) *AutoLayout {
	if style.Display != DisplayFlex {
		style.Display = DisplayFlex
	}
	return &AutoLayout{x: x, y: y, style: style, dirty: true}
}

// Add registers a child Shape with an optional layout style.
// If the shape implements BoundedShape, its size and position are
// queried and updated automatically.
func (al *AutoLayout) Add(s Shape, st ItemStyle) *AutoLayout {
	n := &node{shape: s, st: st}

	if g, ok := s.(*Group); ok {
		g.SetMode(FrameMode)
	}
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

// Size returns the current outer dimensions of the AutoLayout container
// as a *geom.Size, including padding on all sides. If the layout is dirty,
// it is recomputed.
func (al *AutoLayout) Size() *geom.Size {
	al.ensureLayout()
	return geom.NewSize(float64(al.w), float64(al.h))
}

// ensureLayout computes a fresh layout if it is marked dirty or empty.
func (al *AutoLayout) ensureLayout() {
	if al.dirty || (al.w == 0 && al.h == 0) {
		al.layoutFlex()
		al.dirty = false
	}
}

// sum4 expands a 4-value margin or padding array into its components:
// top, right, bottom, left.
func sum4(a [4]int) (t, r, b, l int) { return a[0], a[1], a[2], a[3] }

// naturalSize returns a node's intrinsic width and height,
// respecting explicit Width/Height overrides in its ItemStyle.
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
	items       []*node
	base        int // total main-axis length (including margins; fixed gaps added during build)
	cross       int // maximum cross-axis length (including margins)
	skippedGaps int // count of ignored container gaps in this line
}

// computeInner calculates container content dimensions (inner width/height)
// and returns padding offsets and gap spacing. Auto-sizing is resolved here
// only as an estimate. Exact auto cross-size for wrapped content is resolved
// later in layoutFlex after lines are built.
func (al *AutoLayout) computeInner(isRow bool) (innerW, innerH, pl, pt, gx, gy int) {
	cs := al.style

	// Estimate content size from children in a single line.
	natMain, natCross := 0, 0
	count := 0
	for _, n := range al.children {
		if n.st.Position == PosAbsolute {
			continue
		}
		baseMain, baseCross := baseMainCross(n, isRow)
		natMain += baseMain
		if baseCross > natCross {
			natCross = baseCross
		}
		count++
	}

	// Padding and gap.
	pt, pr, pb, pl := sum4(cs.Padding)
	gx, gy = int(cs.Gap.X), int(cs.Gap.Y)

	// Resolve inner content box dimensions.
	if cs.Width > 0 {
		innerW = cs.Width - pl - pr
	} else if isRow {
		innerW = natMain
		if count > 1 {
			innerW += gx * (count - 1)
		}
	} else {
		innerW = natCross
	}
	if cs.Height > 0 {
		innerH = cs.Height - pt - pb
	} else if isRow {
		innerH = natCross
	} else {
		innerH = natMain
		if count > 1 {
			innerH += gy * (count - 1)
		}
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
// The gap before an item is skipped if that item's ItemStyle.IgnoreGapBefore is true.
// For Column with Height==0 (auto height), gaps are ignored only visually, not in limit calculation.
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

	autoHeightColumn := !isRow && al.style.Height == 0
	gapMain := gx
	if !isRow {
		gapMain = gy
	}

	for _, n := range al.children {
		if n.st.Position == PosAbsolute {
			continue
		}
		baseMain, baseCross := baseMainCross(n, isRow)

		// Compute tentative line length if this item is added.
		itemWithGap := baseMain
		if len(cur.items) > 0 {
			if !n.st.IgnoreGapBefore {
				itemWithGap += gapMain
			}
		}

		// Compute the effective line limit, respecting auto-height column logic.
		effectiveLimit := mainLimit
		if !autoHeightColumn && effectiveLimit > 0 {
			effectiveLimit -= 0 // keep consistent path, placeholder for future adjustments
			if effectiveLimit < 0 {
				effectiveLimit = 0
			}
		}

		// Wrap to a new line if needed.
		if al.style.Wrap && len(cur.items) > 0 && cur.base+itemWithGap > effectiveLimit {
			push()
		}

		// Add this item to the current line.
		if len(cur.items) > 0 {
			if !n.st.IgnoreGapBefore {
				cur.base += gapMain
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
// handling justify, align, flex grow/shrink, line distribution, and stretch logic.
// The fixed container gap before an item is skipped if that item's IgnoreGapBefore is true.
// For each line, the effective main-axis limit is reduced by the number of skipped gaps.
func (al *AutoLayout) placeLines(lines []line, isRow bool, innerW, innerH, pl, pt, gx, gy int) {
	cs := al.style

	mainLimit := innerW
	crossLimit := innerH
	gapMain := gx
	gapCross := gy
	if !isRow {
		mainLimit = innerH
		crossLimit = innerW
		gapMain = gy
		gapCross = gx
	}

	// Cross-axis distribution across multiple lines (AlignContent).
	totalCross := 0
	if len(lines) > 0 {
		for _, ln := range lines {
			totalCross += ln.cross
		}
		totalCross += gapCross * (len(lines) - 1)
	}
	leftoverCross := crossLimit - totalCross
	if leftoverCross < 0 {
		leftoverCross = 0
	}
	crossStartOffset := 0
	extraPerLine := 0
	switch cs.AlignContent {
	case AlignItemsCenter:
		crossStartOffset = leftoverCross / 2
	case AlignItemsEnd:
		crossStartOffset = leftoverCross
	case AlignItemsStretch:
		if len(lines) > 0 && leftoverCross > 0 {
			extraPerLine = leftoverCross / len(lines)
		}
	default: // Start
	}

	crossOffset := crossStartOffset

	for li := range lines {
		ln := &lines[li]

		// Line-specific effective main-axis limit after subtracting skipped gaps.
		effectiveLimit := mainLimit - ln.skippedGaps*gapMain
		if effectiveLimit < 0 {
			effectiveLimit = 0
		}

		// Precompute base sizes and factors per item.
		type itemRec struct {
			n                *node
			baseMainContent  int // without margins
			baseMainWithMarg int // with margins
			mt, mr, mb, ml   int
			nw, nh           int
			sizeMainContent  int // resolved main size without margins
			sizeCross        int // resolved cross size without margins
			align            AlignItems
		}
		recs := make([]itemRec, 0, len(ln.items))
		var totalGrow, totalShrink float64
		sumBaseWithMargins := 0

		for _, n := range ln.items {
			nw, nh := naturalSize(n)
			mt, mr, mb, ml := sum4(n.st.Margin)

			// Base main content size (no margins).
			var baseMainContent int
			if isRow {
				if n.st.FlexBasis > 0 {
					baseMainContent = n.st.FlexBasis
				} else if n.st.Width > 0 {
					baseMainContent = n.st.Width
				} else {
					baseMainContent = nw
				}
			} else {
				if n.st.FlexBasis > 0 {
					baseMainContent = n.st.FlexBasis
				} else if n.st.Height > 0 {
					baseMainContent = n.st.Height
				} else {
					baseMainContent = nh
				}
			}

			baseWithMargins := baseMainContent
			if isRow {
				baseWithMargins += ml + mr
			} else {
				baseWithMargins += mt + mb
			}
			sumBaseWithMargins += baseWithMargins

			rec := itemRec{
				n:                n,
				baseMainContent:  baseMainContent,
				baseMainWithMarg: baseWithMargins,
				mt:               mt, mr: mr, mb: mb, ml: ml,
				nw: nw, nh: nh,
				align: func() AlignItems {
					if n.st.AlignSelf != nil {
						return *n.st.AlignSelf
					}
					return cs.AlignItems
				}(),
			}
			recs = append(recs, rec)

			// Flex factors.
			totalGrow += n.st.FlexGrow
			if n.st.FlexShrink == 0 {
				totalShrink += 1
			} else {
				totalShrink += n.st.FlexShrink
			}
		}

		// Count fixed gaps actually used between items, honoring IgnoreGapBefore of the next item.
		totalGaps := 0
		if len(recs) > 1 {
			for i := 1; i < len(recs); i++ {
				if !recs[i].n.st.IgnoreGapBefore {
					totalGaps += gapMain
				}
			}
		}

		// Free space available for flexing (after margins and fixed gaps) with adjusted limit.
		flexFree := effectiveLimit - sumBaseWithMargins - totalGaps

		// Distribute flex grow/shrink with remainder handling.
		switch {
		case flexFree > 0 && totalGrow > 0:
			// First pass: floors.
			floors := make([]int, len(recs))
			fracs := make([]float64, len(recs))
			sumFloors := 0
			for i, r := range recs {
				share := float64(flexFree) * (r.n.st.FlexGrow / totalGrow)
				f := int(math.Floor(share))
				floors[i] = f
				fracs[i] = share - float64(f)
				sumFloors += f
			}
			rem := flexFree - sumFloors
			// Assign remainders by descending fractional parts.
			idx := make([]int, len(recs))
			for i := range idx {
				idx[i] = i
			}
			sort.Slice(idx, func(i, j int) bool { return fracs[idx[i]] > fracs[idx[j]] })
			for k := 0; k < rem && k < len(idx); k++ {
				floors[idx[k]]++
			}
			for i := range recs {
				recs[i].sizeMainContent = recs[i].baseMainContent + floors[i]
				if recs[i].sizeMainContent < 0 {
					recs[i].sizeMainContent = 0
				}
			}

		case flexFree < 0 && totalShrink > 0:
			need := -flexFree
			floors := make([]int, len(recs))
			fracs := make([]float64, len(recs))
			sumFloors := 0
			for i, r := range recs {
				sh := r.n.st.FlexShrink
				if sh == 0 {
					sh = 1
				}
				share := float64(need) * (sh / totalShrink)
				f := int(math.Floor(share))
				floors[i] = f
				fracs[i] = share - float64(f)
				sumFloors += f
			}
			rem := need - sumFloors
			idx := make([]int, len(recs))
			for i := range idx {
				idx[i] = i
			}
			sort.Slice(idx, func(i, j int) bool { return fracs[idx[i]] > fracs[idx[j]] })
			for k := 0; k < rem && k < len(idx); k++ {
				floors[idx[k]]++
			}
			for i := range recs {
				recs[i].sizeMainContent = recs[i].baseMainContent - floors[i]
				if recs[i].sizeMainContent < 0 {
					recs[i].sizeMainContent = 0
				}
			}
		default:
			for i := range recs {
				recs[i].sizeMainContent = recs[i].baseMainContent
			}
		}

		// After flexing, recompute remaining free space for justify-content.
		used := 0
		for _, r := range recs {
			if isRow {
				used += r.sizeMainContent + r.ml + r.mr
			} else {
				used += r.sizeMainContent + r.mt + r.mb
			}
		}
		used += totalGaps
		remaining := effectiveLimit - used
		if remaining < 0 {
			remaining = 0
		}

		// JustifyContent offset and extra spacing.
		offset, extra := 0, 0
		switch cs.Justify {
		case JustifyStart:
			offset = 0
		case JustifyCenter:
			offset = remaining / 2
		case JustifyEnd:
			offset = remaining
		case JustifySpaceBetween:
			if len(recs) > 1 {
				extra = remaining / (len(recs) - 1)
			}
		case JustifySpaceAround:
			if len(recs) > 0 {
				extra = remaining / len(recs)
				offset = extra / 2
			}
		case JustifySpaceEvenly:
			if len(recs) > 0 {
				extra = remaining / (len(recs) + 1)
				offset = extra
			}
		}

		// Resolve cross sizes and positions per item.
		lineCrossSize := ln.cross + extraPerLine
		mainCursor := offset

		for idx, r := range recs {
			// Cross sizing.
			sizeCross := resolveCrossSize(r.n, isRow, r.nw, r.nh)
			switch r.align {
			case AlignItemsStretch:
				// Fill line's cross size minus vertical/horizontal margins.
				if isRow {
					sizeCross = geom.MaxInt(1, lineCrossSize-(r.mt+r.mb))
				} else {
					sizeCross = geom.MaxInt(1, lineCrossSize-(r.ml+r.mr))
				}
			}

			// Cross position inside line.
			crossPos := 0
			switch r.align {
			case AlignItemsStart:
				if isRow {
					crossPos = r.mt
				} else {
					crossPos = r.ml
				}
			case AlignItemsCenter:
				if isRow {
					crossPos = (lineCrossSize-sizeCross-(r.mt+r.mb))/2 + r.mt
				} else {
					crossPos = (lineCrossSize-sizeCross-(r.ml+r.mr))/2 + r.ml
				}
			case AlignItemsEnd:
				if isRow {
					crossPos = lineCrossSize - sizeCross - r.mb
				} else {
					crossPos = lineCrossSize - sizeCross - r.mr
				}
			case AlignItemsStretch:
				if isRow {
					crossPos = r.mt
				} else {
					crossPos = r.ml
				}
			}

			// Final coordinates in container space.
			if isRow {
				x := al.x + pl + mainCursor + r.ml
				y := al.y + pt + crossOffset + crossPos
				w := r.sizeMainContent
				h := sizeCross

				r.n.x, r.n.y = x, y
				r.n.w, r.n.h = w, h

				mainCursor += r.sizeMainContent + r.ml + r.mr
				if idx < len(recs)-1 {
					// Add fixed gap unless the next item skips it, then add justify extra.
					if !recs[idx+1].n.st.IgnoreGapBefore {
						mainCursor += gapMain
					}
					mainCursor += extra
				}
			} else {
				x := al.x + pl + crossOffset + crossPos
				y := al.y + pt + mainCursor + r.mt
				w := sizeCross
				h := r.sizeMainContent

				r.n.x, r.n.y = x, y
				r.n.w, r.n.h = w, h

				mainCursor += r.sizeMainContent + r.mt + r.mb
				if idx < len(recs)-1 {
					if !recs[idx+1].n.st.IgnoreGapBefore {
						mainCursor += gapMain
					}
					mainCursor += extra
				}
			}
		}

		// Move to next line.
		crossOffset += lineCrossSize + gapCross
	}
}

// positionAbsolute sets coordinates for out-of-flow elements (PosAbsolute)
// relative to the container’s padding box. Margins are honored.
func (al *AutoLayout) positionAbsolute(innerW, innerH, pl, pt int) {
	for _, n := range al.children {
		if n.st.Position != PosAbsolute {
			continue
		}
		w, h := naturalSize(n)
		n.w, n.h = w, h

		mt, mr, mb, ml := sum4(n.st.Margin)

		cx0 := al.x + pl
		cy0 := al.y + pt
		cx1 := cx0 + innerW
		cy1 := cy0 + innerH

		x, y := cx0, cy0
		if n.st.Left != nil {
			x = cx0 + *n.st.Left + ml
		} else if n.st.Right != nil {
			x = cx1 - *n.st.Right - n.w - mr
		}
		if n.st.Top != nil {
			y = cy0 + *n.st.Top + mt
		} else if n.st.Bottom != nil {
			y = cy1 - *n.st.Bottom - n.h - mb
		}
		n.x, n.y = x, y
	}
}

// layoutFlex executes the complete flex layout pipeline, computing final positions and sizes.
// It resolves auto cross-sizes, applies wrapping, gaps, and IgnoreGapBefore logic.
func (al *AutoLayout) layoutFlex() (innerW, innerH int) {
	isRow := al.style.Direction == Row

	innerW, innerH, pl, pt, gx, gy := al.computeInner(isRow)
	mainLimit := innerW
	if !isRow {
		mainLimit = innerH
	}

	lines := al.buildLines(isRow, mainLimit, gx, gy)

	// --- Auto cross-size resolution ---
	// Row + auto Height: sum of line cross sizes + gaps.
	if al.style.Height == 0 && isRow {
		sum := 0
		for i, ln := range lines {
			sum += ln.cross
			if i < len(lines)-1 {
				sum += gy
			}
		}
		innerH = sum
	}

	// Column + auto Height: recompute by summing items and only non-ignored gaps.
	if al.style.Height == 0 && !isRow {
		total := 0
		gapMain := gy
		for _, ln := range lines {
			lineMain := 0
			for i, n := range ln.items {
				bm, _ := baseMainCross(n, false) // main=Y for Column
				lineMain += bm
				if i > 0 && !ln.items[i].st.IgnoreGapBefore {
					lineMain += gapMain
				}
			}
			if lineMain > total {
				total = lineMain
			}
		}
		innerH = total
	}

	// Column + auto Width (same as Row + auto Height).
	if al.style.Width == 0 && !isRow {
		sum := 0
		for i, ln := range lines {
			sum += ln.cross
			if i < len(lines)-1 {
				sum += gx
			}
		}
		innerW = sum
	}

	al.placeLines(lines, isRow, innerW, innerH, pl, pt, gx, gy)
	al.positionAbsolute(innerW, innerH, pl, pt)

	ptop, pright, pbottom, pleft := sum4(al.style.Padding)
	al.w = innerW + pleft + pright
	al.h = innerH + ptop + pbottom

	return innerW, innerH
}

// Draw performs layout, sorts children by ZIndex, and draws each one in order.
// Shapes that implement BoundedShape receive updated coordinates via SetPosition
// before drawing. If a shape implements Resizable or Boundable, its size is
// also propagated.
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
