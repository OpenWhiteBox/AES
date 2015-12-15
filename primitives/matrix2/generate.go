package matrix2

import (
	"io"

	"github.com/OpenWhiteBox/AES/primitives/number"
)

// GenerateTrueRandom creates a random singular or non-singular n by n matrix.
func GenerateTrueRandom(reader io.Reader, n int) Matrix {
	m := Matrix(make([]Row, n))

	for i := 0; i < n; i++ { // Generate random n x n matrix.
		temp := make([]byte, n)
		reader.Read(temp)

		m[i] = Row(make([]number.ByteFieldElem, n))
		for j := 0; j < n; j++ {
			m[i][j] = number.ByteFieldElem(temp[j])
		}
	}

	return m
}

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
