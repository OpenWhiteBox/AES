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

// Invert computes the multiplicative inverse of a matrix, if it exists.
func (e Matrix) Invert() (Matrix, bool) {
	inv, _, frees := e.gaussJordan()
	return inv, len(frees) == 0
}

// FindPivot finds a row with non-zero entry in column col, starting at the given row and moving down. It returns the
// index of the row or -1 if one does not exist.
func (e Matrix) FindPivot(row, col int) int {
	out, _ := e.Size()

	for i := row; i < out; i++ {
		if !e[i][col].IsZero() {
			return i
		}
	}

	return -1
}

// Size returns the dimensions of the matrix in (Rows, Columns) order.
func (e Matrix) Size() (int, int) {
	if len(e) == 0 {
		return 0, 0
	} else {
		return len(e), e[0].Size()
	}
}

func (e Matrix) String() string {
	out := []rune{}

	for _, row := range e {
		out = append(out, []rune(row.String())...)
		out = append(out, '\n')
	}

	return string(out)
}
