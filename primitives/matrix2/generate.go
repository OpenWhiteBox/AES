package matrix2

import (
	"github.com/OpenWhiteBox/AES/primitives/number"
)

// GenerateIdentity creates the n by n identity matrix.
func GenerateIdentity(n int) Matrix {
	out := GenerateEmpty(n)

	for i := 0; i < n; i++ {
		out[i][i] = number.ByteFieldElem(0x01)
	}

	return out
}

// GenerateEmpty creates a matrix with all entries set to 0.
func GenerateEmpty(n int) Matrix {
	out := make([]Row, n)

	for i := 0; i < n; i++ {
		out[i] = make([]number.ByteFieldElem, n)
	}

	return Matrix(out)
}
