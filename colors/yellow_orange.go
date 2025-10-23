package colors

import (
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

var (
	LightYellow          = patterns.Color{R: 255, G: 255, B: 224, A: 255} // #FFFFE0
	LemonChiffon         = patterns.Color{R: 255, G: 250, B: 205, A: 255} // #FFFACD
	Cornsilk             = patterns.Color{R: 255, G: 248, B: 220, A: 255} // #FFF8DC
	LightGoldenrodYellow = patterns.Color{R: 250, G: 250, B: 210, A: 255} // #FAFAD2
	PaleGoldenrod        = patterns.Color{R: 238, G: 232, B: 170, A: 255} // #EEE8AA
	Khaki                = patterns.Color{R: 240, G: 230, B: 140, A: 255} // #F0E68C
	Saffron              = patterns.Color{R: 244, G: 196, B: 48, A: 255}  // #F4C430
	Yellow               = patterns.Color{R: 255, G: 255, B: 0, A: 255}   // #FFFF00
	Gold                 = patterns.Color{R: 255, G: 215, B: 0, A: 255}   // #FFD700
	Amber                = patterns.Color{R: 255, G: 191, B: 0, A: 255}   // #FFBF00
	Goldenrod            = patterns.Color{R: 218, G: 165, B: 32, A: 255}  // #DAA520
	DarkGoldenrod        = patterns.Color{R: 184, G: 134, B: 11, A: 255}  // #B8860B
	Orange               = patterns.Color{R: 255, G: 165, B: 0, A: 255}   // #FFA500
	DarkOrange           = patterns.Color{R: 255, G: 140, B: 0, A: 255}   // #FF8C00
	Pumpkin              = patterns.Color{R: 255, G: 117, B: 24, A: 255}  // #FF7518
)

var YellowOrangePalette = []patterns.Color{
	LightYellow, LemonChiffon, Cornsilk, LightGoldenrodYellow, PaleGoldenrod,
	Khaki, Saffron, Yellow, Gold, Amber,
	Goldenrod, DarkGoldenrod, Orange, DarkOrange, Pumpkin,
}
