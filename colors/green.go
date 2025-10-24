package colors

import (
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

var (
	Green             = patterns.Color{R: 0, G: 255, B: 0, A: 255}     // #00FF00
	Honeydew          = patterns.Color{R: 240, G: 255, B: 240, A: 255} // #F0FFF0
	MintCream         = patterns.Color{R: 245, G: 255, B: 250, A: 255} // #F5FFFA
	PaleGreen         = patterns.Color{R: 152, G: 251, B: 152, A: 255} // #98FB98
	LightGreen        = patterns.Color{R: 144, G: 238, B: 144, A: 255} // #90EE90
	MediumSpringGreen = patterns.Color{R: 0, G: 250, B: 154, A: 255}   // #00FA9A
	SpringGreen       = patterns.Color{R: 0, G: 255, B: 127, A: 255}   // #00FF7F
	Chartreuse        = patterns.Color{R: 127, G: 255, B: 0, A: 255}   // #7FFF00
	LawnGreen         = patterns.Color{R: 124, G: 252, B: 0, A: 255}   // #7CFC00
	Lime              = patterns.Color{R: 0, G: 255, B: 0, A: 255}     // #00FF00
	LimeGreen         = patterns.Color{R: 50, G: 205, B: 50, A: 255}   // #32CD32
	YellowGreen       = patterns.Color{R: 154, G: 205, B: 50, A: 255}  // #9ACD32
	MediumSeaGreen    = patterns.Color{R: 60, G: 179, B: 113, A: 255}  // #3CB371
	SeaGreen          = patterns.Color{R: 46, G: 139, B: 87, A: 255}   // #2E8B57
	ForestGreen       = patterns.Color{R: 34, G: 139, B: 34, A: 255}   // #228B22
	DarkGreen         = patterns.Color{R: 0, G: 100, B: 0, A: 255}     // #006400
)

var GreenPalette = []patterns.Color{
	Green, Honeydew, MintCream, PaleGreen, LightGreen, MediumSpringGreen,
	SpringGreen, Chartreuse, LawnGreen, Lime, LimeGreen,
	YellowGreen, MediumSeaGreen, SeaGreen, ForestGreen, DarkGreen,
}
