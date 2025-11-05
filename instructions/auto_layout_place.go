package instructions

import (
	"math"
	"sort"

	"github.com/Krispeckt/glimo/internal/core/geom"
)

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

		// Free space available for flexing (after margins and fixed gaps).
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
