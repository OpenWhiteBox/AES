package matrix2

import (
	"github.com/OpenWhiteBox/AES/primitives/number"
)

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
