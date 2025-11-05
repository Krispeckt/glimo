package glimo_test

import (
	"image"
	"testing"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/instructions"
	"github.com/Krispeckt/glimo/internal/core/geom"
	"github.com/stretchr/testify/require"
)

// mockShape implements BoundedShape for layout verification.
type mockShape struct {
	name      string
	x, y      int
	w, h      int
	drawCalls int
}

func newMock(name string, w, h int) *mockShape { return &mockShape{name: name, w: w, h: h} }

func (m *mockShape) Draw(_, _ *image.RGBA) { m.drawCalls++ }

func (m *mockShape) Position() (int, int) { return m.x, m.y }
func (m *mockShape) SetPosition(x, y int) { m.x, m.y = x, y }
func (m *mockShape) Size() *geom.Size     { return geom.NewSize(float64(m.w), float64(m.h)) }
func (m *mockShape) SetBounds(x, y, w, h int) {
	m.x, m.y, m.w, m.h = x, y, w, h
}
func (m *mockShape) SetSize(w, h int) { m.w, m.h = w, h }

// helpers
func newCanvases() (*image.RGBA, *image.RGBA) {
	base := image.NewRGBA(image.Rect(0, 0, 800, 600))
	overlay := image.NewRGBA(image.Rect(0, 0, 800, 600))
	return base, overlay
}

func TestAutoLayoutRowJustifyAndGap(t *testing.T) {
	type tc struct {
		name      string
		style     instructions.ContainerStyle
		expectPos [][2]int // [(x,y) per item]
	}
	m1 := newMock("a", 50, 20)
	m2 := newMock("b", 30, 20)

	cases := []tc{
		{
			name: "row_start_gap",
			style: instructions.ContainerStyle{
				Display:    instructions.DisplayFlex,
				Direction:  instructions.Row,
				Wrap:       false,
				Padding:    [4]int{5, 5, 5, 5},
				Gap:        instructions.Vector2{X: 10, Y: 0},
				Justify:    instructions.JustifyStart,
				AlignItems: instructions.AlignItemsStart,
				Width:      200,
				Height:     60,
			},
			// x1=10+5=15, y1=20+5=25; x2=15+50+10=75, y2=25
			expectPos: [][2]int{{15, 25}, {75, 25}},
		},
		{
			name: "row_center_gap",
			style: instructions.ContainerStyle{
				Display:    instructions.DisplayFlex,
				Direction:  instructions.Row,
				Wrap:       false,
				Padding:    [4]int{5, 5, 5, 5},
				Gap:        instructions.Vector2{X: 10, Y: 0},
				Justify:    instructions.JustifyCenter,
				AlignItems: instructions.AlignItemsStart,
				Width:      200, // innerW=190; base=50+10+30=90; free=100; offset=50
				Height:     60,
			},
			// x1=10+5+50=65; x2=65+50+10=125; y=25
			expectPos: [][2]int{{65, 25}, {125, 25}},
		},
	}

	for _, cse := range cases {
		t.Run(cse.name, func(t *testing.T) {
			// reset mocks
			m1.x, m1.y, m1.drawCalls = 0, 0, 0
			m2.x, m2.y, m2.drawCalls = 0, 0, 0

			al := instructions.NewAutoLayout(10, 20, cse.style)
			al.Add(m1, instructions.ItemStyle{})
			al.Add(m2, instructions.ItemStyle{})

			base, overlay := newCanvases()
			al.Draw(base, overlay)

			require.Equal(t, cse.expectPos[0][0], m1.x)
			require.Equal(t, cse.expectPos[0][1], m1.y)
			require.Equal(t, cse.expectPos[1][0], m2.x)
			require.Equal(t, cse.expectPos[1][1], m2.y)
			require.Greater(t, m1.drawCalls, 0)
			require.Greater(t, m2.drawCalls, 0)
		})
	}
}

