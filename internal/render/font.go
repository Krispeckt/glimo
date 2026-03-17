package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/Krispeckt/glimo/internal/core/geom"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const defaultDPI = 72

// VariationAxis holds the tag and value for a single variable-font design axis.
//
// Common tags (4 ASCII chars):
//
//	"wght" — weight   (e.g. 100–900)
//	"wdth" — width    (e.g. 50–200)
//	"ital" — italic   (0 or 1)
//	"slnt" — slant    (negative = italic lean)
//	"opsz" — optical size
type VariationAxis struct {
	Tag   string  // 4-character OpenType axis tag
	Value float32 // requested axis value (clamped to font min/max at face creation)
}

// AvailableAxis describes a design axis as declared in the font file.
type AvailableAxis struct {
	Tag     string  // 4-character OpenType tag
	Min     float32 // minimum allowed value
	Default float32 // default value when axis is not overridden
	Max     float32 // maximum allowed value
}

// Font wraps an OpenType/TrueType font (including variable fonts) with
// pixel-accurate rendering helpers that match Figma and CSS positioning semantics.
//
// # Coordinate model
//
// All Y coordinates use the "visual-top" model:
//
//	y = baseline − ascent   (top of the tallest glyph, no leading offset)
//
// This eliminates the leading-induced drift that occurs when the classic
// "line-box top" model is used with fonts that have non-zero internal leading.
//
// # Variable fonts
//
// Use SetAxis to configure design axes before rendering. The axis values are
// reflected in the font metrics and glyph outlines when the underlying
// golang.org/x/image/font/opentype implementation supports them.
// Use AvailableAxes to discover what axes a font exposes.
//
// # Thread safety
//
// A Font value must not be shared between goroutines while being mutated
// (SetFontSizePt, SetDPI, SetLetterSpacingPercent, SetAxis). After all
// configuration is done, the value is safe to use concurrently for
// rendering from multiple goroutines.
type Font struct {
	sf        *opentype.Font // underlying parsed font (opentype.Font ≡ sfnt.Font)
	sizePt    float64        // logical size in points
	dpi       float64        // dots per inch (default 72 → 1pt = 1px)
	letterPct float64        // tracking as % of font size (positive = looser)
	axisMu    sync.RWMutex
	axes      []VariationAxis // user-set variable-font axes (nil for static fonts)
}

// Loading

// LoadFont loads a font file from disk.
// Accepts TrueType (.ttf), OpenType (.otf), and variable fonts in either format.
func LoadFont(path string, sizePt float64) (*Font, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("glimo/render: open font %q: %w", path, err)
	}
	return LoadFontFromBytes(data, sizePt)
}

// LoadFontFromBytes parses a font from a byte slice.
// Accepts TrueType, OpenType, and variable fonts.
func LoadFontFromBytes(data []byte, sizePt float64) (*Font, error) {
	sf, err := opentype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("glimo/render: parse font: %w", err)
	}
	f := &Font{sf: sf, dpi: defaultDPI}
	return f.SetFontSizePt(sizePt), nil
}

// MustLoadFont loads a font from disk and panics on failure.
func MustLoadFont(path string, sizePt float64) *Font {
	f, err := LoadFont(path, sizePt)
	if err != nil {
		panic(err)
	}
	return f
}

// MustLoadFontFromBytes parses a font from memory and panics on failure.
func MustLoadFontFromBytes(data []byte, sizePt float64) *Font {
	f, err := LoadFontFromBytes(data, sizePt)
	if err != nil {
		panic(err)
	}
	return f
}

// Configuration

// SetDPI sets the dots-per-inch scaling factor (default 72).
// At 72 DPI, 1pt = 1px; at 96 DPI, 1pt ≈ 1.33px.
func (f *Font) SetDPI(dpi float64) *Font {
	if dpi <= 0 {
		dpi = defaultDPI
	}
	f.dpi = dpi
	return f
}

// SetFontSizePt sets the logical font size in points (minimum 0.01).
func (f *Font) SetFontSizePt(pt float64) *Font {
	if pt <= 0 {
		pt = 0.01
	}
	f.sizePt = pt
	return f
}

// SetLetterSpacingPercent sets additional tracking as a percentage of the font
// size. Positive values loosen; negative values tighten letter spacing.
func (f *Font) SetLetterSpacingPercent(percent float64) *Font {
	f.letterPct = percent
	return f
}

// Clone returns a new *Font with the same configuration as f.
// The axes slice is deep-copied. The new Font has an independent mutex.
func (f *Font) Clone() *Font {
	f.axisMu.RLock()
	axesCopy := make([]VariationAxis, len(f.axes))
	copy(axesCopy, f.axes)
	f.axisMu.RUnlock()
	return &Font{
		sf:        f.sf,
		sizePt:    f.sizePt,
		dpi:       f.dpi,
		letterPct: f.letterPct,
		axes:      axesCopy,
	}
}

