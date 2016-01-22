// Package matrix implements basic operations on matrices in GF(2) and the random generation of new ones.
package matrix

func rowsToColumns(x int) int {
	out := x / 8
	if x%8 != 0 {
		out++
	}

	return out
}

// Matrix is a logical, or (0, 1)-matrix
type Matrix []Row

// Mul right-multiplies a matrix by a row.
func (e Matrix) Mul(f Row) Row {
	out, in := e.Size()
	if in != f.Size() {
		panic("Can't multiply by row that is wrong size!")
	}

	res := NewRow(out)
	for i, row := range e {
		if row.DotProduct(f) {
			res.SetBit(i, true)
		}
	}

	return res
}

// Add adds two binary matrices from GF(2)^nxm.
func (e Matrix) Add(f Matrix) Matrix {
	a, _ := e.Size()

	out := make([]Row, a)
	for i, _ := range out {
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

	out := GenerateEmpty(n, q)
	g := f.Transpose()

	for i, e_i := range e {
		for j, g_j := range g {
			out[i].SetBit(j, e_i.DotProduct(g_j))
		}
	}

	return out
}

// Invert computes the multiplicative inverse of a matrix, if it exists.
func (e Matrix) Invert() (Matrix, bool) {
	inv, _, frees := e.gaussJordan()
	return inv, len(frees) == 0
}

// Transpose returns the transpose of a matrix.
func (e Matrix) Transpose() Matrix {
	n, m := e.Size()
	out := GenerateEmpty(m, n)

	for i, _ := range out {
		for j := 0; j < n; j++ {
			out[i].SetBit(j, e[j].GetBit(i) == 1)
		}
	}

	return out
}

// Trace returns the trace (sum/parity of elements on the diagonal) of a matrix: 0x00 or 0x01.
func (e Matrix) Trace() (out byte) {
	for i, e_i := range e {
		out ^= e_i.GetBit(i)
	}

	return
}

// FindPivot finds a row with non-zero entry in column col, starting at the given row and moving down. It returns the
// index of the given row or -1 if one does not exist.
func (e Matrix) FindPivot(row, col int) int {
	for i, e_i := range e[row:] {
		if e_i.GetBit(col) == 1 {
			return row + i
		}
	}

	return -1
}

// Dup returns a duplicate of this matrix.
func (e Matrix) Dup() Matrix {
	n, m := e.Size()
	f := GenerateEmpty(n, m)

	for i, _ := range e {
		copy(f[i], e[i])
	}

	return f
}

// Size returns the dimensions of the matrix in (Rows, Columns) order.
func (e Matrix) Size() (int, int) {
	if len(e) == 0 {
		return 0, 0
	} else {
		return len(e), e[0].Size()
	}
}

// String converts the matrix to space-and-dot notation.
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
