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

// Invert computes the two-sided multiplicative inverse of a matrix, if it exists.
func (e Matrix) Invert() (Matrix, bool) { // Gauss-Jordan Method
	a, b := e.Size()
	if a != b {
		panic("Can't invert a non-square matrix!")
	}

	out := GenerateIdentity(a) // The augmentation matrix for e. Will be mutated into e's inverse.

	f := make([]Row, a) // Duplicate e away so we don't mutate it while we're turning it into the identity.
	copy(f, e)

	for row := 0; row < a; row++ {
		// Find a row with a non-zero entry (a 1) in the (row)th position
		candId := 255

		for i := row; i < a; i++ {
			if f[i].GetBit(row) == 1 {
				candId = i
				break
			}
		}

		if candId == 255 { // If we can't find one, the matrix isn't invertible.
			return out, false
		}

		// Move it to the top
		f[row], f[candId] = f[candId], f[row]
		out[row], out[candId] = out[candId], out[row]

		// Cancel out the (row)th position for every row above and below it.
		for i := 0; i < a; i++ {
			if i == row {
				continue
			}

			if f[i].GetBit(row) == 1 {
				f[i] = f[i].Add(f[row])
				out[i] = out[i].Add(out[row])
			}
		}
	}

	return out, true
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
	out := make([]Row, n)
	for i := 0; i < n; i++ {
		out[i] = make([]byte, n/8)
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
