package instructions

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/rivo/uniseg"

	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/Krispeckt/glimo/internal/render"
)

// wrapTextScaled splits text by logical paragraphs, wraps per line using the current WrapMode,
// applies per-line scaling via t.fontForLine, and enforces maxLines with an ellipsis
// when there is undisplayed content.
//
// Notes:
// - Line endings are normalized to '\n'.
// - Unicode grapheme clusters are respected for all symbol-level operations.
// - NBSP (U+00A0) is treated as non-breaking in word mode (it stays inside tokens).
// - Hyphen/wrapSymbol is appended only when breaking inside a "word" boundary.
// - Measurement caching is per Font pointer; pointer stability is assumed.
//
// Complexity:
// - Word mode uses prefix sums per line to avoid string joins during fit checks.
// - Symbol mode uses binary search over grapheme clusters.
func (t *Text) wrapTextScaled() []string {
	// Fast path: no wrapping requested.
	if t.maxWidth <= 0 {
		return strings.Split(normalizeNewlines(t.text), "\n")
	}

	var out []string
	truncated := false
	lineIdx := 0

	// Helper: append a line and, if maxLines is reached while more content exists,
	// append an ellipsis and mark as truncated.
	appendAndMaybeTruncate := func(s string, hasMore bool) {
		if truncated {
			return
		}
		out = append(out, s)
		if t.maxLines > 0 && len(out) == t.maxLines && hasMore {
			lastFont := t.fontForLine(t.maxLines - 1)
			out = appendEllipsisGraphemes(out, lastFont, t.maxWidth)
			truncated = true
		}
	}

	text := normalizeNewlines(t.text)
	paras := strings.Split(text, "\n")

	for pi, p := range paras {
		if truncated {
			break
		}

		// Preserve empty line as paragraph break.
		if p == "" {
			appendAndMaybeTruncate("", pi < len(paras)-1)
			lineIdx++
			continue
		}

		var sub []string
		if t.wrapMode == WrapBySymbol {
			sub = t.wrapParaBySymbolsScaled(p, &lineIdx)
		} else {
			sub = t.wrapParaByWordsScaled(p, &lineIdx)
		}

		for si, s := range sub {
			if truncated {
				break
			}
			hasMore := si < len(sub)-1 || pi < len(paras)-1
			appendAndMaybeTruncate(s, hasMore)
		}

		// Preserve blank line following a paragraph if it exists.
		if !truncated && pi < len(paras)-1 && paras[pi+1] == "" {
			appendAndMaybeTruncate("", pi+1 < len(paras)-1)
			lineIdx++
		}
	}

	return out
}

