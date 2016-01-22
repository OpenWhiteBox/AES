package matrix

import (
	"crypto/rand"
	"io"
)

// GenerateIdentity generates the n-by-n identity matrix.
func GenerateIdentity(n int) Matrix {
	return GeneratePartialIdentity(n, IgnoreNoRows)
}

// GeneratePartialIdentity generates the n-by-n identity matrix on some rows and leaves others zero (the rows where
// ignore(row) == true).
func GeneratePartialIdentity(n int, ignore RowIgnore) Matrix {
	out := GenerateEmpty(n, n)

	for i := 0; i < n; i++ {
		if !ignore(i) {
			out[i].SetBit(i, true)
		}
	}

	return out
}

// GenerateFull generates the n-by-n matrix with all entries set to 1.
func GenerateFull(n, m int) Matrix {
	out := GenerateEmpty(n, m)

	for i, _ := range out {
		for j, _ := range out[i] {
			out[i][j] = 0xff
		}
	}

	return out
}

// GenerateEmpty generates the n-by-n matrix with all entries set to 0.
func GenerateEmpty(n, m int) Matrix {
	out := make([]Row, n)

	for i := 0; i < n; i++ {
		out[i] = NewRow(m)
	}

	return Matrix(out)
}

// GenerateRandomRow generates a random n-component row.
func GenerateRandomRow(reader io.Reader, n int) Row {
	out := Row(make([]byte, rowsToColumns(n)))
	reader.Read(out)

	return out
}

// GenerateRandomNonZeroRow generates a random non-zero n-component row.
func GenerateRandomNonZeroRow(reader io.Reader, n int) Row {
	out := NewRow(n)

	for out.IsZero() {
		out = GenerateRandomRow(reader, n)
	}

	return out
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
		for col := 0; col < rowsToColumns(n); col++ {
			if !ignore(rowsToColumns(row), col) {
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

// GenerateTrueRandom generates a random n-by-n matrix (not guaranteed to be invertible) using the random source random
// (for example, crypto/rand.Reader).
func GenerateTrueRandom(reader io.Reader, n int) Matrix {
	m := make([]Row, n)

	for i, _ := range m { // Generate random n x n matrix.
		m[i] = GenerateRandomRow(rand.Reader, n)
	}

	return m
}

// GeneratePermutationMatrix generates an 8n-by-8n permutation matrix corresponding to a permutation of {0, ..., n-1}.
func GeneratePermutationMatrix(permutation []int) Matrix {
	n := len(permutation)
	out := GenerateEmpty(8*n, 8*n)

	for i, j := range permutation {
		for k := 0; k < 8; k++ {
			out[8*i+k].SetBit(8*j+k, true)
		}
	}

	return out
}