// SetAxis sets a variable-font design axis by its 4-character OpenType tag.
// Returns the Font for method chaining.
//
// Example:
//
//	font.SetAxis("wght", 700).SetAxis("wdth", 120)
func (f *Font) SetAxis(tag string, value float32) *Font {
	f.axisMu.Lock()
	defer f.axisMu.Unlock()
	for i, a := range f.axes {
		if a.Tag == tag {
			f.axes[i].Value = value
			return f
		}
	}
	f.axes = append(f.axes, VariationAxis{Tag: tag, Value: value})
	return f
}

// Accessors

// HeightPt returns the font size in points.
func (f *Font) HeightPt() float64 { return f.sizePt }

// HeightPx returns the font size converted to pixels at the current DPI.
func (f *Font) HeightPx() float64 { return f.sizePt * f.dpi / 72.0 }

// DPI returns the current DPI value.
func (f *Font) DPI() float64 { return f.dpi }

// OpenTypeFont returns the underlying parsed font.
// The returned value is safe to use concurrently for metric queries.
func (f *Font) OpenTypeFont() *opentype.Font { return f.sf }

// Axes returns a snapshot of the currently configured variable-font axes.
// Returns nil if no axes have been set.
func (f *Font) Axes() []VariationAxis {
	f.axisMu.RLock()
	defer f.axisMu.RUnlock()
	if len(f.axes) == 0 {
		return nil
	}
	cp := make([]VariationAxis, len(f.axes))
	copy(cp, f.axes)
	return cp
}

// AvailableAxes returns the design axes defined in the font file.
//
// Note: the current golang.org/x/image/font/sfnt backend does not expose fvar
// table parsing, so this always returns nil. Axis values set via SetAxis are
// stored but do not affect glyph rendering until a backend with variable-font
// support is wired in.
func (f *Font) AvailableAxes() []AvailableAxis { return nil }

// IsVariable reports whether the font contains variable-font axes (fvar table).
//
// Note: returns false because the current backend does not parse the fvar table.
func (f *Font) IsVariable() bool { return false }

//  Face caching

// cacheKey returns a stable string that uniquely identifies this font
// configuration (pointer, size, dpi, tracking, sorted axes).
func (f *Font) cacheKey() string {
	f.axisMu.RLock()
	axesCopy := make([]VariationAxis, len(f.axes))
	copy(axesCopy, f.axes)
	f.axisMu.RUnlock()

	// Sort axes by tag so the key is deterministic regardless of insertion order.
	sort.Slice(axesCopy, func(i, j int) bool { return axesCopy[i].Tag < axesCopy[j].Tag })

	var sb strings.Builder
	fmt.Fprintf(&sb, "%p|%.4f|%.2f|%.4f", f.sf, f.sizePt, f.dpi, f.letterPct)
	for _, a := range axesCopy {
		fmt.Fprintf(&sb, "|%s:%.4f", a.Tag, a.Value)
	}
	return sb.String()
}

// Face returns a cached font.Face for the current size, DPI, and axes.
// The returned face is safe for concurrent use.
func (f *Font) Face() font.Face {
	key := f.cacheKey()
	if face, ok := getFontCache().get(key); ok {
		return face
	}

	face, err := opentype.NewFace(f.sf, &opentype.FaceOptions{
		Size:    f.sizePt,
		DPI:     f.dpi,
		Hinting: font.HintingNone,
	})
	if err != nil {
		// LoadFontFromBytes validates the font, so this should never happen.
		panic(fmt.Sprintf("glimo/render: create font face: %v", err))
	}
	getFontCache().put(key, face)
	return face
}

// Metrics

// TrackingPx returns the per-character additional spacing in pixels.
func (f *Font) TrackingPx() float64 {
	return (f.letterPct / 100.0) * f.HeightPx()
}

// AscentPx returns the ascent (baseline to top of tallest glyph) in pixels.
// Sub-pixel precision is preserved; round only when an integer is required.
func (f *Font) AscentPx() float64 {
	return float64(f.Face().Metrics().Ascent) / 64.0
}

// DescentPx returns the descent (baseline to bottom of lowest glyph) in pixels.
// This value is positive even though glyphs extend below the baseline.
func (f *Font) DescentPx() float64 {
	return float64(f.Face().Metrics().Descent) / 64.0
}

// LineHeightPx returns the full typographic line height (ascent + descent + leading)
// in pixels. This is the baseline-to-baseline advance for single-spacing (1×).
func (f *Font) LineHeightPx() float64 {
	return float64(f.Face().Metrics().Height) / 64.0
}