// wrapParaByWordsScaled wraps a paragraph at word boundaries.
// If a single word exceeds width, it is split progressively by grapheme
// under the current line font to ensure readability.
//
// Tokenization policy:
// - Split only on ASCII space ' ' and TAB '\t'.
// - NBSP (U+00A0) remains inside tokens and will not break lines by itself.
// - Runs of separators collapse to a single gap in output by design.
func (t *Text) wrapParaByWordsScaled(p string, lineIdxPtr *int) []string {
	words := splitWordsPreserveNBSP(p)
	if len(words) == 0 {
		*lineIdxPtr++
		return []string{""}
	}

	var lines []string

	// Local measurement cache per font for lower overhead.
	cache := make(map[*render.Font]map[string]float64)
	measure := func(f *render.Font, s string) float64 {
		if f == nil || s == "" {
			return 0
		}
		m, ok := cache[f]
		if !ok {
			m = make(map[string]float64)
			cache[f] = m
		}
		if w, ok := m[s]; ok {
			return w
		}
		w, _ := f.MeasureString(s)
		if math.IsNaN(w) || w < 0 {
			w = 0
		}
		m[s] = w
		return w
	}

	joinWithSpaces := func(ws []string) string {
		switch len(ws) {
		case 0:
			return ""
		case 1:
			return ws[0]
		default:
			return strings.Join(ws, " ")
		}
	}

	i := 0
	for i < len(words) {
		f := t.fontForLine(*lineIdxPtr)
		width := t.maxWidth

		// If one word is too long, split it progressively by graphemes.
		if measure(f, words[i]) > width {
			chunks := t.splitLongTokenProgressive(words[i], lineIdxPtr, measure)
			lines = append(lines, chunks...)
			i++
			continue
		}

		// Precompute prefix sums of word widths for the current line's font.
		// widthOf(i, j) returns width of words[i:j] with single ASCII spaces between.
		spaceW := measure(f, " ")
		rem := words[i:]
		wW := make([]float64, len(rem))
		for k := range rem {
			wW[k] = measure(f, rem[k])
		}
		pref := make([]float64, len(rem)+1)
		for k := 1; k <= len(rem); k++ {
			pref[k] = pref[k-1] + wW[k-1]
			if k > 1 {
				pref[k] += spaceW
			}
		}
		widthOf := func(a, b int) float64 { // [a, b) over rem
			if a >= b {
				return 0
			}
			return pref[b] - pref[a]
		}

		// Fit as many words as possible using binary search over prefix sums.
		lo, hi := 1, len(rem)
		if wW[0] > width {
			// Defensive: should be handled above.
			lines = append(lines, rem[0])
			*lineIdxPtr++
			i++
			continue
		}
		for lo <= hi {
			mid := (lo + hi) >> 1
			if widthOf(0, mid) <= width {
				lo = mid + 1
			} else {
				hi = mid - 1
			}
		}
		count := hi
		line := joinWithSpaces(rem[:count])
		lines = append(lines, line)
		*lineIdxPtr++
		i += count
	}

	return lines
}

// wrapParaBySymbolsScaled wraps a paragraph by grapheme clusters.
// It optionally appends wrapSymbol at a break if the next cluster continues a word.
func (t *Text) wrapParaBySymbolsScaled(p string, lineIdxPtr *int) []string {
	var lines []string

	clusters, offs := splitGraphemes(p)
	if len(clusters) == 0 {
		*lineIdxPtr++
		return []string{""}
	}

	cache := make(map[*render.Font]map[string]float64)
	measure := func(f *render.Font, s string) float64 {
		if f == nil || s == "" {
			return 0
		}
		m, ok := cache[f]
		if !ok {
			m = make(map[string]float64)
			cache[f] = m
		}
		if w, ok := m[s]; ok {
			return w
		}
		w, _ := f.MeasureString(s)
		if math.IsNaN(w) || w < 0 {
			w = 0
		}
		m[s] = w
		return w
	}

	start := 0
	for start < len(clusters) {
		f := t.fontForLine(*lineIdxPtr)
		width := t.maxWidth

		// Binary search the largest prefix [start:end) that fits.
		lo, hi := start+1, len(clusters)
		best := start
		for lo <= hi {
			mid := (lo + hi) >> 1
			cand := p[offs[start]:offs[mid]]
			if measure(f, cand) <= width {
				best = mid
				lo = mid + 1
			} else {
				hi = mid - 1
			}
		}

		// If nothing fits under this font, force 1 cluster and move on.
		end := best
		if end == start {
			end = start + 1
		}

		line := p[offs[start]:offs[end]]

		// Append wrapSymbol at break if the next cluster continues a word boundary.
		if end < len(clusters) && t.wrapSymbol != "" {
			prevLast := lastBaseRune(line)
			nextFirst := firstBaseRune(clusters[end])
			if isWordBaseRune(prevLast) && isWordBaseRune(nextFirst) {
				// Ensure line+wrapSymbol still fits. If not, shorten by one cluster if possible.
				if measure(f, line+t.wrapSymbol) > width && end > start+1 {
					end--
					line = p[offs[start]:offs[end]]
				}
				if measure(f, line+t.wrapSymbol) <= width {
					line += t.wrapSymbol
				}
			}
		}

		lines = append(lines, trimRightSpacesNBSP(line))
		*lineIdxPtr++
		start = end
	}

	return lines
}

