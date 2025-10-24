package colors

import (
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

var (
	Blue           = patterns.Color{R: 0, G: 0, B: 255, A: 255}     // #0000FF
	PowderBlue     = patterns.Color{R: 176, G: 224, B: 230, A: 255} // #B0E0E6
	LightBlue      = patterns.Color{R: 173, G: 216, B: 230, A: 255} // #ADD8E6
	SkyBlue        = patterns.Color{R: 135, G: 206, B: 235, A: 255} // #87CEEB
	DeepSkyBlue    = patterns.Color{R: 0, G: 191, B: 255, A: 255}   // #00BFFF
	CornflowerBlue = patterns.Color{R: 100, G: 149, B: 237, A: 255} // #6495ED
	DodgerBlue     = patterns.Color{R: 30, G: 144, B: 255, A: 255}  // #1E90FF
	SteelBlue      = patterns.Color{R: 70, G: 130, B: 180, A: 255}  // #4682B4
	RoyalBlue      = patterns.Color{R: 65, G: 105, B: 225, A: 255}  // #4169E1
	CobaltBlue     = patterns.Color{R: 0, G: 71, B: 171, A: 255}    // #0047AB
	Sapphire       = patterns.Color{R: 15, G: 82, B: 186, A: 255}   // #0F52BA
	Ultramarine    = patterns.Color{R: 18, G: 10, B: 143, A: 255}   // #120A8F
	MediumBlue     = patterns.Color{R: 0, G: 0, B: 205, A: 255}     // #0000CD
	Navy           = patterns.Color{R: 0, G: 0, B: 128, A: 255}     // #000080
	MidnightBlue   = patterns.Color{R: 25, G: 25, B: 112, A: 255}   // #191970
	Aquamarine     = patterns.Color{R: 127, G: 255, B: 212, A: 255} // #7FFFD4
)

var BluePalette = []patterns.Color{
	Blue, PowderBlue, LightBlue, SkyBlue, DeepSkyBlue, CornflowerBlue,
	DodgerBlue, SteelBlue, RoyalBlue, CobaltBlue, Sapphire,
	Ultramarine, MediumBlue, Navy, MidnightBlue, Aquamarine,
}
