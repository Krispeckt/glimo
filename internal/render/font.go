package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"strings"

	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const defaultDPI = 72

// Font wraps a TrueType font with pixel-accurate rendering helpers.
// It mimics CSS and Figma layout behavior for text measurement and positioning.
type Font struct {
	tt            *truetype.Font // underlying TrueType font
	sizePt        float64        // logical font size in points
	dpi           float64        // dots per inch scaling
	letterPercent float64        // tracking as percent of font size
	capRatio      float64        // fallback cap height ratio
}

// Loading

// LoadFont loads a .ttf file from disk and returns a Font object at the given point size.
// 1pt = 1/72 inch. Defaults to 72 DPI (1pt = 1px).
func LoadFont(path string, sizePt float64) (*Font, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadFontFromBytes(data, sizePt)
}

// LoadFontFromBytes parses a TrueType font from memory.
// Useful for embedding fonts or loading from resources.
func LoadFontFromBytes(data []byte, sizePt float64) (*Font, error) {
	ttf, err := truetype.Parse(data)
	if err != nil {
		return nil, err
	}
	f := &Font{
		tt:            ttf,
		dpi:           defaultDPI,
		letterPercent: 0.0,
		capRatio:      0.85,
	}
	return f.SetFontSizePt(sizePt), nil
}

// MustLoadFont loads a .ttf font from disk and panics on error.
// Intended for static initialization at package level.
func MustLoadFont(path string, sizePt float64) *Font {
	f, err := LoadFont(path, sizePt)
	if err != nil {
		panic(err)
	}
	return f
}

// MustLoadFontFromBytes parses a TrueType font from bytes and panics on error.
// Used for embedding fonts with Go’s //go:embed directive.
func MustLoadFontFromBytes(data []byte, sizePt float64) *Font {
	f, err := LoadFontFromBytes(data, sizePt)
	if err != nil {
		panic(err)
	}
	return f
}

// Configuration

// SetDPI sets the font’s DPI scaling. Defaults to 72 if <= 0.
// Higher DPI simulates scaled rendering for export or preview.
func (f *Font) SetDPI(dpi float64) *Font {
	if dpi <= 0 {
		dpi = defaultDPI
	}
	f.dpi = dpi
	return f
}

// SetFontSizePt sets the font size in points (1pt = 1/72 inch).
// Ensures a minimum value > 0 to avoid invalid scaling.
func (f *Font) SetFontSizePt(pt float64) *Font {
	if pt <= 0 {
		pt = 0.01
	}
	f.sizePt = pt
	return f
}

// SetLetterSpacingPercent defines tracking (letter spacing) as a percentage of font size.
// Positive values loosen spacing, negative values tighten spacing.
func (f *Font) SetLetterSpacingPercent(percent float64) *Font {
	f.letterPercent = percent
	return f
}

// Accessors

// HeightPt returns the font size in points.
func (f *Font) HeightPt() float64 { return f.sizePt }

// HeightPx returns the font size converted to pixels for the current DPI.
func (f *Font) HeightPx() float64 { return f.sizePt * f.dpi / 72.0 }

// DPI returns the current DPI value.
func (f *Font) DPI() float64 { return f.dpi }

// TrueTypeFont exposes the underlying truetype.Font instance.
func (f *Font) TrueTypeFont() *truetype.Font { return f.tt }

// cacheKey builds a unique cache key for font face reuse.
func (f *Font) cacheKey() string {
	return fmt.Sprintf("%p_%.3f_%.1f", f.tt, f.sizePt, f.dpi)
}

// Face caching

// Face returns a truetype.Face configured with the current size and DPI.
// Faces are cached to prevent redundant allocations and ensure consistent rendering.
func (f *Font) Face() font.Face {
	key := f.cacheKey()
	if face, ok := fontCache.get(key); ok {
		return face
	}
	face := truetype.NewFace(f.tt, &truetype.Options{
		Size:    f.sizePt,
		DPI:     f.dpi,
		Hinting: font.HintingNone,
	})
	fontCache.put(key, face)
	return face
}

// Metrics

// TrackingPx returns the tracking offset (in pixels) applied between glyphs.
func (f *Font) TrackingPx() float64 {
	return (f.letterPercent / 100.0) * f.HeightPx()
}

