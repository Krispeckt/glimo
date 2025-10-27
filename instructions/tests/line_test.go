package glimo_test

import (
	"testing"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/instructions"
	"github.com/stretchr/testify/require"
)

func TestInstructionLine(t *testing.T) {
	type testCase struct {
		name  string
		setup func(*testing.T, *instructions.Layer)
	}

	cases := []testCase{
		{
			name: "straight_stroke",
			setup: func(t *testing.T, ctx *instructions.Layer) {
				bg, err := colors.HEX("#D9D9D9")
				require.NoError(t, err)

				ctx.LoadInstruction(
					instructions.NewRectangle(40, 15, 10, 10).SetFillColor(bg),
				)
				ctx.LoadInstruction(
					instructions.NewLine().
						SetLineWidth(4).
						SetStrokePattern(
							colors.NewLinearGradientWithBlend(20, 20, 200, 200, colors.BlendExclusion, 0.5).
								AddColorStop(0, colors.Amethyst).
								AddColorStop(1, colors.Pumpkin),
						).
						MoveTo(20, 20).
						LineTo(200, 20).
						Stroke(),
				)
			},
		},
		{
			name: "rectangle_fill",
			setup: func(t *testing.T, ctx *instructions.Layer) {
				ctx.LoadInstruction(
					instructions.NewLine().
						SetFillPattern(colors.NewSolid(colors.DarkGreen)).
						MoveTo(50, 50).
						LineTo(200, 50).
						LineTo(200, 200).
						LineTo(50, 200).
						ClosePath().
						Fill(),
				)
			},
		},
		{
			name: "quadratic_curve",
			setup: func(t *testing.T, ctx *instructions.Layer) {
				ctx.LoadInstruction(
					instructions.NewLine().
						SetStrokePattern(colors.NewSolid(colors.CobaltBlue)).
						MoveTo(30, 200).
						QuadraticTo(128, 30, 230, 200).
						SetLineWidth(3).
						Stroke(),
				)
			},
		},
		{
			name: "cubic_curve",
			setup: func(t *testing.T, ctx *instructions.Layer) {
				ctx.LoadInstruction(
					instructions.NewLine().
						SetStrokePattern(colors.NewSolid(colors.OrangeRed)).
						MoveTo(20, 200).
						CubicTo(80, 20, 180, 20, 240, 200).
						SetLineWidth(4).
						Stroke(),
				)
			},
		},
		{
			name: "dashed_line",
			setup: func(t *testing.T, ctx *instructions.Layer) {
				ctx.LoadInstruction(
					instructions.NewLine().
						SetLineWidth(4).
						SetDashes([]float64{10, 5}).
						SetStrokePattern(colors.NewSolid(colors.Aquamarine)).
						MoveTo(20, 50).
						LineTo(230, 50).
						Stroke(),
				)
			},
		},
		{
			name: "line_caps_and_joins",
			setup: func(t *testing.T, ctx *instructions.Layer) {
				ctx.LoadInstruction(
					instructions.NewLine().
						SetLineWidth(10).
						SetStrokePattern(colors.NewSolid(colors.MediumPurple)).
						SetLineCap(instructions.LineCapButt).MoveTo(20, 20).LineTo(120, 20).Stroke().
						SetLineCap(instructions.LineCapRound).MoveTo(20, 60).LineTo(120, 60).Stroke().
						SetLineCap(instructions.LineCapSquare).MoveTo(20, 100).LineTo(120, 100).Stroke(),
				)
			},
		},
		{
			name: "clip_preserve",
			setup: func(t *testing.T, ctx *instructions.Layer) {
				ctx.LoadInstructions(
					instructions.NewLine().
						SetFillPattern(colors.NewSolid(colors.IndianRed)).
						MoveTo(30, 30).
						LineTo(226, 30).
						LineTo(226, 226).
						LineTo(30, 226).
						ClosePath().
						ClipPreserve(),
					instructions.NewLine().
						SetFillPattern(colors.NewSolid(colors.Pumpkin)).
						MoveTo(50, 50).
						LineTo(150, 50).
						LineTo(150, 150).
						ClosePath().
						Fill(),
				)
			},
		},
		{
			name: "fill_rule_even_odd",
			setup: func(t *testing.T, ctx *instructions.Layer) {
				ctx.LoadInstruction(
					instructions.NewLine().
						SetFillRule(instructions.FillRuleEvenOdd).
						SetFillPattern(colors.NewSolid(colors.MidnightBlue)).
						MoveTo(50, 50).LineTo(200, 50).LineTo(200, 200).LineTo(50, 200).ClosePath().
						MoveTo(80, 80).LineTo(170, 80).LineTo(170, 170).LineTo(80, 170).ClosePath().
						Fill(),
				)
			},
		},
		{
			name: "complex_chain",
			setup: func(t *testing.T, ctx *instructions.Layer) {
				ctx.LoadInstruction(
					instructions.NewLine().
						SetLineWidth(5).
						SetDashes([]float64{15, 5}).
						SetStrokePattern(colors.NewSolid(colors.OrangeRed)).
						MoveTo(20, 230).
						QuadraticTo(128, 100, 230, 230).
						CubicTo(230, 230, 128, 128, 20, 230).
						ClosePath().
						Stroke(),
				)
			},
		},
	}

	for _, cse := range cases {
		t.Run(cse.name, func(t *testing.T) {
			layer := newLayer(t, 256, 256)
			require.NotNil(t, layer, "layer should not be nil")

			require.NotPanics(t, func() {
				cse.setup(t, layer)
			}, "setup should not panic")

			err := layer.Export("./output/line_" + cse.name + ".png")
			require.NoError(t, err, "export failed for %s", cse.name)
		})
	}
}
