package glimo_test

import (
	"testing"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/instructions"
	"github.com/stretchr/testify/require"
)

func TestInstructionRectangle(t *testing.T) {
	type testCase struct {
		name  string
		setup func(*testing.T, *instructions.Layer)
	}

	cases := []testCase{
		{
			name: "basic",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
					instructions.NewRectangle(20, 20, 120, 80).
						SetFillPattern(colors.NewSolidWithBlend(colors.Pumpkin, colors.BlendPlusLighter, 1)).
						SetStrokeColor(colors.RebeccaPurple.SetBlendMode(colors.BlendHue)).
						SetLineWidth(3),
				)
			},
		},
		{
			name: "zero_size",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
					instructions.NewRectangle(10, 10, 0, 0),
				)
			},
		},
		{
			name: "rounded_corners",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
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
			name: "no_stroke",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
					instructions.NewRectangle(30, 30, 100, 100).
						SetFillColor(colors.RebeccaPurple).
						SetStrokeColor(colors.Transparent),
				)
			},
		},
		{
			name: "no_fill",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
					instructions.NewRectangle(40, 40, 80, 80).
						SetFillPattern(nil).
						SetStrokeColor(colors.MediumSpringGreen).
						SetLineWidth(4),
				)
			},
		},
		{
			name: "nested",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
					instructions.NewRectangle(50, 50, 300, 300).
						SetFillColor(colors.SkyBlue).
						SetStrokeColor(colors.Navy).
						SetLineWidth(3),
					instructions.NewRectangle(100, 100, 200, 200).
						SetFillColor(colors.LightYellow).
						SetStrokeColor(colors.MediumPurple).
						SetLineWidth(2).
						SetRadius(30),
					instructions.NewRectangle(150, 150, 100, 100).
						SetFillColor(colors.Orange).
						SetStrokeColor(colors.IndianRed).
						SetLineWidth(2).
						SetCornerRadii(10, 30, 50, 0),
				)
			},
		},
		{
			name: "multiple_sizes",
			setup: func(t *testing.T, c *instructions.Layer) {
				var instrs []instructions.Shape
				offsetX := 20.0
				offsetY := 20.0
				sizes := []struct {
					w, h, radius float64
				}{
					{50, 50, 0},
					{100, 50, 10},
					{150, 100, 20},
					{50, 150, 25},
				}
				for i, s := range sizes {
					rect := instructions.NewRectangle(offsetX, offsetY+float64(i)*90, s.w, s.h).
						SetFillColor(colors.OrangeRed).
						SetStrokeColor(colors.MidnightBlue).
						SetRadius(s.radius)
					instrs = append(instrs, rect)
					offsetX += 10
				}
				c.LoadInstructions(instrs...)
			},
		},
		{
			name: "transparency",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
					instructions.NewRectangle(20, 20, 120, 120).
						SetFillColor(colors.RGBA(255, 0, 0, 128)).
						SetStrokeColor(colors.RGBA(255, 255, 255, 255)).
						SetLineWidth(2),
					instructions.NewRectangle(60, 60, 120, 120).
						SetFillColor(colors.RGBA(0, 0, 255, 128)).
						SetStrokeColor(colors.RGBA(255, 255, 255, 255)).
						SetLineWidth(2),
				)
			},
		},
		{
			name: "bounds",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
					instructions.NewRectangle(-50, -50, 300, 300).
						SetFillColor(colors.CornflowerBlue).
						SetStrokeColor(colors.DavysGray).
						SetLineWidth(5),
				)
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

			err := canvas.Export("./output/rect_" + cse.name + ".png")
			require.NoError(t, err, "export failed for %s", cse.name)
		})
	}
}
