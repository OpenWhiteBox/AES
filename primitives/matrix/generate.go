// Functions for generating new matrices.
package matrix

import (
	"io"
)

// GenerateIdentity creates the n by n identity matrix.
func GenerateIdentity(n int) Matrix {
	return GeneratePartialIdentity(n, IgnoreNoRows)
}

// GeneratePartialIdentity creates the n by n identity matrix on some rows and leaves others zero.
func GeneratePartialIdentity(n int, ignore RowIgnore) Matrix {
	out := GenerateEmpty(n)

	for i := 0; i < n; i++ {
		if !ignore(i) {
			out[i].SetBit(i, true)
		}
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
	if !ok { // Try again
		return GenerateRandom(reader, n) // Performance bottleneck.
	}

	return m // This one works!
}

// GenerateRandomPartial creates an invertible matrix which is random in some locations and the identity in others,
// depending on an ignore function.
func GenerateRandomPartial(reader io.Reader, n int, ignore ByteIgnore, idIgnore RowIgnore) (Matrix, Matrix) {
	m := GeneratePartialIdentity(n, idIgnore)

	for row := 0; row < n; row++ {
		for col := 0; col < n/8; col++ {
			if !ignore(row/8, col) {
				reader.Read(m[row][col : col+1])
			}
		}
	}

	mInv, ok := m.Invert()
	if !ok {
		return GenerateRandomPartial(reader, n, ignore, idIgnore)
	}

	return m, mInv
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
