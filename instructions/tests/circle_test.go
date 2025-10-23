package glimo_test

import (
	"testing"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/effects"
	"github.com/Krispeckt/glimo/instructions"
	"github.com/stretchr/testify/require"
)

func TestInstructionCircle(t *testing.T) {
	type testCase struct {
		name  string
		setup func(*testing.T, *instructions.Layer)
	}

	cases := []testCase{
		{
			name: "basic",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions([]instructions.Shape{
					instructions.NewCircle(50, 50, 40).
						SetFillColor(colors.Orange).
						SetStrokeColor(colors.Navy).
						SetLineWidth(3),
				})
			},
		},
		{
			name: "zero_radius",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions([]instructions.Shape{
					instructions.NewCircle(100, 100, 0),
				})
			},
		},
		{
			name: "no_stroke",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions([]instructions.Shape{
					instructions.NewCircle(120, 120, 60).
						SetFillColor(colors.MediumSpringGreen).
						SetStrokeColor(colors.Transparent),
				})
			},
		},
		{
			name: "no_fill",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions([]instructions.Shape{
					instructions.NewCircle(150, 150, 50).
						SetFillPattern(nil).
						SetStrokeColor(colors.RebeccaPurple).
						SetLineWidth(4),
				})
			},
		},
		{
			name: "nested",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions([]instructions.Shape{
					instructions.NewCircle(100, 100, 80).
						SetFillColor(colors.SkyBlue).
						SetStrokeColor(colors.Navy).
						SetLineWidth(3),
					instructions.NewCircle(120, 120, 50).
						SetFillColor(colors.LightYellow).
						SetStrokeColor(colors.MediumPurple).
						SetLineWidth(2),
					instructions.NewCircle(140, 140, 25).
						SetFillColor(colors.Orange).
						SetStrokeColor(colors.IndianRed).
						SetLineWidth(2),
				})
			},
		},
		{
			name: "multiple_sizes",
			setup: func(t *testing.T, c *instructions.Layer) {
				var instrs []instructions.Shape
				offsetX := 30.0
				offsetY := 30.0
				radii := []float64{20, 40, 60, 80}
				for i, r := range radii {
					circle := instructions.NewCircle(offsetX+float64(i)*100, offsetY, r).
						SetFillColor(colors.Coral).
						SetStrokeColor(colors.MidnightBlue).
						SetLineWidth(2)
					instrs = append(instrs, circle)
				}
				c.LoadInstructions(instrs)
			},
		},
		{
			name: "transparency_overlap",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions([]instructions.Shape{
					instructions.NewCircle(60, 60, 60).
						SetFillColor(colors.RGBA(255, 0, 0, 128)).
						SetStrokeColor(colors.White).
						SetLineWidth(2),
					instructions.NewCircle(100, 100, 60).
						SetFillColor(colors.RGBA(0, 0, 255, 128)).
						SetStrokeColor(colors.White).
						SetLineWidth(2),
				})
			},
		},
		{
			name: "outside_stroke",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions([]instructions.Shape{
					instructions.NewCircle(200, 200, 60).
						SetFillColor(colors.Crimson).
						SetStrokeColor(colors.Black).
						SetStrokePosition(instructions.StrokeOutside).
						SetLineWidth(8),
				})
			},
		},
		{
			name: "inside_stroke",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions([]instructions.Shape{
					instructions.NewCircle(300, 100, 60).
						SetFillColor(colors.Gold).
						SetStrokeColor(colors.ForestGreen).
						SetStrokePosition(instructions.StrokeInside).
						SetLineWidth(6),
				})
			},
		},
		{
			name: "bounds",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions([]instructions.Shape{
					instructions.NewCircle(-30, -30, 100).
						SetFillColor(colors.CornflowerBlue).
						SetStrokeColor(colors.DavysGray).
						SetLineWidth(5),
				})
			},
		},
		{
			name: "drop_shadow",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions([]instructions.Shape{
					instructions.NewCircle(20, 20, 100).
						SetFillColor(colors.Amethyst).
						AddEffects(effects.NewDropShadow(0, 4, 4, 0, colors.Amethyst, 0.5)),
				})
			},
		},
	}

	for _, cse := range cases {
		t.Run(cse.name, func(t *testing.T) {
			canvas := newLayer(t, 600, 400)
			require.NotNil(t, canvas, "canvas should not be nil")

			require.NotPanics(t, func() {
				cse.setup(t, canvas)
			}, "setup should not panic")

			err := canvas.Export("./output/circle_" + cse.name + ".png")
			require.NoError(t, err, "export failed for %s", cse.name)
		})
	}
}
