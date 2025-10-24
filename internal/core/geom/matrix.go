package geom

import "math"

// Matrix represents a 2D affine transformation.
// The matrix layout corresponds to the following form:
//
//	| XX  XY  X0 |
//	| YX  YY  Y0 |
//	|  0   0   1 |
//
// It allows translation, rotation, scaling, and shearing transformations.
type Matrix struct {
	XX, YX, XY, YY, X0, Y0 float64
}

// Identity returns the identity transformation matrix.
func Identity() Matrix {
	return Matrix{
		1, 0,
		0, 1,
		0, 0,
	}
}

// Translate returns a translation matrix that moves points by (x, y).
func Translate(x, y float64) Matrix {
	return Matrix{
		1, 0,
		0, 1,
		x, y,
	}
}

// Scale returns a scaling matrix that scales X and Y coordinates independently.
func Scale(x, y float64) Matrix {
	return Matrix{
		x, 0,
		0, y,
		0, 0,
	}
}

// Rotate returns a rotation matrix for the given angle in radians.
func Rotate(angle float64) Matrix {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return Matrix{
		c, s,
		-s, c,
		0, 0,
	}
}

// Shear returns a shearing matrix that skews the X and Y axes by the given factors.
func Shear(x, y float64) Matrix {
	return Matrix{
		1, y,
		x, 1,
		0, 0,
	}
}

// Operations

// Multiply returns the result of matrix multiplication a * b.
// The operation composes the transformation represented by b after a.
func (a Matrix) Multiply(b Matrix) Matrix {
	return Matrix{
		a.XX*b.XX + a.YX*b.XY,
		a.XX*b.YX + a.YX*b.YY,
		a.XY*b.XX + a.YY*b.XY,
		a.XY*b.YX + a.YY*b.YY,
		a.X0*b.XX + a.Y0*b.XY + b.X0,
		a.X0*b.YX + a.Y0*b.YY + b.Y0,
	}
}

// Transformations

// TransformVector applies the linear (rotation, scale, shear) part of the matrix
// to the vector (x, y) without translation.
func (a Matrix) TransformVector(x, y float64) (tx, ty float64) {
	tx = a.XX*x + a.XY*y
	ty = a.YX*x + a.YY*y
	return
}

// TransformPoint applies the full affine transformation (including translation)
// to the point (x, y).
func (a Matrix) TransformPoint(x, y float64) (tx, ty float64) {
	tx = a.XX*x + a.XY*y + a.X0
	ty = a.YX*x + a.YY*y + a.Y0
	return
}

// Composition Helpers

// Translate returns a new matrix obtained by applying a translation (x, y)
// before the existing transformation.
func (a Matrix) Translate(x, y float64) Matrix {
	return Translate(x, y).Multiply(a)
}

// Scale returns a new matrix obtained by applying a scaling (x, y)
// before the existing transformation.
func (a Matrix) Scale(x, y float64) Matrix {
	return Scale(x, y).Multiply(a)
}

// Rotate returns a new matrix obtained by applying a rotation (angle radians)
// before the existing transformation.
func (a Matrix) Rotate(angle float64) Matrix {
	return Rotate(angle).Multiply(a)
}

// Shear returns a new matrix obtained by applying a shearing (x, y)
// before the existing transformation.
func (a Matrix) Shear(x, y float64) Matrix {
	return Shear(x, y).Multiply(a)
}
