package glimo_test

import (
	_ "image/png"
	"testing"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/effects"
	"github.com/Krispeckt/glimo/instructions"
	"github.com/stretchr/testify/require"
)

func TestInstructionImage(t *testing.T) {
	src := mustLoadImage(t, "./testdata/image.png")

	type tc struct {
		name string
		w, h int
		cfg  func(*instructions.Image)
	}

	cases := []tc{
		{
			name: "image_contain",
			w:    800, h: 600,
			cfg: func(im *instructions.Image) {
				im.SetSize(800, 600).SetFit(instructions.FitContain)
			},
		},
		{
			name: "image_cover",
			w:    800, h: 600,
			cfg: func(im *instructions.Image) {
				im.SetSize(800, 600).SetFit(instructions.FitCover)
			},
		},
		{
			name: "image_stretch",
			w:    800, h: 600,
			cfg: func(im *instructions.Image) {
				im.SetSize(800, 600).SetFit(instructions.FitStretch)
			},
		},
		{
			name: "image_flipH",
			w:    600, h: 600,
			cfg: func(im *instructions.Image) {
				im.SetSize(600, 600).Mirror(true, false)
			},
		},
		{
			name: "image_flipV",
			w:    600, h: 600,
			cfg: func(im *instructions.Image) {
				im.SetSize(600, 600).Mirror(false, true)
			},
		},
		{
			name: "image_rotate_expand",
			w:    600, h: 600,
			cfg: func(im *instructions.Image) {
				im.SetSize(400, 400).Rotate(33).SetExpand(true).SetBackground(colors.Transparent)
			},
		},
		{
			name: "image_rotate_crop",
			w:    400, h: 400,
			cfg: func(im *instructions.Image) {
				im.SetSize(400, 400).Rotate(33).SetExpand(false).SetBackground(colors.Transparent)
			},
		},
		{
			name: "image_opacity_shadow",
			w:    900, h: 700,
			cfg: func(im *instructions.Image) {
				im.SetSize(800, 600).
					SetOpacity(0.5).
					AddEffect(effects.NewDropShadow(0, 4, 12, 0, colors.Black, 1))
			},
		},
		{
			name: "image_with_mask_rectangle",
			w:    800, h: 600,
			cfg: func(im *instructions.Image) {
				im.SetSize(800, 600).
					SetFit(instructions.FitContain).
					SetMaskFromShape(
						instructions.NewRectangle(0, 150, 250, 200).
							SetFillColor(colors.Black),
					)
			},
		},
	}

	for _, cse := range cases {
		t.Run(cse.name, func(t *testing.T) {
			layer := newLayer(t, cse.w, cse.h)
			require.NotNil(t, layer, "layer should not be nil")

			im := instructions.NewImage(src, 0, 0)
			require.NotNil(t, im, "image instruction should not be nil")

			cse.cfg(im)
			require.NotPanics(t, func() {
				layer.LoadInstruction(im)
			}, "AddInstructions should not panic")

			err := layer.Export("./output/line_" + cse.name + ".png")
			require.NoError(t, err, "export failed for %s", cse.name)
		})
	}
}
