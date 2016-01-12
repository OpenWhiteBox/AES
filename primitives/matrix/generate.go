package matrix

import (
	"io"
)

// GenerateIdentity generates the n-by-n identity matrix.
func GenerateIdentity(n int) Matrix {
	return GeneratePartialIdentity(n, IgnoreNoRows)
}

// GeneratePartialIdentity generates the n-by-n identity matrix on some rows and leaves others zero (the rows where
// ignore(row) == true).
func GeneratePartialIdentity(n int, ignore RowIgnore) Matrix {
	out := GenerateEmpty(n)

	for i := 0; i < n; i++ {
		if !ignore(i) {
			out[i].SetBit(i, true)
		}
	}

	return out
}

// GenerateFull generates the n-by-n matrix with all entries set to 1.
func GenerateFull(n int) Matrix {
	out := GenerateEmpty(n)
	for i := 0; i < n; i++ {
		for j := 0; j < n/8; j++ {
			out[i][j] = 0xff
		}
	}

	return out
}

// GenerateEmpty generates the n-by-n matrix with all entries set to 0.
func GenerateEmpty(n int) Matrix {
	out := make([]Row, n)

	for i := 0; i < n; i++ {
		out[i] = make([]byte, rowsToColumns(n))
	}

	return Matrix(out)
}

// GenerateRandom generates a random invertible n-by-n matrix using the random source random (for example,
// crypto/rand.Reader).
func GenerateRandom(reader io.Reader, n int) Matrix {
	m := GenerateTrueRandom(reader, n)

	_, ok := m.Invert()
	if !ok { // Try again
		return GenerateRandom(reader, n) // Performance bottleneck.
	}

	return m // This one works!
}

// GenerateRandomPartial generates an invertible n-by-n matrix which is random in some locations and the identity / zero
// in others, using the random source random (for example, crypto/rand.Reader). idIgnore is passes to
// GeneratePartialIdentity--it sets which rows of the identity are zero. The generated matrix is filled with random data
// everywhere that ignore(row, col) == false.
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

// GenerateTrueRandom generates a random  n-by-n matrix (not guaranteed to be invertible) using the random source random
// (for example, crypto/rand.Reader).
func GenerateTrueRandom(reader io.Reader, n int) Matrix {
	m := Matrix(make([]Row, n))

	for i := 0; i < n; i++ { // Generate random n x n matrix.
		row := Row(make([]byte, rowsToColumns(n)))
		reader.Read(row)

		m[i] = row
	}

	return m
}
