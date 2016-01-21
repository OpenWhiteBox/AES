// Package gfmatrix implements basic operations on matrices over Rijndael's field and the random generation of new ones.
package gfmatrix

// Matrix represents a GF(2^8)-matrix.
type Matrix []Row

// Mul right-multiplies a matrix by a row.
func (e Matrix) Mul(f Row) Row {
	out, in := e.Size()
	if in != f.Size() {
		panic("Can't multiply by row that is wrong size!")
	}

	res := NewRow(out)
	for i := 0; i < out; i++ {
		res[i] = e[i].DotProduct(f)
	}

	return res
}

// Transpose returns the transpose of a matrix.
func (e Matrix) Transpose() Matrix {
	n, m := e.Size()
	out := GenerateEmpty(m, n)

	for i, row := range e {
		for j, elem := range row {
			out[j][i] = elem.Dup()
		}
	}

	return out
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

// Dup returns a duplicate of this matrix.
func (e Matrix) Dup() Matrix {
	n, m := e.Size()
	out := GenerateEmpty(n, m)

	for i, row := range e {
		for j, elem := range row {
			out[i][j] = elem.Dup()
		}
	}

	return out
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
