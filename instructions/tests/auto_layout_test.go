package glimo_test

import (
	"image"
	"testing"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/instructions"
	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/stretchr/testify/require"
)

/*
mockShape implements BoundedShape for precise layout verification.
We expose x,y,w,h and count draw calls. The layout engine will call SetBounds,
so w/h reflect post-layout sizes after grow/shrink/stretch.
*/
type mockShape struct {
	name      string
	x, y      int
	w, h      int
	drawCalls int
}

func newMock(name string, w, h int) *mockShape { return &mockShape{name: name, w: w, h: h} }

func (m *mockShape) Draw(_, _ *image.RGBA) { m.drawCalls++ }
func (m *mockShape) Position() (int, int)  { return m.x, m.y }
func (m *mockShape) SetPosition(x, y int)  { m.x, m.y = x, y }
func (m *mockShape) Size() *geom.Size      { return geom.NewSize(float64(m.w), float64(m.h)) }
func (m *mockShape) SetBounds(x, y, w, h int) {
	m.x, m.y, m.w, m.h = x, y, w, h
}
func (m *mockShape) SetSize(w, h int) { m.w, m.h = w, h }

// canvas helpers
func newCanvases() (*image.RGBA, *image.RGBA) {
	base := image.NewRGBA(image.Rect(0, 0, 800, 600))
	overlay := image.NewRGBA(image.Rect(0, 0, 800, 600))
	return base, overlay
}

type itemCase struct {
	name    string
	w, h    int
	style   instructions.ItemStyle
	expectX *int
	expectY *int
	expectW *int
	expectH *int
}

type testCase struct {
	name             string
	originX, originY int
	style            instructions.ContainerStyle
	items            []itemCase
	expectOuterW     *float64
	expectOuterH     *float64
}