func TestAutoLayoutRowAlignItems(t *testing.T) {
	// m1.h=20, m2.h=40; line cross=40; center => y1=20+5+10=35, y2=20+5+0=25
	m1 := newMock("a", 50, 20)
	m2 := newMock("b", 30, 40)

	style := instructions.ContainerStyle{
		Display:    instructions.DisplayFlex,
		Direction:  instructions.Row,
		Padding:    [4]int{5, 5, 5, 5},
		Gap:        instructions.Vector2{X: 10, Y: 0},
		Justify:    instructions.JustifyStart,
		AlignItems: instructions.AlignItemsCenter,
		Width:      200,
		Height:     80,
	}

	al := instructions.NewAutoLayout(10, 20, style)
	al.Add(m1, instructions.ItemStyle{})
	al.Add(m2, instructions.ItemStyle{})

	base, overlay := newCanvases()
	al.Draw(base, overlay)

	require.Equal(t, 15, m1.x)
	require.Equal(t, 35, m1.y) // centered within 40px cross
	require.Equal(t, 75, m2.x)
	require.Equal(t, 25, m2.y) // taller item stays at top in center formula
}

func TestAutoLayoutWrapRow(t *testing.T) {
	// innerW=110; items: 70 and 60; 70+10+60>110 => wrap.
	m1 := newMock("a", 70, 20)
	m2 := newMock("b", 60, 18)

	style := instructions.ContainerStyle{
		Display:    instructions.DisplayFlex,
		Direction:  instructions.Row,
		Wrap:       true,
		Padding:    [4]int{5, 5, 5, 5},
		Gap:        instructions.Vector2{X: 10, Y: 6},
		Justify:    instructions.JustifyStart,
		AlignItems: instructions.AlignItemsStart,
		Width:      120, // innerW=110
		Height:     200,
	}

	al := instructions.NewAutoLayout(10, 20, style)
	al.Add(m1, instructions.ItemStyle{})
	al.Add(m2, instructions.ItemStyle{})

	base, overlay := newCanvases()
	al.Draw(base, overlay)

	// First line y=25; second line y = 25 + cross(=20) + gy(=6) = 51
	require.Equal(t, 15, m1.x)
	require.Equal(t, 25, m1.y)
	require.Equal(t, 15, m2.x)
	require.Equal(t, 51, m2.y)
}

func TestAutoLayoutColumnDirection(t *testing.T) {
	// column: main=Y, cross=X
	m1 := newMock("a", 30, 30)
	m2 := newMock("b", 50, 50)

	style := instructions.ContainerStyle{
		Display:    instructions.DisplayFlex,
		Direction:  instructions.Column,
		Wrap:       false,
		Padding:    [4]int{5, 5, 5, 5},
		Gap:        instructions.Vector2{X: 0, Y: 10},
		Justify:    instructions.JustifyStart,
		AlignItems: instructions.AlignItemsStart,
		Width:      200,
		Height:     200,
	}

	al := instructions.NewAutoLayout(10, 20, style)
	al.Add(m1, instructions.ItemStyle{})
	al.Add(m2, instructions.ItemStyle{})

	base, overlay := newCanvases()
	al.Draw(base, overlay)

	// x fixed at 15, y steps by heights + gy
	require.Equal(t, 15, m1.x)
	require.Equal(t, 25, m1.y)
	require.Equal(t, 15, m2.x)
	require.Equal(t, 65, m2.y) // 25 + 30 + 10
}

func TestAutoLayoutAbsolutePositioning(t *testing.T) {
	abs := newMock("abs", 30, 20)

	top := 10
	right := 15

	style := instructions.ContainerStyle{
		Display:   instructions.DisplayFlex,
		Direction: instructions.Row,
		Padding:   [4]int{5, 5, 5, 5},
		Gap:       instructions.Vector2{X: 8, Y: 8},
		Width:     200, // innerW=190
		Height:    100, // innerH=90
	}

	al := instructions.NewAutoLayout(10, 20, style)
	al.Add(abs, instructions.ItemStyle{
		Position: instructions.PosAbsolute,
		Top:      &top,
		Right:    &right,
		ZIndex:   10,
	})

	base, overlay := newCanvases()
	al.Draw(base, overlay)

	// padding-box: cx0=10+5=15; cy0=20+5=25; cx1=10+5+190=205
	// x = 205 - 15 - 30 = 160; y = 25 + 10 = 35
	require.Equal(t, 160, abs.x)
	require.Equal(t, 35, abs.y)
}

