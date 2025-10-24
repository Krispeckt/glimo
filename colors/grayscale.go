package colors

import (
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
)

var (
	Black       = patterns.Color{R: 0, G: 0, B: 0, A: 255}       // #000000
	EerieBlack  = patterns.Color{R: 27, G: 27, B: 27, A: 255}    // #1B1B1B
	Jet         = patterns.Color{R: 52, G: 52, B: 52, A: 255}    // #343434
	Onyx        = patterns.Color{R: 53, G: 56, B: 57, A: 255}    // #353839
	DavysGray   = patterns.Color{R: 85, G: 85, B: 85, A: 255}    // #555555
	DimGray     = patterns.Color{R: 105, G: 105, B: 105, A: 255} // #696969
	SonicSilver = patterns.Color{R: 117, G: 117, B: 117, A: 255} // #757575
	Gray        = patterns.Color{R: 128, G: 128, B: 128, A: 255} // #808080
	DarkGray    = patterns.Color{R: 169, G: 169, B: 169, A: 255} // #A9A9A9
	Silver      = patterns.Color{R: 192, G: 192, B: 192, A: 255} // #C0C0C0
	LightGray   = patterns.Color{R: 211, G: 211, B: 211, A: 255} // #D3D3D3
	Gainsboro   = patterns.Color{R: 220, G: 220, B: 220, A: 255} // #DCDCDC
	Platinum    = patterns.Color{R: 229, G: 228, B: 226, A: 255} // #E5E4E2
	WhiteSmoke  = patterns.Color{R: 245, G: 245, B: 245, A: 255} // #F5F5F5
	White       = patterns.Color{R: 255, G: 255, B: 255, A: 255} // #FFFFFF
)

var GrayscalePalette = []patterns.Color{
	Black, EerieBlack, Jet, Onyx, DavysGray,
	DimGray, SonicSilver, Gray, DarkGray, Silver,
	LightGray, Gainsboro, Platinum, WhiteSmoke, White,
}
