package instructions

import (
	"image"
	"math"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/effects"
	"github.com/Krispeckt/glimo/internal/containers"
	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
	"github.com/Krispeckt/glimo/internal/render"
	xdraw "golang.org/x/image/draw"
)

// WrapMode defines how text lines are broken when exceeding maximum width.
type WrapMode int

const (
	// WrapByWord breaks lines at whitespace boundaries only.
	WrapByWord WrapMode = iota
	// WrapBySymbol breaks lines at character level and optionally inserts a hyphenation symbol.
	WrapBySymbol
)

// AlignText defines the horizontal alignment behavior of rendered lines.
type AlignText int

const (
	// AlignTextLeft aligns text to the left edge.
	AlignTextLeft AlignText = iota
	// AlignTextCenter centers text horizontally within the line box.
	AlignTextCenter
	// AlignTextRight aligns text to the right edge.
	AlignTextRight
	// todo: AlignTextJustify — not implemented yet.
)

// Text describes a multi-line text block capable of wrapping, aligning,
// scaling, and applying stroke/fill effects.
//
// Features include:
//   - Word or symbol wrapping with optional hyphenation.
//   - Left/center/right alignment within fixed width or anchor-based layout.
//   - Progressive per-line scaling for dynamic typography.
//   - Automatic or manual line spacing.
//   - Pattern or gradient fill based on canvas coordinates.
//   - Morphological stroke expansion using alpha dilation.
//   - Pre- and post-processing effects via a flexible effect container.
type Text struct {
	text         string
	x, y         float64
	font         *render.Font
	colorPattern patterns.Pattern
	maxWidth     float64
	lineSpacing  float64
	align        AlignText
	wrapMode     WrapMode
	wrapSymbol   string
	maxLines     int
	scaleStep    float64

	strokePatternColor patterns.Pattern
	strokeWidth        float64

	effects containers.Effects
}

// NewText constructs a Text instance with default configuration.
// Defaults:
//   - WrapByWord and AlignTextLeft
//   - No stroke or fill pattern
//   - lineSpacing = 0 → automatic by font metrics
//   - maxWidth = 0 → no wrapping, anchor-based alignment
func NewText(text string, x, y float64, font *render.Font) *Text {
	return &Text{
		text:               text,
		x:                  x,
		y:                  y,
		font:               font,
		colorPattern:       nil,
		align:              AlignTextLeft,
		wrapMode:           WrapByWord,
		wrapSymbol:         "-",
		maxWidth:           0,
		lineSpacing:        0,
		maxLines:           0,
		scaleStep:          0,
		strokePatternColor: nil,
		strokeWidth:        0,
		effects:            containers.Effects{},
	}
}

// SetAlign configures horizontal line alignment.
func (t *Text) SetAlign(a AlignText) *Text {
	t.align = a
	return t
}

// SetLineSpacing defines custom spacing as a percentage of line height.
func (t *Text) SetLineSpacing(percent float64) *Text {
	t.lineSpacing = percent / 100.0
	return t
}

// SetWrap sets wrapping mode and hyphenation symbol.
func (t *Text) SetWrap(mode WrapMode, symbol string) *Text {
	t.wrapMode = mode
	t.SetWrapSymbol(symbol)
	return t
}

// SetWrapMode changes wrapping behavior without modifying the hyphenation symbol.
func (t *Text) SetWrapMode(mode WrapMode) *Text {
	t.wrapMode = mode
	return t
}

// SetWrapSymbol defines a custom symbol used during symbol-based wrapping.
// Falls back to "-" when an empty string is provided.
func (t *Text) SetWrapSymbol(sym string) *Text {
	if sym == "" {
		t.wrapSymbol = "-"
	} else {
		t.wrapSymbol = sym
	}
	return t
}

// SetMaxWidth limits the maximum text box width in pixels.
// A value of 0 disables wrapping and aligns relative to anchor coordinates.
func (t *Text) SetMaxWidth(w float64) *Text {
	t.maxWidth = math.Max(w, 0)
	return t
}

// SetMaxLines limits the number of rendered lines. Zero means no limit.
func (t *Text) SetMaxLines(n int) *Text {
	t.maxLines = n
	return t
}

// SetScaleStep applies a per-line font size increment (in points).
// Positive values enlarge text progressively; negative values shrink it.
func (t *Text) SetScaleStep(pt float64) *Text {
	t.scaleStep = pt
	return t
}