func TestAutoLayoutFlexGrowShrinkRow(t *testing.T) {
	// innerW=280; base = 50 + 10 + 50 = 110; free=170
	// grow: a=1, b=3 => sizes: a=50+42=92, b=50+127=177
	a := newMock("a", 50, 20)
	b := newMock("b", 50, 20)

	style := instructions.ContainerStyle{
		Display:    instructions.DisplayFlex,
		Direction:  instructions.Row,
		Padding:    [4]int{5, 5, 5, 5},
		Gap:        instructions.Vector2{X: 10, Y: 0},
		Justify:    instructions.JustifyStart,
		AlignItems: instructions.AlignItemsStart,
		Width:      300, // innerW=280
		Height:     80,
	}

	al := instructions.NewAutoLayout(10, 20, style)
	al.Add(a, instructions.ItemStyle{FlexGrow: 1})
	al.Add(b, instructions.ItemStyle{FlexGrow: 3})

	base, overlay := newCanvases()
	al.Draw(base, overlay)

	// innerW=290; base = 50 + 10 + 50 = 110; free=180
	// grow: a=1, b=3 => sizes: a=95, b=185
	require.Equal(t, 15, a.x)
	require.Equal(t, 25, a.y)
	require.Equal(t, 120, b.x)
	require.Equal(t, 25, b.y)
}

func TestAutoLayoutAlignSelfOverrides(t *testing.T) {
	// container AlignItemsStart, item overrides to Center
	m := newMock("m", 40, 20)

	style := instructions.ContainerStyle{
		Display:    instructions.DisplayFlex,
		Direction:  instructions.Row,
		Padding:    [4]int{5, 5, 5, 5},
		Gap:        instructions.Vector2{X: 0, Y: 0},
		Justify:    instructions.JustifyStart,
		AlignItems: instructions.AlignItemsStart,
		Width:      200,
		Height:     60,
	}
	align := instructions.AlignItemsCenter

	al := instructions.NewAutoLayout(10, 20, style)
	al.Add(m, instructions.ItemStyle{AlignSelf: &align})

	base, overlay := newCanvases()
	al.Draw(base, overlay)

	// line cross equals item height => center => y = 25 (no offset)
	require.Equal(t, 15, m.x)
	require.Equal(t, 25, m.y)
}

func TestAutoLayoutOutput(t *testing.T) {
	layer := newLayer(t, 500, 1000)

	layer.LoadInstruction(
		instructions.NewAutoLayout(
			10, 20,
			instructions.ContainerStyle{
				Display:    instructions.DisplayFlex,
				Direction:  instructions.Column,
				Padding:    [4]int{16, 16, 16, 16},
				Gap:        instructions.Vector2{X: 0, Y: 16},
				Justify:    instructions.JustifyStart,
				AlignItems: instructions.AlignItemsStart,
				Width:      500,
				Height:     1000,
			},
		).Add(
			instructions.NewRectangle(0, 0, 100, 100).
				SetRadius(20).
				SetFillColor(colors.Coral).
				SetStrokeColor(colors.Navy),
			instructions.ItemStyle{},
		).Add(
			instructions.NewRectangle(0, 0, 100, 100).
				SetRadius(20).
				SetFillColor(colors.Coral).
				SetStrokeColor(colors.Navy),
			instructions.ItemStyle{},
		),
	)

	err := layer.Export("./output/auto_layout.png")
	require.NoError(t, err, "export failed for auto_layout")
}