// AscentPx returns the ascent (distance from baseline to top) in pixels.
func (f *Font) AscentPx() float64 {
	m := f.Face().Metrics()
	return float64(m.Ascent >> 6)
}

// DescentPx returns the descent (distance from baseline to bottom) in pixels.
func (f *Font) DescentPx() float64 {
	m := f.Face().Metrics()
	return float64(m.Descent >> 6)
}

// LineHeightPx returns the total line height (ascent + descent + leading) in pixels.
func (f *Font) LineHeightPx() float64 {
	m := f.Face().Metrics()
	return float64(m.Height >> 6)
}

// LeadingPx returns the vertical leading (extra space between lines) in pixels.
func (f *Font) LeadingPx() float64 {
	m := f.Face().Metrics()
	return float64((m.Height - (m.Ascent + m.Descent)) >> 6)
}

// CapHeightPx estimates the visual cap height (“h” height) in pixels.
// Falls back to 85% of ascent if glyph metrics are unavailable.
func (f *Font) CapHeightPx() float64 {
	face := f.Face()
	if b, _, ok := face.GlyphBounds('h'); ok {
		h := float64(b.Max.Y-b.Min.Y) / 64.0
		if h > 0 {
			return h
		}
	}
	return f.AscentPx() * f.capRatio
}

// Layout helpers

// BaselineForTopY returns the baseline y coordinate for a given top y value.
// Matches CSS line box behavior: baseline = top + ascent + (leading / 2).
func (f *Font) BaselineForTopY(topY float64) float64 {
	a := f.AscentPx()
	leading := f.LeadingPx()
	return topY + a + leading/2
}

// TopYForBaseline returns the top y coordinate for a given baseline y value.
// Inverse of BaselineForTopY. Matches CSS vertical alignment model.
func (f *Font) TopYForBaseline(baselineY float64) float64 {
	a := f.AscentPx()
	d := f.DescentPx()
	lineHeight := f.LineHeightPx()
	leading := lineHeight - (a + d)
	return baselineY - a - leading/2
}

// Drawing

// DrawString draws a single line of text on the destination image.
// Tracking and kerning are applied between glyphs, not after the final one.
// The baseline is aligned to pixel grid to avoid blur.
func (f *Font) DrawString(dst draw.Image, col color.Color, s string, x, baselineY float64) fixed.Point26_6 {
	if s == "" {
		return fixed.Point26_6{X: geom.Fix(x), Y: geom.Fix(baselineY)}
	}
	face := f.Face()
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col),
		Face: face,
		Dot: fixed.Point26_6{
			X: geom.Fix(math.Round(x)),
			Y: geom.Fix(math.Round(baselineY)),
		},
	}
	track := geom.Fix(f.TrackingPx())
	runes := []rune(s)
	for i, r := range runes {
		d.DrawString(string(r))
		if i < len(runes)-1 {
			d.Dot.X += track
		}
	}
	return d.Dot
}

// Measurement

// MeasureString measures the pixel width and height of a single-line string.
// Width includes glyph advances and tracking between characters.
// Height equals the line height in pixels.
func (f *Font) MeasureString(s string) (w, h float64) {
	if s == "" {
		return 0, 0
	}
	face := f.Face()
	adv := font.MeasureString(face, s)
	w = float64(adv >> 6)
	runes := []rune(s)
	if len(runes) > 1 {
		w += float64(len(runes)-1) * f.TrackingPx()
	}
	h = f.LineHeightPx()
	return
}

// MeasureMultilineString measures a multi-line text block in pixels.
// Width = max line width. Height = number of lines × lineHeightPx.
// If lineHeightPx <= 0, the font’s intrinsic line height is used.
func (f *Font) MeasureMultilineString(s string, lineHeightPx float64) (width, height float64) {
	lines := strings.Split(s, "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return 0, 0
	}
	if lineHeightPx <= 0 {
		lineHeightPx = f.LineHeightPx()
	}
	for _, line := range lines {
		w, _ := f.MeasureString(line)
		if w > width {
			width = w
		}
	}
	height = float64(len(lines)) * lineHeightPx
	return
}
