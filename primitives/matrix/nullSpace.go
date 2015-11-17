// Methods related to finding the nullspace of a matrix.
package matrix

import (
	"bytes"
)

// NullSpace returns one non-trivial element of the matrix's nullspace.
func (e Matrix) NullSpace() Row {
	a, b := e.Size()

	zero, full := Row(make([]byte, rowsToColumns(a))), Row(make([]byte, b/8))
	for i, _ := range full {
		full[i] = 0xff
	}

	if bytes.Compare(e.Mul(full), zero) == 0 {
		return full
	}

	// Apply Gauss-Jordan Elimination to the matrix to get it in a simpler form.
	f, c, ok := e.gaussJordan(GenerateIdentity(a), 0, a, IgnoreNoRows)

	if ok { // If the matrix is invertible, we can only return 0.
		return Row(make([]byte, b/8))
	}

	// Find an element in the nullspace of the failing sub-matrix.
	left, right := f.Slice(c + 1) // c+1 because the cth column is always all zero.
	low, high := right[c:].NullSpace(), []byte{}

	if c != 0 {
		// Calculate the "weight" of the right side of the matrix.  If it's even/odd, we need the corresponding bit in the
		// output vector set to 0/1 so the left side of the matrix (a square identity) cancels it out.
		high = right[:c].Mul(low).Add(left[0:c].Transpose()[c])
	}

	// The output vector is high || 1 || low, so the upper left and right cancel each other, the bottom right is zero
	// because we found an element in its nullspace, and the bottom left is zero because its a zero matrix.
	out := Row(make([]byte, b/8))
	copy(out, high)
	out.SetBit(c, true)

	for i := c + 1; i < b; i++ {
		out.SetBit(i, low.GetBit(i-c-1) == 1)
	}

	return out
}

// Slice cuts the matrix in half, down the given column. It returns the left and right halves.
func (e Matrix) Slice(col int) (Matrix, Matrix) {
	a, b := e.Size()

	left, right := make([]Row, a), make([]Row, a)

	for i, row := range e {
		left[i], right[i] = make([]byte, rowsToColumns(col)), make([]byte, rowsToColumns(b-col))

		for j := 0; j < col; j++ {
			left[i].SetBit(j, row.GetBit(j) == 1)
		}

		for j := col; j < b; j++ {
			right[i].SetBit(j-col, row.GetBit(j) == 1)
		}
	}
	return Matrix(left), Matrix(right)
}