// LeadingPx returns the internal leading (extra vertical space above/below the
// ascent/descent within a line box) in pixels.
func (f *Font) LeadingPx() float64 {
	m := f.Face().Metrics()
	h := float64(m.Height) / 64.0
	a := float64(m.Ascent) / 64.0
	d := float64(m.Descent) / 64.0
	return h - a - d
}

// CapHeightPx returns the capital-letter height in pixels, measured from
// the baseline to the top of 'H'. Falls back to 70 % of ascent.
func (f *Font) CapHeightPx() float64 {
	if b, _, ok := f.Face().GlyphBounds('H'); ok {
		if h := float64(b.Max.Y-b.Min.Y) / 64.0; h > 0 {
			return h
		}
	}
	return f.AscentPx() * 0.70
}

// XHeightPx returns the x-height (top of lowercase 'x') in pixels.
// Falls back to 52 % of ascent.
func (f *Font) XHeightPx() float64 {
	if b, _, ok := f.Face().GlyphBounds('x'); ok {
		if h := float64(b.Max.Y-b.Min.Y) / 64.0; h > 0 {
			return h
		}
	}
	return f.AscentPx() * 0.52
}

// Layout helpers

// BaselineFromVisualTop converts a visual-top Y coordinate to the baseline Y.
//
//	baseline = visualTop + ascent
//
// This is the correct helper for the visual-top coordinate model used by
// all glimo text primitives.
func (f *Font) BaselineFromVisualTop(visualTopY float64) float64 {
	return visualTopY + f.AscentPx()
}

// VisualTopFromBaseline converts a baseline Y coordinate to visual-top Y.
//
//	visualTop = baseline − ascent
func (f *Font) VisualTopFromBaseline(baselineY float64) float64 {
	return baselineY - f.AscentPx()
}

// BaselineForTopY converts a line-box top Y to a baseline Y using the CSS model
// (includes half-leading above the ascent).
//
// Deprecated: Use BaselineFromVisualTop, which uses the drift-free visual-top model.
func (f *Font) BaselineForTopY(topY float64) float64 {
	return topY + f.AscentPx() + f.LeadingPx()/2
}

// TopYForBaseline converts a baseline Y to a line-box top Y (CSS model).
//
// Deprecated: Use VisualTopFromBaseline.
func (f *Font) TopYForBaseline(baselineY float64) float64 {
	return baselineY - f.AscentPx() - f.LeadingPx()/2
}

// Drawing

// DrawString renders s at the given position.
//
//   - x is the sub-pixel left edge.
//   - baselineY is snapped to the nearest integer for sharp horizontal stems.
//
// When tracking is zero the entire string is drawn in one call, preserving
// kerning pairs. When tracking is non-zero glyphs are drawn individually
// with the additional spacing inserted between them.
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
			X: geom.Fix(x),
			Y: geom.Fix(math.Round(baselineY)), // integer Y → sharp baseline
		},
	}
	if f.letterPct == 0 {
		// Fast path: one call, full kerning.
		d.DrawString(s)
	} else {
		track := geom.Fix(f.TrackingPx())
		runes := []rune(s)
		for i, r := range runes {
			d.DrawString(string(r))
			if i < len(runes)-1 {
				d.Dot.X += track
			}
		}
	}
	return d.Dot
}

// Measurement

// MeasureString measures the rendered width and height of a single-line string.
//
// Width includes glyph advances, kerning, and tracking.
// Height equals LineHeightPx (the full typographic line height).
func (f *Font) MeasureString(s string) (w, h float64) {
	if s == "" {
		return 0, 0
	}
	face := f.Face()
	adv := font.MeasureString(face, s)
	// Use sub-pixel division (not >> 6) to preserve fractional-pixel precision.
	w = float64(adv) / 64.0
	if f.letterPct != 0 {
		runes := []rune(s)
		if len(runes) > 1 {
			w += float64(len(runes)-1) * f.TrackingPx()
		}
	}
	h = f.LineHeightPx()
	return
}

// MeasureMultilineString measures a multi-line string split by "\n".
// Width = widest line; Height = N × lineHeightPx.
func (f *Font) MeasureMultilineString(s string, lineHeightPx float64) (width, height float64) {
	lines := strings.Split(s, "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return 0, 0
	}
	if lineHeightPx <= 0 {
		lineHeightPx = f.LineHeightPx()
	}
	for _, line := range lines {
		if w, _ := f.MeasureString(line); w > width {
			width = w
		}
	}
	height = float64(len(lines)) * lineHeightPx
	return
}
