package colors

import (
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

var (
	MistyRose  = patterns.Color{R: 255, G: 228, B: 225, A: 255} // #FFE4E1
	LightCoral = patterns.Color{R: 240, G: 128, B: 128, A: 255} // #F08080
	Salmon     = patterns.Color{R: 250, G: 128, B: 114, A: 255} // #FA8072
	Coral      = patterns.Color{R: 255, G: 127, B: 80, A: 255}  // #FF7F50
	Tomato     = patterns.Color{R: 255, G: 99, B: 71, A: 255}   // #FF6347
	OrangeRed  = patterns.Color{R: 255, G: 69, B: 0, A: 255}    // #FF4500
	Red        = patterns.Color{R: 255, G: 0, B: 0, A: 255}     // #FF0000
	Vermilion  = patterns.Color{R: 227, G: 66, B: 52, A: 255}   // #E34234
	Scarlet    = patterns.Color{R: 255, G: 36, B: 0, A: 255}    // #FF2400
	Crimson    = patterns.Color{R: 220, G: 20, B: 60, A: 255}   // #DC143C
	IndianRed  = patterns.Color{R: 205, G: 92, B: 92, A: 255}   // #CD5C5C
	FireBrick  = patterns.Color{R: 178, G: 34, B: 34, A: 255}   // #B22222
	DarkRed    = patterns.Color{R: 139, G: 0, B: 0, A: 255}     // #8B0000
	Carmine    = patterns.Color{R: 150, G: 0, B: 24, A: 255}    // #960018
	Burgundy   = patterns.Color{R: 128, G: 0, B: 32, A: 255}    // #800020
)

var RedPalette = []patterns.Color{
	MistyRose, LightCoral, Salmon, Coral, Tomato,
	OrangeRed, Red, Vermilion, Scarlet, Crimson,
	IndianRed, FireBrick, DarkRed, Carmine, Burgundy,
}