/*
TestAutoLayout_Cases
Single entry point that runs a table of layout edge-cases.
All comments are American English and include exact formulas.
*/
func TestAutoLayout_Cases(t *testing.T) {
	intp := func(v int) *int { return &v }
	floatp := func(v float64) *float64 { return &v }

	cases := []testCase{
		{
			name:    "row_wrap_aligncontent_stretch_lines_only",
			originX: 10, originY: 20,
			// Row + Wrap + AlignContent=Stretch with fixed height.
			// innerH = Height - (pt+pb) = 120 - (5+5) = 110
			// 3 items with heights 20,20,20; lines → [a] and [b,c]
			// totalCross = cross(line1)+gapY+cross(line2) = 20 + 10 + 20 = 50
			// leftover = 110 - 50 = 60  → extraPerLine = 60/2 = 30
			// line1 top = 20 + 5 = 25
			// line2 top = 25 + (20+30) + 10 = 85
			style: instructions.ContainerStyle{
				Display:      instructions.DisplayFlex,
				Direction:    instructions.Row,
				Wrap:         true,
				Padding:      [4]int{5, 5, 5, 5},
				Gap:          instructions.Vector2{X: 10, Y: 10},
				Justify:      instructions.JustifyStart,
				AlignItems:   instructions.AlignItemsStart,   // items keep own height
				AlignContent: instructions.AlignItemsStretch, // lines get stretched
				Width:        140,
				Height:       120,
			},
			items: []itemCase{
				{name: "a", w: 70, h: 20, style: instructions.ItemStyle{}, expectY: intp(25)},
				{name: "b", w: 60, h: 20, style: instructions.ItemStyle{}, expectY: intp(85)},
				{name: "c", w: 50, h: 20, style: instructions.ItemStyle{}, expectY: intp(85)},
			},
		},
		{
			name:    "absolute_right_bottom_with_margins",
			originX: 10, originY: 20,
			// Position from padding-box with side margins on the same side:
			// padding: 8 on all sides
			// innerW = 200 - 8 - 8 = 184; innerH = 100 - 8 - 8 = 84
			// cx0 = 10 + 8 = 18; cx1 = 18 + 184 = 202
			// cy0 = 20 + 8 = 28; cy1 = 28 + 84 = 112
			// elem: w=40,h=20, right=15,bottom=10, margin=[2,3,4,5] (t,r,b,l)
			// x = cx1 - right - w - mr = 202 - 15 - 40 - 3 = 144
			// y = cy1 - bottom - h - mb = 112 - 10 - 20 - 4 = 78
			style: instructions.ContainerStyle{
				Display:   instructions.DisplayFlex,
				Direction: instructions.Row,
				Padding:   [4]int{8, 8, 8, 8},
				Width:     200,
				Height:    100,
			},
			items: []itemCase{
				{name: "abs", w: 40, h: 20, style: instructions.ItemStyle{
					Position: instructions.PosAbsolute,
					Right:    intp(15),
					Bottom:   intp(10),
					Margin:   [4]int{2, 3, 4, 5},
				}, expectX: intp(144), expectY: intp(78)},
			},
		},
		{
			name:    "mixed_flex_grow_and_shrink_row",
			originX: 10, originY: 20,
			// innerW = 180 - 5 - 5 = 170
			// baseMainContent sum = 60 + 60 = 120; gaps = 10
			// flexFree = 170 - 120 - 10 = 40 > 0 → only grow applies
			// totalGrow = 2 (only first item) → a gets +40, b stays 60
			// a.w = 100; b.w = 60
			style: instructions.ContainerStyle{
				Display:    instructions.DisplayFlex,
				Direction:  instructions.Row,
				Padding:    [4]int{5, 5, 5, 5},
				Gap:        instructions.Vector2{X: 10, Y: 0},
				Width:      180,
				Height:     60,
				Justify:    instructions.JustifyStart,
				AlignItems: instructions.AlignItemsStart,
			},
			items: []itemCase{
				{name: "a", w: 60, h: 20, style: instructions.ItemStyle{FlexGrow: 2}, expectX: intp(15), expectY: intp(25), expectW: intp(100), expectH: intp(20)},
				{name: "b", w: 60, h: 20, style: instructions.ItemStyle{FlexShrink: 2}, expectX: intp(15 + 100 + 10), expectY: intp(25), expectW: intp(60), expectH: intp(20)},
			},
		},
		{
			name:    "ignore_gap_before_first_and_middle",
			originX: 10, originY: 20,
			// Normal positions without skipping gap: x = 15, 65, 115
			// Last item skips gapBefore → x(c) = 105
			style: instructions.ContainerStyle{
				Display:    instructions.DisplayFlex,
				Direction:  instructions.Row,
				Padding:    [4]int{5, 5, 5, 5},
				Gap:        instructions.Vector2{X: 10, Y: 0},
				Width:      200,
				Height:     60,
				Justify:    instructions.JustifyStart,
				AlignItems: instructions.AlignItemsStart,
			},
			items: []itemCase{
				{name: "a", w: 40, h: 20, style: instructions.ItemStyle{IgnoreGapBefore: true}, expectX: intp(15), expectY: intp(25)},
				{name: "b", w: 40, h: 20, style: instructions.ItemStyle{}, expectX: intp(65), expectY: intp(25)},
				{name: "c", w: 40, h: 20, style: instructions.ItemStyle{IgnoreGapBefore: true}, expectX: intp(105), expectY: intp(25)},
			},
		},
		{
			name:    "alignitems_stretch_auto_height_row",
			originX: 10, originY: 20,
			// Single line, cross(height) max = 30; AlignItems=Stretch
			// → items stretched to lineCross (minus margins=0) → h=30 each
			// Auto container height:
			//  innerH = sum(ln.crossUsed) = 30; outerH = innerH + pt + pb = 30 + 5 + 5 = 40
			style: instructions.ContainerStyle{
				Display:    instructions.DisplayFlex,
				Direction:  instructions.Row,
				Padding:    [4]int{5, 5, 5, 5},
				Gap:        instructions.Vector2{X: 10, Y: 0},
				Width:      200,
				Height:     0, // auto
				AlignItems: instructions.AlignItemsStretch,
				Justify:    instructions.JustifyStart,
			},
			items: []itemCase{
				{name: "a", w: 30, h: 20, style: instructions.ItemStyle{}, expectY: intp(25), expectH: intp(30)},
				{name: "b", w: 30, h: 30, style: instructions.ItemStyle{}, expectY: intp(25), expectH: intp(30)},
			},
			expectOuterH: floatp(40.0),
		},
		{
			name:    "aligncontent_center_wrap_row",
			originX: 10, originY: 20,
			// innerW = 120 - 5 - 5 = 110
			// items: 70 and 60 → wrap into two lines
			// totalCross = 20 + 6 + 20 = 46
			// leftover = innerH(=110) - 46 = 64 → offset = 32 (center)
			// line1 y = 20 + 5 + 32 = 57
			// line2 y = 57 + 20 + 6 = 83
			style: instructions.ContainerStyle{
				Display:      instructions.DisplayFlex,
				Direction:    instructions.Row,
				Wrap:         true,
				Padding:      [4]int{5, 5, 5, 5},
				Gap:          instructions.Vector2{X: 10, Y: 6},
				Justify:      instructions.JustifyStart,
				AlignItems:   instructions.AlignItemsStart,
				AlignContent: instructions.AlignItemsCenter,
				Width:        120,
				Height:       120,
			},
			items: []itemCase{
				{name: "a", w: 70, h: 20, style: instructions.ItemStyle{}, expectX: intp(15), expectY: intp(57)},
				{name: "b", w: 60, h: 20, style: instructions.ItemStyle{}, expectX: intp(15), expectY: intp(83)},
			},
		},
		{
			name:    "column_wrap_auto_width_uses_crossUsed",
			originX: 10, originY: 20,
			// Column direction: main=Y, cross=X.
			// innerH = 120 - 5 - 5 = 110, item heights each 60 → 3 columns.
			// Column widths (cross): 40 | 80 | 70; gapX=12.
			// Auto innerW = 40 + 12 + 80 + 12 + 70 = 214 → outerW = 214 + 5 + 5 = 224.
			// Column x positions: contentLeft=10+5=15; x2 = 15 + 40 + 12 = 67; x3 = 67 + 80 + 12 = 159.
			style: instructions.ContainerStyle{
				Display:    instructions.DisplayFlex,
				Direction:  instructions.Column,
				Wrap:       true,
				Padding:    [4]int{5, 5, 5, 5},
				Gap:        instructions.Vector2{X: 12, Y: 10},
				Justify:    instructions.JustifyStart,
				AlignItems: instructions.AlignItemsStart,
				Width:      0,   // auto width
				Height:     120, // fixed height
			},
			items: []itemCase{
				{name: "a", w: 40, h: 60, style: instructions.ItemStyle{}, expectX: intp(15)},
				{name: "b", w: 80, h: 60, style: instructions.ItemStyle{}, expectX: intp(67)},
				{name: "c", w: 70, h: 60, style: instructions.ItemStyle{}, expectX: intp(159)},
			},
			expectOuterW: floatp(224.0),
		},
		{
			name:    "row_justify_variants_end_and_spacearound",
			originX: 10, originY: 20,
			// innerW = 200 - 5 - 5 = 190
			// used = 50 + 10 + 30 = 90 → remaining = 100
			// End: offset = 100 → x1 = 10+5+100 = 115; x2 = 115+50+10 = 175.
			// SpaceAround: extra = remaining / n = 100/2 = 50; offset=25
			//   x1 = 10+5+25 = 40; x2 = 40+50+50 = 140; but there is also fixed gap=10 between items:
			//   Justify extra applies in addition to fixed gap: x2 = 40 + 50 /*w*/ + 10 /*gap*/ + 50 /*extra*/ = 150.
			style: instructions.ContainerStyle{
				Display:    instructions.DisplayFlex,
				Direction:  instructions.Row,
				Padding:    [4]int{5, 5, 5, 5},
				Gap:        instructions.Vector2{X: 10, Y: 0},
				Width:      200,
				Height:     60,
				AlignItems: instructions.AlignItemsStart,
			},
			items: []itemCase{
				// We’ll run twice: once with End, once with SpaceAround. See below.
				{name: "m1", w: 50, h: 20, style: instructions.ItemStyle{}},
				{name: "m2", w: 30, h: 20, style: instructions.ItemStyle{}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// For the "row_justify_variants_end_and_spacearound" case,
			// run it twice with different Justify expectations.
			if tc.name == "row_justify_variants_end_and_spacearound" {
				// Build once per variant to isolate state.
				run := func(j instructions.JustifyContent, exp1, exp2 int) {
					// fresh mocks
					m1 := newMock("m1", tc.items[0].w, tc.items[0].h)
					m2 := newMock("m2", tc.items[1].w, tc.items[1].h)

					style := tc.style
					style.Justify = j

					al := instructions.NewAutoLayout(tc.originX, tc.originY, style)
					al.Add(m1, tc.items[0].style)
					al.Add(m2, tc.items[1].style)

					base, overlay := newCanvases()
					al.Draw(base, overlay)

					require.Equal(t, exp1, m1.x)
					require.Equal(t, exp2, m2.x)
				}
				run(instructions.JustifyEnd, 115, 175)
				run(instructions.JustifySpaceAround, 40, 150)
				return
			}

			// Normal single pass for other cases.
			al := instructions.NewAutoLayout(tc.originX, tc.originY, tc.style)
			mocks := make([]*mockShape, 0, len(tc.items))
			for _, it := range tc.items {
				m := newMock(it.name, it.w, it.h)
				mocks = append(mocks, m)
				al.Add(m, it.style)
			}

			base, overlay := newCanvases()
			al.Draw(base, overlay)

			// Assert per-item expectations if provided.
			for i, it := range tc.items {
				m := mocks[i]
				if it.expectX != nil {
					require.Equalf(t, *it.expectX, m.x, "[%s] x mismatch", it.name)
				}
				if it.expectY != nil {
					require.Equalf(t, *it.expectY, m.y, "[%s] y mismatch", it.name)
				}
				if it.expectW != nil {
					require.Equalf(t, *it.expectW, m.w, "[%s] w mismatch", it.name)
				}
				if it.expectH != nil {
					require.Equalf(t, *it.expectH, m.h, "[%s] h mismatch", it.name)
				}
				// Ensure each item was drawn at least once unless explicitly disabled.
				require.Greaterf(t, m.drawCalls, 0, "[%s] expected at least one draw call", it.name)
			}

			// Optional container outer size assertions.
			sz := al.Size()
			if tc.expectOuterW != nil {
				require.Equal(t, *tc.expectOuterW, sz.Width(), "outer width mismatch")
			}
			if tc.expectOuterH != nil {
				require.Equal(t, *tc.expectOuterH, sz.Height(), "outer height mismatch")
			}
		})
	}
}

/*
TestAutoLayout_RenderSmoke
Lightweight image-output check:
- Build a real layout with two rectangles.
- Draw to base/overlay.
- Probe a couple of pixels which must be non-zero alpha.
Formulas:

	contentLeft = originX + paddingLeft
	contentTop  = originY + paddingTop
	item0: (contentLeft, contentTop)
	item1: y = contentTop + h0 + gapY
*/
func TestAutoLayout_RenderSmoke(t *testing.T) {
	style := instructions.ContainerStyle{
		Display:    instructions.DisplayFlex,
		Direction:  instructions.Column,
		Padding:    [4]int{16, 16, 16, 16},
		Gap:        instructions.Vector2{X: 0, Y: 16},
		Justify:    instructions.JustifyStart,
		AlignItems: instructions.AlignItemsStart,
		Width:      500,
		Height:     1000,
	}

	al := instructions.NewAutoLayout(10, 20, style)

	// Two solid rectangles to guarantee visible pixels.
	r1 := instructions.NewRectangle(0, 0, 100, 100).
		SetRadius(12).
		SetFillColor(colors.Coral).
		SetStrokeColor(colors.Navy)
	r2 := instructions.NewRectangle(0, 0, 100, 100).
		SetRadius(12).
		SetFillColor(colors.Coral).
		SetStrokeColor(colors.Navy)

	al.Add(r1, instructions.ItemStyle{})
	al.Add(r2, instructions.ItemStyle{})

	base, overlay := newCanvases()
	al.Draw(base, overlay)

	// Compute sample points inside each rectangle.
	contentLeft := 10 + 16
	contentTop := 20 + 16
	p1x, p1y := contentLeft+50, contentTop+50        // inside first rect
	p2x, p2y := contentLeft+50, contentTop+100+16+50 // inside second rect (y + h0 + gap + 50)

	px1B := base.RGBAAt(p1x, p1y)
	px1O := overlay.RGBAAt(p1x, p1y)
	px2B := base.RGBAAt(p2x, p2y)
	px2O := overlay.RGBAAt(p2x, p2y)

	// At least one of the layers must have non-zero alpha where the rectangles are drawn.
	require.True(t, px1B.A > 0 || px1O.A > 0, "first rect pixel must be visible")
	require.True(t, px2B.A > 0 || px2O.A > 0, "second rect pixel must be visible")

	// Sanity: container outer size is preserved.
	sz := al.Size()
	require.Equal(t, 500.0, sz.Width())
	require.Equal(t, 1000.0, sz.Height())
}
