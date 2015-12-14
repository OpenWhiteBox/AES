package matrix2

import (
	"github.com/OpenWhiteBox/AES/primitives/number"
)

// NullSpace returns one non-trivial element of the matrix's nullspace.
func (e Matrix) NullSpace() Row {
	_, in := e.Size()

	full := Row(make([]number.ByteFieldElem, in))
	for i, _ := range full {
		full[i] = number.ByteFieldElem(0x01)
	}

	if e.Mul(full).IsZero() {
		return full
	}

	// Apply Gauss-Jordan Elimination to the matrix to get it in a simpler form.
	_, f, c, ok := e.gaussJordan()

	if ok { // If the matrix is invertible, we can only return 0.
		return Row(make([]number.ByteFieldElem, in))
	}

	// Find an element in the nullspace of the failing sub-matrix.
	left, right := f.Slice(c + 1) // c+1 because the cth column is always all zero.
	low, high := right[c:].NullSpace(), []number.ByteFieldElem{}

	if c != 0 {
		// Calculate the "weight" of the right side of the matrix.  If it's even/odd, we need the corresponding bit in the
		// output vector set to 0/1 so the left side of the matrix (a square identity) cancels it out.
		high = right[:c].Mul(low).Add(left[0:c].Transpose()[c])
	}

	// The output vector is high || 1 || low, so the upper left and right cancel each other, the bottom right is zero
	// because we found an element in its nullspace, and the bottom left is zero because its a zero matrix.
	res := Row(make([]number.ByteFieldElem, in))
	copy(res, high)
	res[c] = number.ByteFieldElem(0x01)

	for i := c + 1; i < in; i++ {
		res[i] = low[i-c-1]
	}

	return res
}

// Slice cuts the matrix in half, down the given column. It returns the left and right halves.
func (e Matrix) Slice(col int) (Matrix, Matrix) {
	a, b := e.Size()

	left, right := make([]Row, a), make([]Row, a)

	for i, row := range e {
		left[i], right[i] = make([]number.ByteFieldElem, col), make([]number.ByteFieldElem, b-col)

		for j := 0; j < col; j++ {
			left[i][j] = row[j]
		}

		for j := col; j < b; j++ {
			right[i][j-col] = row[j]
		}
	}
	return Matrix(left), Matrix(right)
}
