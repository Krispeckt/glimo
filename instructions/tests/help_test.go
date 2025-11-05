package glimo_test

import (
	"image"
	"os"
	"testing"

	"github.com/Krispeckt/glimo/instructions"
	"github.com/stretchr/testify/require"
)

func mustLoadImage(t *testing.T, p string) image.Image {
	t.Helper()
	f, err := os.Open(p)
	require.NoError(t, err)

	defer func() {
		_ = f.Close()
	}()

	img, _, err := image.Decode(f)
	require.NoError(t, err)
	return img
}

func newLayer(t *testing.T, w, h int) *instructions.Layer {
	t.Helper()
	return instructions.NewLayer(w, h)
}