func TestAutoLayoutIgnoreGapBefore(t *testing.T) {
	// Two items, second skips gap before itself.
	// innerW = 200 - 5 - 5 = 190
	// a.width=50, b.width=30
	// normally: x(a)=15, x(b)=15+50+10=75
	// with b.IgnoreGapBefore = true => x(b)=15+50=65 (gap skipped)
	a := newMock("a", 50, 20)
	b := newMock("b", 30, 20)

	style := instructions.ContainerStyle{
		Display:    instructions.DisplayFlex,
		Direction:  instructions.Row,
		Padding:    [4]int{5, 5, 5, 5},
		Gap:        instructions.Vector2{X: 10, Y: 0},
		Justify:    instructions.JustifyStart,
		AlignItems: instructions.AlignItemsStart,
		Width:      200,
		Height:     60,
	}

	al := instructions.NewAutoLayout(10, 20, style)
	al.Add(a, instructions.ItemStyle{})
	al.Add(b, instructions.ItemStyle{IgnoreGapBefore: true})

	base, overlay := newCanvases()
	al.Draw(base, overlay)

	// Позиции.
	require.Equal(t, 15, a.x)
	require.Equal(t, 25, a.y)
	require.Equal(t, 65, b.x) // gap skipped
	require.Equal(t, 25, b.y)
	require.Greater(t, a.drawCalls, 0)
	require.Greater(t, b.drawCalls, 0)

	// Проверка размеров контейнера.
	sz := al.Size()
	require.Equal(t, 200.0, sz.Width())
	require.Equal(t, 60.0, sz.Height())

	// Checking the “length” of the occupied line.
	// Left border of content-box = origin.x + paddingLeft = 10 + 5 = 15.
	// Right border of last element = b.x + b.w = 65 + 30 = 95.
	// Occupied length = 95 - 15 = 80 (50 + 0 + 30).
	contentLeft := 10 + 5
	used := (b.x + b.w) - contentLeft
	require.Equal(t, 80, used)

	// gap: 190 - 1*10 = 180.
	innerW := style.Width - style.Padding[3] - style.Padding[1]
	effectiveLimit := innerW - int(style.Gap.X) // один пропущенный gap
	require.LessOrEqual(t, used, effectiveLimit)

	a2 := newMock("a2", 50, 20)
	b2 := newMock("b2", 30, 20)
	al2 := instructions.NewAutoLayout(10, 20, style)
	al2.Add(a2, instructions.ItemStyle{})
	al2.Add(b2, instructions.ItemStyle{})

	base2, overlay2 := newCanvases()
	al2.Draw(base2, overlay2)

	require.Equal(t, 75, b2.x)                    // 15 + 50 + 10
	require.Equal(t, 90, (b2.x+b2.w)-contentLeft) // 50 + 10 + 30
}

func TestAutoLayoutIgnoreGapBeforeColumn(t *testing.T) {
	// Column direction, Height==0: auto height should match total main-axis size.
	// If the second item skips its vertical gap, the total column height decreases by the gap size (10px).

	m1 := newMock("a", 30, 30)
	m2 := newMock("b", 40, 40)

	style := instructions.ContainerStyle{
		Display:    instructions.DisplayFlex,
		Direction:  instructions.Column,
		Wrap:       false,
		Padding:    [4]int{5, 5, 5, 5},
		Gap:        instructions.Vector2{X: 0, Y: 10},
		Justify:    instructions.JustifyStart,
		AlignItems: instructions.AlignItemsStart,
		Width:      200,
		Height:     0, // auto height
	}

	// Case 1: normal behavior (no gap ignored)
	// Content height = 30 + 10 + 40 = 80
	// Total height including padding = 80 + 5(top) + 5(bottom) = 90
	al1 := instructions.NewAutoLayout(10, 20, style)
	al1.Add(m1, instructions.ItemStyle{})
	al1.Add(m2, instructions.ItemStyle{})
	base1, overlay1 := newCanvases()
	al1.Draw(base1, overlay1)
	sz1 := al1.Size()
	require.Equal(t, 90.0, sz1.Height())

	// Case 2: second item ignores the gap before itself
	// Content height = 30 + 0 + 40 = 70
	// Total height including padding = 70 + 5(top) + 5(bottom) = 80
	m1b := newMock("a", 30, 30)
	m2b := newMock("b", 40, 40)
	al2 := instructions.NewAutoLayout(10, 20, style)
	al2.Add(m1b, instructions.ItemStyle{})
	al2.Add(m2b, instructions.ItemStyle{IgnoreGapBefore: true})
	base2, overlay2 := newCanvases()
	al2.Draw(base2, overlay2)
	sz2 := al2.Size()
	require.Equal(t, 80.0, sz2.Height())

	// Position check: the second item must be placed directly after the first one (no 10px gap)
	require.Equal(t, 25, m1b.y) // 10 + 5 + 10 = 25 (container origin + padding)
	require.Equal(t, 55, m2b.y) // 25 + 30 = 55 (no gap applied)
}