// SetStrokeWithPattern defines a stroke using a color or gradient pattern.
func (t *Text) SetStrokeWithPattern(p patterns.Pattern, width float64) *Text {
	t.strokePatternColor = p
	t.strokeWidth = math.Max(width, 0)
	return t
}

// SetStrokeWithColor defines a stroke using a solid color pattern.
func (t *Text) SetStrokeWithColor(c patterns.Color, width float64) *Text {
	t.strokePatternColor = c.MakeSolidPattern()
	t.strokeWidth = math.Max(width, 0)
	return t
}

// SetColorPattern applies a pattern or gradient fill for text glyphs.
func (t *Text) SetColorPattern(p patterns.Pattern) *Text {
	t.colorPattern = p
	return t
}

// SetSolidColor applies a uniform color fill by generating a solid pattern.
func (t *Text) SetSolidColor(c patterns.Color) *Text {
	t.colorPattern = c.MakeSolidPattern()
	return t
}

// AddEffect adds a single post-processing effect to the text rendering pipeline.
func (t *Text) AddEffect(e effects.Effect) *Text {
	t.effects.Add(e)
	return t
}

// AddEffects appends multiple effects to the rendering pipeline in order.
func (t *Text) AddEffects(es ...effects.Effect) *Text {
	t.effects.AddList(es)
	return t
}

// SetPosition updates the anchor coordinates of the text block.
func (t *Text) SetPosition(x, y int) {
	t.x, t.y = float64(x), float64(y)
}

// Position returns the integer coordinates where the text block originates.
func (t *Text) Position() (int, int) { return int(t.x), int(t.y) }

// Size computes the bounding box of the rendered text.
// Returns zero if text or font is undefined.
func (t *Text) Size() *geom.Size {
	if t.font == nil || t.text == "" {
		return geom.NewSize(0, 0)
	}

	lines := t.wrapTextScaled()
	if len(lines) == 0 {
		return geom.NewSize(0, 0)
	}

	spacing := t.lineSpacing
	if spacing <= 0 {
		spacing = t.autoSpacing(lines)
	}

	var maxLineWidth float64
	for i, line := range lines {
		w, _ := t.fontForLine(i).MeasureString(line)
		if w > maxLineWidth {
			maxLineWidth = w
		}
	}

	lineHeight := t.fontForLine(0).LineHeightPx()
	totalHeight := lineHeight*float64(len(lines)) +
		lineHeight*spacing*float64(len(lines)-1)

	width := t.maxWidth
	if width <= 0 {
		width = maxLineWidth
	}

	return geom.NewSize(width, totalHeight)
}

// Draw renders the text block into the given base and overlay images.
// The method performs optional stroke, fill, and post-processing effects.
func (t *Text) Draw(base, overlay *image.RGBA) {
	if t.font == nil || t.text == "" {
		return
	}

	lines := t.wrapTextScaled()
	spacing := t.lineSpacing
	if spacing <= 0 {
		spacing = t.autoSpacing(lines)
	}

	t.effects.PreApplyAll(overlay)

	yTop := t.y
	for i, line := range lines {
		if t.maxLines > 0 && i >= t.maxLines {
			break
		}
		lineFont := t.fontForLine(i)
		w, _ := lineFont.MeasureString(line)
		x := t.alignX(t.x, w, t.align)

		if t.strokePatternColor != nil && t.strokeWidth > 0 {
			t.drawStroke(base, overlay, lineFont, line, x, yTop)
		}
		t.drawProcess(base, overlay, lineFont, line, x, yTop, t.colorPattern)

		yTop += lineFont.LineHeightPx() * spacing
	}

	t.effects.PostApplyAll(overlay)
}

// ssScale determines supersampling factor based on font pixel height.
// Small fonts get higher supersampling to minimize aliasing.
func ssScale(px float64) int {
	if px < 12 {
		return 3
	}
	if px < 16 {
		return 2
	}
	return 1
}

