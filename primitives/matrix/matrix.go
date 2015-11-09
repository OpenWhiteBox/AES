// Basic operations on matrices in GF(2) and the random generation of new ones.
package matrix

import (
	"bytes"
	"fmt"
	"io"
)

var weight [4]uint64 = [4]uint64{
	0x6996966996696996, 0x9669699669969669,
	0x9669699669969669, 0x6996966996696996,
}

func rowsToColumns(x int) int {
	out := x / 8
	if x%8 != 0 {
		out++
	}

	return out
}

type Row []byte

// Add adds two vectors from GF(2)^n.
func (e Row) Add(f Row) Row {
	if len(e) != len(f) {
		panic("Can't add rows that are different sizes!")
	}

	out := make([]byte, len(e))
	for i := 0; i < len(e); i++ {
		out[i] = e[i] ^ f[i]
	}

	return Row(out)
}

// Mul component-wise multiplies two vectors.
func (e Row) Mul(f Row) Row {
	if len(e) != len(f) {
		panic("Can't multiply rows that are different sizes!")
	}

	out := make([]byte, len(e))
	for i := 0; i < len(e); i++ {
		out[i] = e[i] & f[i]
	}

	return Row(out)
}

// DotProduct computes the dot product of two vectors.
func (e Row) DotProduct(f Row) bool {
	parity := uint64(0)

	for _, g_i := range e.Mul(f) {
		parity ^= (weight[g_i/64] >> (g_i % 64)) & 1
	}

	return parity == 1
}

func (e Row) Weight() (w int) {
	for i := 0; i < e.Size(); i++ {
		if e.GetBit(i) == 1 {
			w += 1
		}
	}

	return
}

// GetBit returns the ith entry of the vector.
func (e Row) GetBit(i int) byte {
	return (e[i/8] >> (uint(i) % 8)) & 1
}

// SetBit sets the ith bit of the vector to 1 is x = true and 0 if x = false.
func (e Row) SetBit(i int, x bool) {
	y := e.GetBit(i)
	if y == 0 && x || y == 1 && !x {
		e[i/8] ^= 1 << (uint(i) % 8)
	}
}

