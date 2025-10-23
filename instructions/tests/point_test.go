package glimo_test

import (
	"testing"

	"github.com/Krispeckt/glimo/colors"
	"github.com/Krispeckt/glimo/instructions"
	"github.com/stretchr/testify/require"
)

func TestInstructionPoint(t *testing.T) {
	c := newLayer(t, 100, 100)
	require.NotNil(t, c, "layer should not be nil")

	require.NotPanics(t, func() {
		c.LoadInstruction(instructions.NewPoint(10, 10).SetColor(colors.Blue))
	}, "LoadInstructions should not panic")

	err := c.Export("./output/point_test.png")
	require.NoError(t, err, "export should succeed")
}
