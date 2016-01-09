package matrix

import (
	"fmt"
)

var weight [4]uint64 = [4]uint64{
	0x6996966996696996, 0x9669699669969669,
	0x9669699669969669, 0x6996966996696996,
}

type Row []byte

// Add adds two vectors from GF(2)^n.
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

// Mul component-wise multiplies two vectors.
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

// Dup returns a duplicate of this row.
func (e Row) Dup() Row {
	f := Row(make([]byte, len(e)))
	copy(f, e)

	return f
}

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
