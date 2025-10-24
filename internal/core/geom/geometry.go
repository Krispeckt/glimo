package geom

import "math"

// Deg2Rad converts degrees to radians.
func Deg2Rad(deg float64) float64 { return deg * math.Pi / 180 }

// RotatedBounds calculates the bounding box size (nw, nh)
// of a rectangle with width w and height h after rotation by angleRad radians.
func RotatedBounds(w, h int, angleRad float64) (nw, nh int) {
	s := math.Abs(math.Sin(angleRad))
	c := math.Abs(math.Cos(angleRad))
	nw = int(math.Ceil(float64(w)*c + float64(h)*s))
	nh = int(math.Ceil(float64(w)*s + float64(h)*c))
	return
}
