package colors

import (
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

var (
	Lavender      = patterns.Color{R: 230, G: 230, B: 250, A: 255} // #E6E6FA
	Thistle       = patterns.Color{R: 216, G: 191, B: 216, A: 255} // #D8BFD8
	Plum          = patterns.Color{R: 221, G: 160, B: 221, A: 255} // #DDA0DD
	Orchid        = patterns.Color{R: 218, G: 112, B: 214, A: 255} // #DA70D6
	MediumOrchid  = patterns.Color{R: 186, G: 85, B: 211, A: 255}  // #BA55D3
	Violet        = patterns.Color{R: 238, G: 130, B: 238, A: 255} // #EE82EE
	MediumPurple  = patterns.Color{R: 147, G: 112, B: 219, A: 255} // #9370DB
	BlueViolet    = patterns.Color{R: 138, G: 43, B: 226, A: 255}  // #8A2BE2
	Amethyst      = patterns.Color{R: 153, G: 102, B: 204, A: 255} // #9966CC
	SlateBlue     = patterns.Color{R: 106, G: 90, B: 205, A: 255}  // #6A5ACD
	RebeccaPurple = patterns.Color{R: 102, G: 51, B: 153, A: 255}  // #663399
	DarkOrchid    = patterns.Color{R: 153, G: 50, B: 204, A: 255}  // #9932CC
	DarkViolet    = patterns.Color{R: 148, G: 0, B: 211, A: 255}   // #9400D3
	Purple        = patterns.Color{R: 128, G: 0, B: 128, A: 255}   // #800080
	Indigo        = patterns.Color{R: 75, G: 0, B: 130, A: 255}    // #4B0082
)

var PurplePalette = []patterns.Color{
	Lavender, Thistle, Plum, Orchid, MediumOrchid,
	Violet, MediumPurple, BlueViolet, Amethyst, SlateBlue,
	RebeccaPurple, DarkOrchid, DarkViolet, Purple, Indigo,
}
