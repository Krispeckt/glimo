package instructions

// layoutFlex executes the full flex layout pipeline, computing final positions and sizes.
// It resolves auto cross-sizes, applies wrapping, gaps, and IgnoreGapBefore logic.
func (al *AutoLayout) layoutFlex() (innerW, innerH int) {
	isRow := al.style.Direction == Row

	innerW, innerH, pl, pt, gx, gy := al.computeInner(isRow)
	mainLimit := innerW
	if !isRow {
		mainLimit = innerH
	}

	lines := al.buildLines(isRow, mainLimit, gx, gy)

	// Initial auto cross-size estimate before placement.
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
