package instructions

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/Krispeckt/glimo/internal/render"
)

// lineWidth returns the configured wrapping width in pixels.
// A zero value disables wrapping and allows unbounded line length.
func (t *Text) lineWidth() float64 {
	if t.maxWidth <= 0 {
		return 0
	}
	return t.maxWidth
}

// wrapTextScaled splits the text into lines based on newline characters
// and performs per-line wrapping with font scaling.
//
// Behavior:
//   - Each paragraph is split on '\n'.
//   - Lines are wrapped according to the current WrapMode.
//   - Per-line scaling via scaleStep is applied progressively.
//   - Ellipsis is appended if maxLines is reached.
//   - Empty lines are preserved as paragraph breaks.
func (t *Text) wrapTextScaled() []string {
	if t.maxWidth <= 0 {
		return strings.Split(t.text, "\n")
	}
	var out []string
	lineIdx := 0
	truncated := false

	flushAndMaybeTruncate := func() bool {
		if t.maxLines > 0 && len(out) > t.maxLines { // requires test coverage
			out = out[:t.maxLines]
			lastFont := t.fontForLine(t.maxLines - 1)
			out = appendEllipsisRunes(out, lastFont, t.maxWidth)
			truncated = true
			return true
		}
		return false
	}

	paras := strings.Split(t.text, "\n")
	for pi, p := range paras {
		if truncated {
			break
		}
		if p == "" {
			out = append(out, "")
			lineIdx++
			if flushAndMaybeTruncate() {
				break
			}
			continue
		}

		var sub []string
		if t.wrapMode == WrapBySymbol {
			sub = t.wrapParaBySymbolsScaled(p, &lineIdx)
		} else {
			sub = t.wrapParaByWordsScaled(p, &lineIdx)
		}

		for _, s := range sub {
			if truncated {
				break
			}
			out = append(out, s)
			if flushAndMaybeTruncate() {
				break
			}
		}

		if pi < len(paras)-1 && paras[pi+1] == "" && !truncated {
			out = append(out, "")
			lineIdx++
			_ = flushAndMaybeTruncate()
		}
	}
	return out
}

// wrapParaByWordsScaled performs wrapping at word boundaries using
// the configured maxWidth and per-line scaling.
//
// Long tokens that exceed width individually are broken using
// breakLongTokenWithFont to ensure readability.
func (t *Text) wrapParaByWordsScaled(p string, lineIdxPtr *int) []string {
	words := strings.Fields(p)
	if len(words) == 0 {
		*lineIdxPtr++
		return []string{""}
	}

	var lines []string
	i := 0
	for i < len(words) {
		lf := t.fontForLine(*lineIdxPtr)
		width := t.lineWidth()

		cur := words[i]
		for i+1 < len(words) {
			next := words[i+1]
			if w, _ := lf.MeasureString(cur + " " + next); w <= width {
				cur += " " + next
				i++
			} else {
				break
			}
		}

		if w0, _ := lf.MeasureString(cur); w0 > width {
			chunks := t.breakLongTokenWithFont(words[i], lf, width)
			for ci, c := range chunks {
				lines = append(lines, c)
				*lineIdxPtr++
				if ci < len(chunks)-1 {
					lf = t.fontForLine(*lineIdxPtr)
					width = t.lineWidth()
				}
			}
			i++
			continue
		}

		lines = append(lines, cur)
		*lineIdxPtr++
		i++
	}
	return lines
}

// wrapParaBySymbolsScaled wraps text by character (rune) count
// when word-level wrapping is not desired or possible.
//
// It appends a wrapSymbol (e.g. '-') at the break point
// if the next character continues the word.
func (t *Text) wrapParaBySymbolsScaled(p string, lineIdxPtr *int) []string {
	var lines []string
	runes := []rune(p)
	start := 0

	for start < len(runes) {
		lf := t.fontForLine(*lineIdxPtr)
		width := t.lineWidth()

		end := start
		for end < len(runes) {
			cand := string(runes[start : end+1])
			if w, _ := lf.MeasureString(cand); w <= width {
				end++
				continue
			}
			break
		}

		if end == start {
			lines = append(lines, string(runes[start]))
			*lineIdxPtr++
			start++
			continue
		}

		line := string(runes[start:end])
		if end < len(runes) && !strings.HasSuffix(line, " ") && t.wrapSymbol != "" {
			line += t.wrapSymbol
		}
		lines = append(lines, strings.TrimRight(line, " "))
		*lineIdxPtr++
		start = end
	}
	return lines
}

// breakLongTokenWithFont splits a single overlong word or token
// into smaller segments that fit within maxWidth,
// optionally appending a wrapSymbol at each break.
func (t *Text) breakLongTokenWithFont(token string, lf *render.Font, width float64) []string {
	var out []string
	var buf string
	for _, r := range token {
		cand := buf + string(r)
		if w, _ := lf.MeasureString(cand); w <= width {
			buf = cand
			continue
		}
		if buf != "" && t.wrapSymbol != "" {
			buf += t.wrapSymbol
		}
		out = append(out, buf)
		buf = string(r)
	}
	if buf != "" {
		out = append(out, buf)
	}
	return out
}

// autoSpacing estimates inter-line spacing multiplier based on
// average fill ratio, font scaling, and adaptive density heuristics.
//
// It dynamically balances compactness and legibility depending
// on how much text fills the maximum width and how the font size scales.
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
	for i, s := range lines {
		lf := t.fontForLine(i)
		if w, _ := lf.MeasureString(s); !math.IsNaN(w) {
			totalWidth += math.Max(w, 0)
		}
		if curPt := lf.HeightPt(); curPt > 0 {
			totalScale += curPt / basePt
		}
	}

	n := float64(len(lines))
	if n == 0 {
		return 1.0
	}

	fill := geom.ClampF64(totalWidth/(t.maxWidth*n), fillMin, fillMax)
	fillT := (fill - fillMin) / (fillMax - fillMin)
	base := geom.Lerp(baseSparse, baseDense, fillT)

	avgScale := geom.ClampF64(totalScale/n, scaleMin, scaleMax)
	atten := geom.ClampF64(1.0-scaleWeight*(1.0-avgScale), attenMin, attenMax)

	return geom.ClampF64(base*atten, spacingMin, spacingMax)
}

// appendEllipsisRunes trims the final line of text to fit an ellipsis character
// ("…") within maxWidth, preserving UTF-8 correctness when slicing runes.
//
// It ensures the ellipsis appears even when truncation removes the entire line.
func appendEllipsisRunes(lines []string, f *render.Font, maxWidth float64) []string {
	const ellipsis = "…"
	if len(lines) == 0 {
		return lines
	}

	lastIdx := len(lines) - 1
	last := lines[lastIdx]

	if w, _ := f.MeasureString(last + ellipsis); w <= maxWidth {
		lines[lastIdx] = last + ellipsis
		return lines
	}

	for last != "" {
		_, size := utf8.DecodeLastRuneInString(last)
		if size <= 0 {
			break
		}
		last = last[:len(last)-size]
		if w, _ := f.MeasureString(last + ellipsis); w <= maxWidth {
			lines[lastIdx] = last + ellipsis
			return lines
		}
	}

	if w, _ := f.MeasureString(ellipsis); w <= maxWidth {
		lines[lastIdx] = ellipsis
	}
	return lines
}
