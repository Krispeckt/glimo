package geom

import (
	"math"
	"math/rand/v2"
)

// SinNoise generates a smooth procedural pattern based on sine and cosine functions.
//
// The function combines multiple sine and cosine terms with random phase offsets
// (via math/rand/v2) to reduce visible repetition and tiling artifacts.
//
// Output range: approximately [-1, 1].
// Note: The pattern is non-deterministic unless the caller seeds the global RNG.
func SinNoise(x, y float64) float64 {
	return math.Sin(x*2.1+rand.Float64())*math.Cos(y*2.3+rand.Float64())*0.5 +
		math.Sin(x*1.7+0.5)*math.Sin(y*1.3+1.5)*0.5
}