// JustifyContent variants: End / SpaceBetween / SpaceAround / SpaceEvenly
func TestAutoLayoutRowJustifyVariants(t *testing.T) {
	build := func(j instructions.JustifyContent) (int, int) {
		m1 := newMock("a", 50, 20)
		m2 := newMock("b", 30, 20)
		style := instructions.ContainerStyle{
			Display:    instructions.DisplayFlex,
			Direction:  instructions.Row,
			Padding:    [4]int{5, 5, 5, 5},
			Gap:        instructions.Vector2{X: 10, Y: 0},
			Justify:    j,
			AlignItems: instructions.AlignItemsStart,
			Width:      200, // innerW = 190
			Height:     60,
		}
		al := instructions.NewAutoLayout(10, 20, style)
		al.Add(m1, instructions.ItemStyle{})
		al.Add(m2, instructions.ItemStyle{})
		base, overlay := newCanvases()
		al.Draw(base, overlay)
		return m1.x, m2.x
	}

	// End: remaining space = 100; offset = 115
	x1, x2 := build(instructions.JustifyEnd)
	require.Equal(t, 115, x1)
	require.Equal(t, 175, x2)

	// SpaceBetween: extra = 100 / (n-1) = 100
	x1, x2 = build(instructions.JustifySpaceBetween)
	require.Equal(t, 15, x1)
	require.Equal(t, 175, x2)

	// SpaceAround: extra = remaining / n = 50; offset = 25
	x1, x2 = build(instructions.JustifySpaceAround)
	require.Equal(t, 40, x1)
	require.Equal(t, 150, x2)

	// SpaceEvenly: extra = remaining / (n+1) = 33; offset = 33
	x1, x2 = build(instructions.JustifySpaceEvenly)
	require.Equal(t, 48, x1)
	require.Equal(t, 141, x2)
}

// Stretch across the cross axis + AlignContent:Stretch stretches line to inner height
func TestAutoLayoutRowAlignItemsStretchAndAlignContentStretch(t *testing.T) {
	a := newMock("a", 50, 20)
	b := newMock("b", 60, 30)
	style := instructions.ContainerStyle{
		Display:      instructions.DisplayFlex,
		Direction:    instructions.Row,
		Padding:      [4]int{5, 5, 5, 5},
		Gap:          instructions.Vector2{X: 10, Y: 6},
		Justify:      instructions.JustifyStart,
		AlignItems:   instructions.AlignItemsStretch, // items stretch along cross-axis
		AlignContent: instructions.AlignItemsStretch, // line stretches to fill inner height
		Width:        300,
		Height:       100, // innerH = 90
	}
	al := instructions.NewAutoLayout(10, 20, style)
	al.Add(a, instructions.ItemStyle{})
	al.Add(b, instructions.ItemStyle{})
	base, overlay := newCanvases()
	al.Draw(base, overlay)

	require.Equal(t, 90, a.h)
	require.Equal(t, 90, b.h)
	require.Equal(t, 25, a.y)
	require.Equal(t, 25, b.y)
}

// --- FlexShrink distribution when space is insufficient (row) ---
func TestAutoLayoutFlexShrinkDistributionRow(t *testing.T) {
	a := newMock("a", 60, 20)
	b := newMock("b", 60, 20)

	// innerW = 110 - 10 = 100
	// base total = 130 → shortage = 30
	// shrink ratios: a=1, b=3 → a loses 8px, b loses 22px
	style := instructions.ContainerStyle{
		Display:    instructions.DisplayFlex,
		Direction:  instructions.Row,
		Padding:    [4]int{5, 5, 5, 5},
		Gap:        instructions.Vector2{X: 10, Y: 0},
		Justify:    instructions.JustifyStart,
		AlignItems: instructions.AlignItemsStart,
		Width:      110,
		Height:     60,
	}
	al := instructions.NewAutoLayout(10, 20, style)
	al.Add(a, instructions.ItemStyle{FlexShrink: 1})
	al.Add(b, instructions.ItemStyle{FlexShrink: 3})
	base, overlay := newCanvases()
	al.Draw(base, overlay)

	require.Equal(t, 52, a.w)
	require.Equal(t, 38, b.w)
	require.Equal(t, 15, a.x)
	require.Equal(t, 77, b.x)
}

