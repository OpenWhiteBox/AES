package gfmatrix

import (
	"github.com/OpenWhiteBox/AES/primitives/number"
)

// Row is a row / vector of elements from GF(2^8).
type Row []number.ByteFieldElem

// Add adds two vectors from GF(2^8)^n.
func (e Row) Add(f Row) Row {
	le, lf := e.Size(), f.Size()
	if le != lf {
		panic("Can't add rows that are different sizes!")
	}

	out := make([]number.ByteFieldElem, le)
	for i := 0; i < le; i++ {
		out[i] = e[i].Add(f[i])
	}

	return Row(out)
}

// ScalarMul multiplies a row by a scalar.
func (e Row) ScalarMul(f number.ByteFieldElem) Row {
	size := e.Size()

	out := make([]number.ByteFieldElem, size)
	for i := 0; i < size; i++ {
		out[i] = e[i].Mul(f)
	}

	return Row(out)
}

// DotProduct computes the dot product of two vectors.
func (e Row) DotProduct(f Row) number.ByteFieldElem {
	size := e.Size()
	if size != f.Size() {
		panic("Can't compute dot product of two vectors of different sizes!")
	}

	res := number.ByteFieldElem(0x00)
	for i := 0; i < size; i++ {
		res = res.Add(e[i].Mul(f[i]))
	}

	return res
}

// IsPermutation returns true if the row is a permutation of all the elements of GF(2^8) and false otherwise.
func (e Row) IsPermutation() bool {
	sums := [256]int{}
	for _, e_i := range e {
		sums[e_i]++
	}

	for _, x := range sums {
		if x != 1 {
			return false
		}
	}

	return true
}

// Height returns the position of the first non-zero entry in the row, or -1 if the row is zero.
func (e Row) Height() int {
	for i, e_i := range e {
		if !e_i.IsZero() {
			return i
		}
	}

	return -1
}

// Equal returns true if two rows are equal and false otherwise.
func (e Row) Equal(f Row) bool {
	if e.Size() != f.Size() {
		panic("Can't compare rows that are different sizes!")
	}

	for i := 0; i < e.Size(); i++ {
		if e[i] != f[i] {
			return false
		}
	}

	return true
}

// IsZero returns whether or not the row is identically zero.
func (e Row) IsZero() bool {
	for _, e_i := range e {
		if !e_i.IsZero() {
			return false
		}
	}

	return true
}

// Size returns the dimension of the vector.
func (e Row) Size() int {
	return len(e)
}
