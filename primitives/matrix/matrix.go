// Basic operations on matrices in GF(2) and the random generation of new ones.
package matrix

import (
	"io"
)

var weight [4]uint64 = [4]uint64{
	0x6996966996696996, 0x9669699669969669,
	0x9669699669969669, 0x6996966996696996,
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

	res := Row(make([]byte, out/8))
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
func (e Matrix) gaussJordan() (Matrix, Matrix, int, bool) {
	a, _ := e.Size()

	aug := GenerateIdentity(a) // The augmentation matrix for e.
	f := make([]Row, a)        // Duplicate e away so we don't mutate it.
	copy(f, e)

	for row, _ := range f {
		// Find a row with a non-zero entry in the (col)th position
		candId := -1
		for i, f_i := range f[row:] {
			if f_i.GetBit(row) == 1 {
				candId = i + row
				break
			}
		}

		if candId == -1 { // If we can't find one, fail and return our partial work.
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
	_, inv, _, ok := e.gaussJordan()
	return inv, ok
}

// Transpose returns the transpose of a matrix.
func (e Matrix) Transpose() Matrix {
	n, m := e.Size()
	out := make([]Row, m)

	for i, _ := range out {
		out[i] = Row(make([]byte, n/8))

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

	lSize, rSize := col/8, (b-col)/8
	if col%8 != 0 {
		lSize, rSize = lSize+1, rSize+1
	}

	for i, row := range e {
		left[i], right[i] = make([]byte, lSize), make([]byte, rSize)

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
	out, cols := make([]Row, n), n/8
	if n%8 != 0 {
		cols++
	}

	for i := 0; i < n; i++ {
		out[i] = make([]byte, cols)
	}

	return Matrix(out)
}

// GenerateRandom creates a random non-singular n by n matrix.
func GenerateRandom(reader io.Reader, n int) Matrix {
	m := Matrix(make([]Row, n))

	for i := 0; i < n; i++ { // Generate random n x n matrix.
		row := Row(make([]byte, n/8))
		reader.Read(row)

		m[i] = row
	}

	_, ok := m.Invert()

	if ok { // Return this one or try again.
		return m
	} else {
		return GenerateRandom(reader, n) // Performance bottleneck.
	}

	return m
}
