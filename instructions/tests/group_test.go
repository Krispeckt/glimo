package glimo_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/instructions"
	"github.com/Krispeckt/glimo/internal/render"
	"github.com/stretchr/testify/require"
)

func TestGroup_Size(t *testing.T) {
	g := instructions.NewGroup()
	require.NotNil(t, g)

	// empty -> 0x0
	require.Equal(t, float64(0), g.Size().Width())
	require.Equal(t, float64(0), g.Size().Height())

	// add shapes with different sizes
	g.AddInstructions(
		instructions.NewRectangle(0, 0, 100, 50).SetFillColor(colors.Coral),
		instructions.NewRectangle(0, 0, 120, 80).SetFillColor(colors.Orange),
		instructions.NewRectangle(0, 0, 90, 120).SetFillColor(colors.SkyBlue),
	)
	// Size is the union by component-wise max(width,height)
	require.Equal(t, float64(120), g.Size().Width())
	require.Equal(t, float64(120), g.Size().Height())
}

func TestGroup_DrawAndExport(t *testing.T) {
	font := render.MustLoadFont("testdata/montserrat.ttf", 48)

	type testCase struct {
		name  string
		setup func(*instructions.Group)
	}

	cases := []testCase{
		{
			name: "basic",
			setup: func(g *instructions.Group) {
				g.AddInstructions(
					instructions.NewRectangle(20, 20, 120, 80).
						SetFillColor(colors.Pumpkin).
						SetStrokeColor(colors.RebeccaPurple).
						SetLineWidth(3),
				)
			},
		},
		{
			name: "overlap_transparency",
			setup: func(g *instructions.Group) {
				g.AddInstructions(
					instructions.NewRectangle(20, 20, 120, 120).
						SetFillColor(colors.RGBA(255, 0, 0, 128)).
						SetStrokeColor(colors.White).
						SetLineWidth(2),
					instructions.NewRectangle(60, 60, 120, 120).
						SetFillColor(colors.RGBA(0, 0, 255, 128)).
						SetStrokeColor(colors.White).
						SetLineWidth(2),
				)
			},
		},
		{
			name: "rounded_and_various",
			setup: func(g *instructions.Group) {
				g.AddInstructions(
					instructions.NewRectangle(20, 20, 100, 100).
						SetRadius(20).
						SetFillColor(colors.Coral).
						SetStrokeColor(colors.Navy),
					instructions.NewRectangle(150, 20, 120, 100).
						SetCornerRadii(5, 20, 40, 10).
						SetFillColor(colors.DavysGray).
						SetStrokeColor(colors.Orange).
						SetLineWidth(2),
				)
			},
		},
		{
			name: "multiple_sizes_chain",
			setup: func(g *instructions.Group) {
				offsetX := 20.0
				offsetY := 20.0
				sizes := []struct {
					w, h, r float64
				}{
					{50, 50, 0},
					{100, 50, 10},
					{150, 100, 20},
					{50, 150, 25},
				}
				var shapes []instructions.BoundedShape
				for i, s := range sizes {
					shapes = append(shapes,
						instructions.NewRectangle(offsetX, offsetY+float64(i)*90, s.w, s.h).
							SetFillColor(colors.OrangeRed).
							SetStrokeColor(colors.MidnightBlue).
							SetRadius(s.r),
					)
					offsetX += 10
				}
				g.AddInstructions(shapes...)
			},
		},
		{
			name: "no_fill_no_stroke",
			setup: func(g *instructions.Group) {
				g.AddInstructions(
					instructions.NewRectangle(40, 40, 80, 80).
						SetFillPattern(nil).
						SetStrokeColor(colors.MediumSpringGreen).
						SetLineWidth(4),
				)
			},
		},
		{
			name: "negative_coords_clipping",
			setup: func(g *instructions.Group) {
				// Intentional negative coords; Group uses size union, so part may clip.
				g.AddInstructions(
					instructions.NewRectangle(-50, -50, 300, 300).
						SetFillColor(colors.CornflowerBlue).
						SetStrokeColor(colors.DavysGray).
						SetLineWidth(5),
				)
			},
		},
		{
			name: "text_blend_mode",
			setup: func(g *instructions.Group) {
				// Intentional negative coords; Group uses size union, so part may clip.
				g.AddInstructions(
					instructions.NewRectangle(0, 0, 500, 500).
						SetFillColor(colors.WhiteSmoke),
					instructions.NewRectangle(10, 10, 400, 400).
						SetFillColor(colors.Pumpkin),
					instructions.NewText("Glimo Test", 20, 20, font).
						SetColorPattern(colors.NewSolidWithBlend(colors.Black, colors.BlendOverlay, 50)),
				)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare group and layer
			g := instructions.NewGroup()
			g.SetPosition(50, 50)
			require.NotNil(t, g)
			tc.setup(g)

			fmt.Println(g.Size())

			layer := instructions.NewLayer(600, 400)
			require.NotNil(t, layer)

			// Draw group at a position and export
			require.NotPanics(t, func() {
				// base may be same as overlay for simple tests
				g.Draw(layer.Image(), layer.Image())
			})

			outPath := filepath.Join("./output", "group_"+tc.name+".png")
			require.NoError(t, layer.Export(outPath), "export failed for %s", tc.name)
		})
	}
}