// splitLongTokenProgressive splits a single overlong token by grapheme clusters,
// producing a sequence of lines, each fitting under the current line font.
// It appends wrapSymbol at internal breaks when splitting inside a word.
//
// Contract:
//   - If a single grapheme cluster is wider than maxWidth, that cluster is yielded raw.
//     The caller is responsible for clipping or downstream scaling.
func (t *Text) splitLongTokenProgressive(token string, lineIdxPtr *int, measure func(*render.Font, string) float64) []string {
	var out []string
	if token == "" {
		return out
	}

	clusters, offs := splitGraphemes(token)
	start := 0
	for start < len(clusters) {
		f := t.fontForLine(*lineIdxPtr)
		width := t.maxWidth

		// If even a single cluster does not fit, yield it raw to avoid infinite loop.
		if measure(f, token[offs[start]:offs[start+1]]) > width {
			out = append(out, token[offs[start]:offs[start+1]])
			*lineIdxPtr++
			start++
			continue
		}

		// Binary search the largest prefix that fits; consider room for wrapSymbol if we will split.
		lo, hi := start+1, len(clusters)
		best := start + 1
		for lo <= hi {
			mid := (lo + hi) >> 1
			cand := token[offs[start]:offs[mid]]
			needSuffix := mid < len(clusters)
			if needSuffix && t.wrapSymbol != "" {
				prevLast := lastBaseRune(cand)
				nextFirst := firstBaseRune(clusters[mid])
				if isWordBaseRune(prevLast) && isWordBaseRune(nextFirst) {
					cand += t.wrapSymbol
				}
			}
			if measure(f, cand) <= width {
				best = mid
				lo = mid + 1
			} else {
				hi = mid - 1
			}
		}

		end := best
		line := token[offs[start]:offs[end]]

		// Attach wrapSymbol at internal break if it still fits and is a word boundary.
		if end < len(clusters) && t.wrapSymbol != "" {
			prevLast := lastBaseRune(line)
			nextFirst := firstBaseRune(clusters[end])
			if isWordBaseRune(prevLast) && isWordBaseRune(nextFirst) && measure(f, line+t.wrapSymbol) <= width {
				line += t.wrapSymbol
			}
		}

		out = append(out, line)
		*lineIdxPtr++
		start = end
	}

	return out
}

// autoSpacing estimates inter-line spacing multiplier based on average fill ratio,
// font scaling, and adaptive density heuristics.
// Returns a clamped multiplier in [spacingMin, spacingMax].
//
// Normalization:
// - Averages are computed over the number of lines actually measured to avoid bias.
func (t *Text) autoSpacing(lines []string) float64 {
	const (
		fillMin     = 0.35
		fillMax     = 1.0
		baseSparse  = 1.22
		baseDense   = 0.98
		scaleMin    = 0.5
		scaleMax    = 1.5
		scaleWeight = 0.35
		attenMin    = 0.8
		attenMax    = 1.15
		spacingMin  = 0.85
		spacingMax  = 1.35
	)

	if len(lines) <= 1 || t.maxWidth <= 0 || t.font == nil {
		return 1.0
	}

	basePt := t.font.HeightPt()
	if basePt <= 0 {
		return 1.0
	}

	var totalWidth, totalScale float64
	var countW, countS int

	for i, s := range lines {
		lf := t.fontForLine(i)
		if lf == nil {
			continue
		}
		if w, _ := lf.MeasureString(s); !math.IsNaN(w) {
			if w < 0 {
				w = 0
			}
			totalWidth += w
			countW++
		}
		if curPt := lf.HeightPt(); curPt > 0 {
			totalScale += curPt / basePt
			countS++
		}
	}

	var fill float64 = 1.0
	if countW > 0 {
		fill = geom.ClampF64(totalWidth/(t.maxWidth*float64(countW)), fillMin, fillMax)
	}
	fillT := (fill - fillMin) / (fillMax - fillMin)
	base := geom.Lerp(baseSparse, baseDense, fillT)

	var avgScale float64 = 1.0
	if countS > 0 {
		avgScale = geom.ClampF64(totalScale/float64(countS), scaleMin, scaleMax)
	}
	atten := geom.ClampF64(1.0-scaleWeight*(1.0-avgScale), attenMin, attenMax)

	return geom.ClampF64(base*atten, spacingMin, spacingMax)
}

