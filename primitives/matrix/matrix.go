// Basic operations on matrices in GF(2) and the random generation of new ones.
package matrix

import (
	"io"
)

type Row []byte

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

func (e Row) DotProduct(f Row) bool {
	if e.Mul(f).Weight()%2 == 1 {
		return true
	} else {
		return false
	}
}

func (e Row) Weight() (weight int) {
	for _, e_i := range e {
		for j := uint(0); j < 8; j++ {
			if (e_i>>j)&1 == 1 {
				weight += 1
			}
		}
	}

	return
}

func (e Row) Size() int {
	return 8 * len(e)
}

type Matrix []Row

func (e Matrix) Mul(f Row) Row {
	out, in := e.Size()
	if in != f.Size() {
		panic("Can't multiply by row that is wrong size!")
	}

	n := uint(out)

	res := make([]byte, n/8)
	for i := uint(0); i < n; i++ {
		if e[i].DotProduct(f) {
			res[i/8] |= 1<<(i%8)
		}
	}

	return Row(res)
}

func (e Matrix) Add(f Matrix) Matrix {
	out := make([]Row, len(e))
	for i := 0; i < len(e); i++ {
		out[i] = e[i].Add(f[i])
	}

	return out
}

func (e Matrix) Invert() (Matrix, bool) { // Gauss-Jordan Method
	a, b := e.Size()
	if a != b {
		panic("Can't invert a non-square matrix!")
	}

	n := uint(a)

	out := make([]Row, n) // Identity matrix
	for i := uint(0); i < n; i++ {
		row := make([]byte, n/8)
		row[i/8] += 1<<(i%8)

		out[i] = row
	}

	f := make([]Row, n) // Duplicate e away so we don't mutate it.
	copy(f, e)

	for row := uint(0); row < n; row++ {
		// Find a row with a non-zero entry (a 1) in the (row)th position
		candId := uint(255)

		for i := row; i < n; i++ {
			if (f[i][row/8]>>(row%8))&1 == 1 {
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
		for i := uint(0); i < n; i++ {
			if i == row {
				continue
			}

			if (f[i][row/8]>>(row%8))&1 ==1 {
				f[i] = f[i].Add(f[row])
				out[i] = out[i].Add(out[row])
			}
		}
	}

	return out, true
}

func (e Matrix) Size() (int, int) {
	return len(e), e[0].Size()
}

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