// drawStroke rasterizes text glyphs, applies morphological dilation to create
// an outline, and composites it onto the destination using the configured stroke pattern.
//
// The stroke is computed in supersampled space for accuracy when required.
func (t *Text) drawStroke(base, overlay *image.RGBA, fnt *render.Font, s string, x, topY float64) {
	if s == "" || t.strokePatternColor == nil || t.strokeWidth <= 0 {
		return
	}

	yq := geom.Quant64(topY)
	xq := geom.Quant64(x)
	scale := ssScale(fnt.HeightPx())

	// Prepare font at working scale.
	ff := *fnt
	if scale > 1 {
		ff.SetFontSizePt(fnt.HeightPt() * float64(scale))
	}

	// Rasterize glyph masks.
	maskBig, maskSmall, bw, bh, dw, dh := rasterizeGlyphMasks(&ff, s, scale)
	if bw <= 0 || bh <= 0 {
		return
	}

	// Integer radius in destination pixels.
	r := t.safeRadius()

	// High-res dilation path for supersampling.
	if scale > 1 {
		radHR := t.safeRadius() * scale
		strokeHR := dilateAlphaDisk(maskBig, radHR)
		subtractInnerMask(strokeHR, maskBig, radHR)
		// Downscale HR stroke to destination size (dw+2r, dh+2r).
		strokeMask := image.NewRGBA(image.Rect(0, 0, dw+2*r, dh+2*r))
		xdraw.BiLinear.Scale(strokeMask, strokeMask.Bounds(), strokeHR, strokeHR.Bounds(), xdraw.Over, nil)

		xi := int(math.Floor(xq)) - r
		yi := int(math.Floor(yq)) - r
		dstRect := image.Rect(xi, yi, xi+strokeMask.Bounds().Dx(), yi+strokeMask.Bounds().Dy())
		compositePatternWithMask(base, overlay, strokeMask, xi, yi, dstRect, t.strokePatternColor)
		return
	}

	// No supersampling: dilate in destination space.
	strokeMask := dilateAlphaDisk(maskSmall, r)
	subtractInnerMask(strokeMask, maskSmall, r)

	xi := int(math.Floor(xq)) - r
	yi := int(math.Floor(yq)) - r
	dstRect := image.Rect(xi, yi, xi+strokeMask.Bounds().Dx(), yi+strokeMask.Bounds().Dy())
	compositePatternWithMask(base, overlay, strokeMask, xi, yi, dstRect, t.strokePatternColor)
}

// drawProcess rasterizes glyphs and composites the fill pattern using alpha coverage.
func (t *Text) drawProcess(base, overlay *image.RGBA, fnt *render.Font, s string, x, topY float64, p patterns.Pattern) {
	if s == "" || p == nil {
		return
	}

	yq := geom.Quant64(topY)
	xq := geom.Quant64(x)
	scale := ssScale(fnt.HeightPx())

	// Prepare font at working scale.
	ff := *fnt
	if scale > 1 {
		ff.SetFontSizePt(fnt.HeightPt() * float64(scale))
	}

	// Rasterize glyph masks.
	_, maskSmall, bw, bh, dw, dh := rasterizeGlyphMasks(&ff, s, scale)
	if bw <= 0 || bh <= 0 {
		return
	}

	xi := int(math.Floor(xq))
	yi := int(math.Floor(yq))
	dstRect := image.Rect(xi, yi, xi+dw, yi+dh)
	compositePatternWithMask(base, overlay, maskSmall, xi, yi, dstRect, p)
}

// alignX computes the horizontal anchor for a line based on alignment and width constraints.
// When maxWidth > 0, alignment occurs within a fixed box. Otherwise, alignment is relative to the anchor.
func (t *Text) alignX(anchorX, lineWidth float64, align AlignText) float64 {
	if t.maxWidth > 0 {
		switch align {
		case AlignTextCenter:
			return anchorX + (t.maxWidth-lineWidth)/2
		case AlignTextRight:
			return anchorX + (t.maxWidth - lineWidth)
		default:
			return anchorX
		}
	}
	switch align {
	case AlignTextCenter:
		return anchorX - lineWidth/2
	case AlignTextRight:
		return anchorX - lineWidth
	default:
		return anchorX
	}
}

// fontForLine returns a new font instance scaled per line index
// according to the configured scaleStep.
func (t *Text) fontForLine(lineIdx int) *render.Font {
	fc := *t.font
	if lineIdx == 0 || t.scaleStep == 0 {
		return &fc
	}
	newPt := t.font.HeightPt() + t.scaleStep*float64(lineIdx)
	if newPt < 1 {
		newPt = 1
	}
	fc.SetFontSizePt(newPt)
	return &fc
}

