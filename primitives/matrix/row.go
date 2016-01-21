package matrix

import (
	"fmt"
)

var weight [4]uint64 = [4]uint64{
	0x6996966996696996, 0x9669699669969669,
	0x9669699669969669, 0x6996966996696996,
}

// A binary row / vector in GF(2)^n.
type Row []byte

// NewRow returns an empty n-component row.
func NewRow(n int) Row {
	return Row(make([]byte, rowsToColumns(n)))
}

// LessThan returns true if row i is "less than" row j. If you use sort a permutation matrix according to LessThan,
// you'll always get the identity matrix.
func LessThan(i, j Row) bool {
	if i.Size() != j.Size() {
		panic("Can't compare rows that are different sizes!")
	}

	for k, _ := range i {
		if i[k] != 0x00 || j[k] != 0x00 {
			if i[k] == 0x00 {
				return false
			} else if j[k] == 0x00 {
				return true
			} else if i[k]&-i[k] < j[k]&-j[k] {
				return true
			} else {
				return false
			}
		}
	}

	return false
}

// Add adds (XORs) two vectors.
func (e Row) Add(f Row) Row {
	le, lf := len(e), len(f)
	if le != lf {
		panic("Can't add rows that are different sizes!")
	}

	out := make([]byte, le)
	for i := 0; i < le; i++ {
		out[i] = e[i] ^ f[i]
	}

	return Row(out)
}

// Mul component-wise multiplies (ANDs) two vectors.
func (e Row) Mul(f Row) Row {
	le, lf := len(e), len(f)
	if le != lf {
		panic("Can't multiply rows that are different sizes!")
	}

	out := make([]byte, le)
	for i := 0; i < le; i++ {
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

// Weight returns the hamming weight of this row.
func (e Row) Weight() (w int) {
	for i := 0; i < e.Size(); i++ {
		if e.GetBit(i) == 1 {
			w += 1
		}
	}

	return
}

// Returns true if e should be used to cancel out a bit in f.
func (e Row) Cancels(f Row) bool {
	for i, _ := range e {
		if e[i] != 0x00 {
			if e[i]&-e[i]&f[i] != 0x00 {
				return true
			} else {
				return false
			}
		}
	}

	return false
}

// GetBit returns the ith component of the vector: 0x00 or 0x01.
func (e Row) GetBit(i int) byte {
	return (e[i/8] >> (uint(i) % 8)) & 1
}

// SetBit sets the ith component of the vector to 0x01 is x = true and 0x00 if x = false.
func (e Row) SetBit(i int, x bool) {
	y := e.GetBit(i)
	if y == 0 && x || y == 1 && !x {
		e[i/8] ^= 1 << (uint(i) % 8)
	}
}

// IsZero returns true if the row is identically zero.
func (e Row) IsZero() bool {
	for _, e_i := range e {
		if e_i != 0x00 {
			return false
		}
	}

	return true
}

// Height returns the position of the first non-zero entry in the row, or -1 if the row is zero.
func (e Row) Height() int {
	for i := 0; i < e.Size(); i++ {
		if e.GetBit(i) == 1 {
			return i
		}
	}

	return -1
}

// Size returns the dimension of the vector.
func (e Row) Size() int {
	return 8 * len(e)
}

// Dup returns a duplicate of this row.
func (e Row) Dup() Row {
	f := Row(make([]byte, len(e)))
	copy(f, e)

	return f
}

// String converts the row into space-and-dot notation.
func (e Row) String() string {
	out := []rune{'|'}

	for _, elem := range e {
		b := []rune(fmt.Sprintf("%8.8b", elem))

		for pos := 7; pos >= 0; pos-- {
			if b[pos] == '0' {
				out = append(out, ' ')
			} else {
				out = append(out, '•')
			}
		}
	}

	return string(append(out, '|', '\n'))
}