// AlignContent:Center with line wrapping (row, wrap)
func TestAutoLayoutAlignContentCenterWrapRow(t *testing.T) {
	m1 := newMock("a", 70, 20)
	m2 := newMock("b", 60, 20)

	style := instructions.ContainerStyle{
		Display:      instructions.DisplayFlex,
		Direction:    instructions.Row,
		Wrap:         true,
		Padding:      [4]int{5, 5, 5, 5},
		Gap:          instructions.Vector2{X: 10, Y: 6},
		Justify:      instructions.JustifyStart,
		AlignItems:   instructions.AlignItemsStart,
		AlignContent: instructions.AlignItemsCenter,
		Width:        120, // innerW = 110 → wraps
		Height:       120, // innerH = 110
	}

	al := instructions.NewAutoLayout(10, 20, style)
	al.Add(m1, instructions.ItemStyle{})
	al.Add(m2, instructions.ItemStyle{})

	base, overlay := newCanvases()
	al.Draw(base, overlay)

	// totalCross = 46; leftover = 64; vertical offset = 32
	require.Equal(t, 15, m1.x)
	require.Equal(t, 57, m1.y)
	require.Equal(t, 15, m2.x)
	require.Equal(t, 83, m2.y)
}

// Column + wrap + auto width computed from used cross space (crossUsed)
func TestAutoLayoutColumnWrapAutoWidthWithCrossUsed(t *testing.T) {
	// Heights (main axis): 60,60,60 → 3 columns
	// Widths (cross axis): 40 | 80 | 70
	// innerW(auto) = 214 → total outer width = 224
	a := newMock("a", 40, 60)
	b := newMock("b", 80, 60)
	c := newMock("c", 70, 60)

	style := instructions.ContainerStyle{
		Display:    instructions.DisplayFlex,
		Direction:  instructions.Column,
		Wrap:       true,
		Padding:    [4]int{5, 5, 5, 5},
		Gap:        instructions.Vector2{X: 12, Y: 10},
		Justify:    instructions.JustifyStart,
		AlignItems: instructions.AlignItemsStart,
		Width:      0,   // auto width
		Height:     120, // innerH = 110
	}

	al := instructions.NewAutoLayout(10, 20, style)
	al.Add(a, instructions.ItemStyle{})
	al.Add(b, instructions.ItemStyle{})
	al.Add(c, instructions.ItemStyle{})

	base, overlay := newCanvases()
	al.Draw(base, overlay)

	sz := al.Size()
	require.Equal(t, 224.0, sz.Width())

	// Expected X positions of columns: 15 | 67 | 159
	require.Equal(t, 15, a.x)
	require.Equal(t, 67, b.x)
	require.Equal(t, 159, c.x)
}

// Absolute positioning: Bottom + Left with margins
func TestAutoLayoutAbsolutePositioningBottomLeftWithMargins(t *testing.T) {
	abs := newMock("abs", 20, 10)
	left := 10
	bottom := 7
	style := instructions.ContainerStyle{
		Display:   instructions.DisplayFlex,
		Direction: instructions.Row,
		Padding:   [4]int{4, 6, 8, 10}, // top,right,bottom,left
		Gap:       instructions.Vector2{X: 8, Y: 8},
		Width:     140, // innerW = 124
		Height:    100, // innerH = 88
	}
	al := instructions.NewAutoLayout(10, 20, style)
	al.Add(abs, instructions.ItemStyle{
		Position: instructions.PosAbsolute,
		Left:     &left,
		Bottom:   &bottom,
		Margin:   [4]int{0, 0, 2, 3}, // mt, mr, mb, ml
	})

	base, overlay := newCanvases()
	al.Draw(base, overlay)

	// Expected final position:
	// cx0 = 10 + 10 = 20, cy1 = 20 + 4 + 88 = 112
	// x = 20 + 10 + 3 = 33; y = 112 - 7 - 10 - 2 = 93
	require.Equal(t, 33, abs.x)
	require.Equal(t, 93, abs.y)
}
