// Package gfmatrix implements basic operations on matrices over Rijndael's field and the random generation of new ones.
package gfmatrix

import (
	"github.com/OpenWhiteBox/AES/primitives/number"
)

// Matrix represents a GF(2^8)-matrix.
type Matrix []Row

// Mul right-multiplies a matrix by a row.
func (e Matrix) Mul(f Row) Row {
	out, in := e.Size()
	if in != f.Size() {
		panic("Can't multiply by row that is wrong size!")
	}

	res := Row(make([]number.ByteFieldElem, out))
	for i := 0; i < out; i++ {
		res[i] = e[i].DotProduct(f)
	}

	return res
}

// Transpose returns the transpose of a matrix.
func (e Matrix) Transpose() Matrix {
	n, m := e.Size()
	out := make([]Row, m)

	for i, _ := range out {
		out[i] = Row(make([]number.ByteFieldElem, n))

		for j := 0; j < n; j++ {
			out[i][j] = e[j][i]
		}
	}

	return Matrix(out)
}

// Size returns the dimensions of the matrix in (Rows, Columns) order.
func (e Matrix) Size() (int, int) {
	return len(e), len(e[0])
}