// safeRadius returns a non-zero integer radius derived from stroke width.
func (t *Text) safeRadius() int {
	return int(math.Max(math.Round(t.strokeWidth), 1))
}

// rasterizeGlyphMasks draws glyphs into an alpha mask at optional supersampled resolution.
// It returns both high- and low-resolution masks for stroke and fill processing.
func rasterizeGlyphMasks(ff *render.Font, s string, scale int) (maskBig *image.RGBA, maskSmall *image.RGBA, bw, bh, dw, dh int) {
	w, h := ff.MeasureString(s)
	descent := ff.DescentPx()
	bw, bh = int(math.Ceil(w)), int(math.Ceil(h+descent))
	if bw <= 0 || bh <= 0 {
		return nil, nil, bw, bh, 0, 0
	}

	maskBig = image.NewRGBA(image.Rect(0, 0, bw, bh))
	baselineY := math.Round(ff.BaselineForTopY(0))
	_ = ff.DrawString(maskBig, colors.Black, s, 0, baselineY)

	if scale == 1 {
		return maskBig, maskBig, bw, bh, bw, bh
	}

	dw = int(math.Max(math.Round(float64(bw)/float64(scale)), 1))
	dh = int(math.Max(math.Round(float64(bh)/float64(scale)), 1))
	maskSmall = image.NewRGBA(image.Rect(0, 0, dw, dh))
	// BiLinear is smoother for alpha masks than CatmullRom here.
	xdraw.BiLinear.Scale(maskSmall, maskSmall.Bounds(), maskBig, maskBig.Bounds(), xdraw.Over, nil)
	return maskBig, maskSmall, bw, bh, dw, dh
}

// dilateAlphaDisk performs alpha-based morphological expansion by a circular kernel.
// It is used to create thickened alpha masks representing stroke outlines.
func dilateAlphaDisk(src *image.RGBA, r int) *image.RGBA {
	if r <= 0 {
		sb := src.Bounds()
		dst := image.NewRGBA(image.Rect(0, 0, sb.Dx(), sb.Dy()))
		copy(dst.Pix, src.Pix)
		return dst
	}

	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, sw+2*r, sh+2*r))
	dw := dst.Bounds().Dx()

	type span struct{ dy, dx0, dx1 int }
	spans := make([]span, 2*r+1)
	for oy := -r; oy <= r; oy++ {
		mx := int(math.Floor(math.Sqrt(float64(r*r - oy*oy))))
		spans[oy+r] = span{dy: oy, dx0: -mx, dx1: mx}
	}

	for y := 0; y < sh; y++ {
		srow := src.Pix[y*src.Stride : y*src.Stride+sw*4]

		for x := 0; x < sw; x++ {
			sa := srow[x*4+3]
			if sa == 0 {
				continue
			}

			for _, sp := range spans {
				dy := y + sp.dy + r
				drow := dst.Pix[dy*dst.Stride : dy*dst.Stride+dw*4]

				start := x + sp.dx0 + r
				end := x + sp.dx1 + r

				i := start*4 + 3
				if sa == 255 {
					for dx := start; dx <= end; dx++ {
						drow[i] = 255
						i += 4
					}
				} else {
					a := sa
					for dx := start; dx <= end; dx++ {
						if a > drow[i] {
							drow[i] = a
						}
						i += 4
					}
				}
			}
		}
	}

	return dst
}

// subtractInnerMask removes the original glyph area from a dilated mask,
// leaving only the outer rim for stroke rendering.
func subtractInnerMask(expanded *image.RGBA, src *image.RGBA, r int) {
	sb := src.Bounds()
	W := expanded.Bounds().Dx()

	for y := 0; y < sb.Dy(); y++ {
		srow := src.Pix[y*src.Stride : y*src.Stride+sb.Dx()*4]
		drow := expanded.Pix[(y+r)*expanded.Stride : (y+r)*expanded.Stride+W*4]
		for x2 := 0; x2 < sb.Dx(); x2++ {
			sa := srow[x2*4+3]
			if sa == 0 {
				continue
			}
			ai := &drow[(x2+r)*4+3]
			if int(*ai) <= int(sa) {
				*ai = 0
			} else {
				*ai = byte(int(*ai) - int(sa))
			}
		}
	}
}