// appendEllipsisGraphemes trims the final line so that an ellipsis ("…") fits.
// It removes text by grapheme clusters to avoid breaking composite glyphs.
// If even a single ellipsis does not fit, the line is left as-is.
// Trailing ASCII spaces and NBSP are trimmed before appending.
func appendEllipsisGraphemes(lines []string, f *render.Font, maxWidth float64) []string {
	const ellipsis = "…"
	if len(lines) == 0 || f == nil {
		return lines
	}

	lastIdx := len(lines) - 1
	last := trimRightSpacesNBSP(lines[lastIdx])

	// If it already fits with the ellipsis, append and return.
	if w, _ := f.MeasureString(last + ellipsis); w <= maxWidth {
		lines[lastIdx] = last + ellipsis
		return lines
	}

	// Remove by grapheme clusters from the end until it fits.
	grs, offs := splitGraphemes(last)
	for len(grs) > 0 {
		grs = grs[:len(grs)-1]
		cut := last[:offs[len(grs)]]
		if w, _ := f.MeasureString(cut + ellipsis); w <= maxWidth {
			lines[lastIdx] = cut + ellipsis
			return lines
		}
	}

	// As a fallback, try ellipsis alone.
	if w, _ := f.MeasureString(ellipsis); w <= maxWidth {
		lines[lastIdx] = ellipsis
	}
	return lines
}

// normalizeNewlines converts CRLF and CR to LF.
func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// splitGraphemes returns grapheme clusters and their string offsets.
// Offsets are len-granular byte indices into the original string.
func splitGraphemes(s string) (clusters []string, offsets []int) {
	g := uniseg.NewGraphemes(s)
	offsets = append(offsets, 0)
	for g.Next() {
		cl := g.Str()
		clusters = append(clusters, cl)
		offsets = append(offsets, offsets[len(offsets)-1]+len(cl))
	}
	return clusters, offsets
}

// splitWordsPreserveNBSP splits by ASCII space and TAB,
// preserving NBSP (U+00A0) inside tokens and collapsing runs of separators.
func splitWordsPreserveNBSP(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	start := -1
	for i, r := range s {
		sep := r == ' ' || r == '\t'
		if sep {
			if start >= 0 {
				out = append(out, s[start:i])
				start = -1
			}
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		out = append(out, s[start:])
	}
	return out
}

// isWordBaseRune reports letters or digits only for base runes.
func isWordBaseRune(r rune) bool {
	if r <= 0 {
		return false
	}
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// lastBaseRune finds the last non-mark, non-format rune in s.
// It skips Mn/Me/Cf trailing marks to reach the base character.
func lastBaseRune(s string) rune {
	for len(s) > 0 {
		r, size := utf8.DecodeLastRuneInString(s)
		if r == utf8.RuneError && size == 1 {
			s = s[:len(s)-1]
			continue
		}
		if !(unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Cf, r)) {
			return r
		}
		s = s[:len(s)-size]
	}
	return -1
}

// firstBaseRune finds the first non-mark, non-format rune in s.
func firstBaseRune(s string) rune {
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError && size == 1 {
			s = s[size:]
			continue
		}
		if !(unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Cf, r)) {
			return r
		}
		s = s[size:]
	}
	return -1
}

// trimRightSpacesNBSP trims trailing ASCII spaces and NBSP.
func trimRightSpacesNBSP(s string) string {
	// Trim ASCII space.
	s = strings.TrimRight(s, " ")
	// Trim NBSP explicitly.
	for len(s) > 0 {
		r, size := utf8.DecodeLastRuneInString(s)
		if r == '\u00A0' {
			s = s[:len(s)-size]
			continue
		}
		break
	}
	return s
}
