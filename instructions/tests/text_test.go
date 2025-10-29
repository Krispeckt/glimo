package glimo_test

import (
	"testing"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/effects"
	"github.com/Krispeckt/glimo/instructions"
	"github.com/Krispeckt/glimo/internal/render"
	"github.com/stretchr/testify/require"
)

func TestInstructionText(t *testing.T) {
	bg, err := colors.HEX("#D9D9D9")
	require.NoError(t, err)

	font := render.MustLoadFont("testdata/montserrat.ttf", 72).SetLetterSpacingPercent(0)

	type testCase struct {
		name  string
		setup func(*testing.T, *instructions.Layer)
	}

	cases := []testCase{
		{
			name: "gradient_fill_with_shadow",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
					instructions.NewText(
						"1234567890 glimo Test Glimo Test Glimo Test Glimo Test Glimo Test Glimo Test Glimo Test Glimo Test Glimo Test",
						0, 150, font,
					).
						SetColorPattern(
							colors.NewLinearGradientWithBlend(0, 140, 1000, 140, colors.BlendNormal, 1).
								AddColorStop(0, colors.Amethyst).
								AddColorStop(1, colors.Pumpkin),
						).
						AddEffect(effects.NewDropShadow(0, 4, 4, 0, colors.Red, 1)).
						SetAlign(instructions.AlignTextCenter).
						SetMaxWidth(1000).
						SetMaxLines(3).
						SetWrapMode(instructions.WrapByWord).
						SetScaleStep(-24),
				)
			},
		},
		{
			name: "solid_fill_text",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
					instructions.NewRectangle(0, 150, 250, 200).SetFillColor(bg),
					instructions.NewText("Solid Color Text", 0, 150, font).
						SetColorPattern(colors.NewSolidWithBlend(colors.IndianRed, colors.BlendExclusion, 0.8)).
						SetAlign(instructions.AlignTextLeft).
						SetMaxWidth(1000),
				)
			},
		},
		{
			name: "multiline_wrapped_text",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
					instructions.NewText(
						"Glimo supports multiple lines and wraps text properly by symbol when width is limited.",
						0, 150, font,
					).
						SetColorPattern(colors.NewSolid(colors.Aquamarine)).
						SetAlign(instructions.AlignTextCenter).
						SetMaxWidth(600).
						SetMaxLines(3).
						SetWrapMode(instructions.WrapBySymbol),
				)
			},
		},
		{
			name: "right_aligned_text",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
					instructions.NewText("Right Aligned Example", 0, 150, font).
						SetColorPattern(colors.NewSolid(colors.MediumPurple)).
						SetAlign(instructions.AlignTextRight).
						SetMaxWidth(800),
				)
			},
		},
		{
			name: "stroke",
			setup: func(t *testing.T, c *instructions.Layer) {
				c.LoadInstructions(
					instructions.NewText("glimo Stroke Example", 0, 150, font).
						SetColorPattern(colors.NewSolid(colors.MediumPurple)).
						SetStrokeWithPattern(colors.NewSolid(colors.MintCream), 4).
						SetMaxWidth(800),
				)
			},
		},
	}

	for _, cse := range cases {
		t.Run(cse.name, func(t *testing.T) {
			canvas := newLayer(t, 1000, 500)
			require.NotNil(t, canvas, "canvas should not be nil")

			require.NotPanics(t, func() {
				cse.setup(t, canvas)
			}, "setup should not panic")

			outPath := "./output/text_" + cse.name + ".png"
			err := canvas.Export(outPath)
			require.NoError(t, err, "export failed for %s", cse.name)
		})
	}
}