// Size returns the dimension of the vector.
func (e Row) Size() int {
	return 8 * len(e)
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
	out := make([]Row, len(e))
	for i := 0; i < len(e); i++ {
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

// RightStretch returns the matrix of right matrix multiplication by a given matrix.
func (e Matrix) RightStretch() Matrix {
	n, m := e.Size()
	nm := n * m

	out := make([]Row, nm)

	for i := 0; i < nm; i++ {
		out[i] = make([]byte, nm/8)
		p, q := i/n, i%n

		for j := 0; j < m; j++ {
			out[i].SetBit(j*m+q, e[p].GetBit(j) == 1)
		}
	}

	return out
}

// LeftStretch returns the matrix of left matrix multiplication by a given matrix.
func (e Matrix) LeftStretch() Matrix {
	n, m := e.Size()
	nm := n * m

	out := make([]Row, nm)

	for i := 0; i < nm; i++ {
		out[i] = make([]byte, nm/8)
		p, q := i/n, i%n

		for j := 0; j < m; j++ {
			out[i].SetBit(j+m*p, e[j].GetBit(q) == 1)
		}
	}

	return out
}

// gaussJordan reduces the matrix according to the Gauss-Jordan Method.  Returns the transformed matrix, the augment
// matrix, at what column elimination failed (if ever), and whether or not the input is invertible (meaning the augment
// matrix is the inverse).
func (e Matrix) gaussJordan(lower, upper int) (Matrix, Matrix, int, bool) {
	a, b := e.Size()

	aug := GenerateIdentity(a) // The augmentation matrix for e.
	f := make([]Row, a)        // Duplicate e away so we don't mutate it.
	copy(f, e)

	for i, _ := range f[lower:upper] {
		row := i + lower

		if row >= b { // The matrix is tall and thin--we've finished before exhausting all the rows.
			break
		}

		// Find a row with a non-zero entry in the (col)th position
		candId := -1
		for j, f_j := range f[row:] {
			if f_j.GetBit(row) == 1 {
				candId = j + row
				break
			}
		}

		if candId == -1 && lower > 0 {
			for j, f_j := range f[:lower] {
				if f_j.GetBit(row) == 1 {
					candId = j
					break
				}
			}

			if candId == -1 {
				return f, aug, row, false
			}
		} else if candId == -1 { // If we can't find one and there's nowhere else, fail and return our partial work.
			return f, aug, row, false
		}

		// Move it to the top
		f[row], f[candId] = f[candId], f[row]
		aug[row], aug[candId] = aug[candId], aug[row]

		// Cancel out the (row)th position for every row above and below it.
		for i, _ := range f {
			if i != row && f[i].GetBit(row) == 1 {
				f[i] = f[i].Add(f[row])
				aug[i] = aug[i].Add(aug[row])
			}
		}
	}

	return f, aug, -1, true
}

// Invert computes the multiplicative inverse of a matrix, if it exists.
func (e Matrix) Invert() (Matrix, bool) {
	_, inv, _, ok := e.gaussJordan(0, len(e))
	return inv, ok
}

// InvertAt will return a partial inverse of a matrix, only exposing certain parts of the input.
func (e Matrix) InvertAt(positions ...int) (Matrix, bool) {
	if len(positions) == 0 {
		return GenerateIdentity(len(e)), true
	}

	eInv, ok := e.Invert()
	if !ok {
		return nil, ok
	}

	f, g1, _, ok := e.gaussJordan(8*positions[0], 8*(positions[0]+1))
	if !ok {
		return nil, ok
	}

	f = f.Transpose()
	f, g2, _, ok := f.gaussJordan(8*positions[0], 8*(positions[0]+1))
	if !ok {
		return nil, ok
	}

	g := g1.Compose(e).Compose(g2.Transpose()).Compose(eInv)

	out, ok := f.Transpose().InvertAt(positions[1:]...)
	if !ok {
		return nil, ok
	}

	return out.Compose(g), true
}

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
	f, _, c, ok := e.gaussJordan(0, len(e))

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

// Size returns the dimensions of the matrix in (Rows, Columns) order.
func (e Matrix) Size() (int, int) {
	return len(e), e[0].Size()
}

func (e Matrix) String() string {
	out := []rune{}

	for _, row := range e {
		for _, elem := range row {
			b := []rune(fmt.Sprintf("%8.8b", elem))

			for pos := 7; pos >= 0; pos-- {
				out = append(out, b[pos])
			}
		}

		out = append(out, '\n')
	}

	return string(out)
}

// GenerateIdentity creates the n by n identity matrix.
func GenerateIdentity(n int) Matrix {
	out := GenerateEmpty(n)
	for i := 0; i < n; i++ {
		out[i].SetBit(i, true)
	}

	return out
}

// GenerateFull creates a matrix with all entries set to 1.
func GenerateFull(n int) Matrix {
	out := GenerateEmpty(n)
	for i := 0; i < n; i++ {
		for j := 0; j < n/8; j++ {
			out[i][j] = 0xff
		}
	}

	return out
}

// GenerateEmpty creates a matrix with all entries set to 0.
func GenerateEmpty(n int) Matrix {
	out := make([]Row, n)

	for i := 0; i < n; i++ {
		out[i] = make([]byte, rowsToColumns(n))
	}

	return Matrix(out)
}

// GenerateRandom creates a random non-singular n by n matrix.
func GenerateRandom(reader io.Reader, n int) Matrix {
	m := GenerateTrueRandom(reader, n)
	_, ok := m.Invert()

	if ok { // Return this one or try again.
		return m
	} else {
		return GenerateRandom(reader, n) // Performance bottleneck.
	}

	return m
}

// GenerateTrueRandom creates a random singular or non-singular n by n matrix.
func GenerateTrueRandom(reader io.Reader, n int) Matrix {
	m := Matrix(make([]Row, n))

	for i := 0; i < n; i++ { // Generate random n x n matrix.
		row := Row(make([]byte, rowsToColumns(n)))
		reader.Read(row)

		m[i] = row
	}

	return m
}
