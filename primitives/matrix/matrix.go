// Basic operations on matrices in GF(2) and the random generation of new ones.
package matrix

func rowsToColumns(x int) int {
	out := x / 8
	if x%8 != 0 {
		out++
	}

	return out
}

// Logical, or (0, 1)-Matrices
type Matrix []Row

// Mul right-multiplies a matrix by a row.
func (e Matrix) Mul(f Row) Row {
	out, in := e.Size()
	if in != f.Size() {
		panic("Can't multiply by row that is wrong size!")
	}

	res := Row(make([]byte, rowsToColumns(out)))
	for i := 0; i < out; i++ {
		if e[i].DotProduct(f) {
			res.SetBit(i, true)
		}
	}

	return res
}

// Add adds two binary matrices from GF(2)^nxm.
func (e Matrix) Add(f Matrix) Matrix {
	a, _ := e.Size()

	out := make([]Row, a)
	for i := 0; i < a; i++ {
		out[i] = e[i].Add(f[i])
	}

	return out
}

// Compose returns the result of composing e with f.
func (e Matrix) Compose(f Matrix) Matrix {
	n, m := e.Size()
	p, q := f.Size()

	if m != p {
		panic("Can't multiply matrices of wrong size!")
	}

	out := make([]Row, n)
	g := f.Transpose()

	for i := 0; i < n; i++ {
		out[i] = Row(make([]byte, q/8))

		for j := 0; j < q; j++ {
			out[i].SetBit(j, e[i].DotProduct(g[j]))
		}
	}

	return Matrix(out)
}

// Invert computes the multiplicative inverse of a matrix, if it exists.
func (e Matrix) Invert() (Matrix, bool) {
	a, _ := e.Size()

	inv := GenerateIdentity(a)
	_, _, ok := e.gaussJordan(inv, 0, a, IgnoreNoRows)
	return inv, ok
}

// Transpose returns the transpose of a matrix.
func (e Matrix) Transpose() Matrix {
	n, m := e.Size()
	out := make([]Row, m)

	for i, _ := range out {
		out[i] = Row(make([]byte, rowsToColumns(n)))

		for j := 0; j < n; j++ {
			out[i].SetBit(j, e[j].GetBit(i) == 1)
		}
	}

	return Matrix(out)
}

// Trace returns the trace (sum of elements on the diagonal) of a matrix.
func (e Matrix) Trace() (out byte) {
	n, _ := e.Size()
	for i := 0; i < n; i++ {
		out ^= (e[i][0] >> uint(i)) & 1
	}

	return
}

// Dup returns a duplicate of this matrix.
func (e Matrix) Dup() Matrix {
	f := make([]Row, len(e))
	copy(f, e)

	return f
}

// Size returns the dimensions of the matrix in (Rows, Columns) order.
func (e Matrix) Size() (int, int) {
	return len(e), e[0].Size()
}

func (e Matrix) String() string {
	out := []rune{}
	_, b := e.Size()

	addBar := func() {
		for i := -2; i < b; i++ {
			out = append(out, '-')
		}
		out = append(out, '\n')
	}

	addBar()
	for _, row := range e {
		out = append(out, []rune(row.String())...)
	}
	addBar()

	return string(out)
}
