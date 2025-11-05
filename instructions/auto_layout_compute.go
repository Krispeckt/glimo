package instructions

// computeInner calculates content-box dimensions (inner width/height)
// and returns padding offsets and gap spacing. Auto-sizing is resolved
// only as an estimate here. Exact auto cross-size for wrapped content is
// resolved later in layoutFlex after lines are placed.
func (al *AutoLayout) computeInner(isRow bool) (innerW, innerH, pl, pt, gx, gy int) {
	cs := al.style

	// Estimate content size assuming a single line.
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

	innerW = clampNonNegative(innerW)
	innerH = clampNonNegative(innerH)
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

		// Tentative line length if this item is added.
		itemWithGap := baseMain
		if len(cur.items) > 0 {
			if !n.st.IgnoreGapBefore {
				itemWithGap += gapMain
			}
		}

		// Effective line limit, respecting auto-height column logic.
		effectiveLimit := mainLimit
		if !autoHeightColumn && effectiveLimit > 0 {
			// placeholder for future adjustments
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
