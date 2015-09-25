// Basic operations on 8x8 matrices in GF(2) and the random generation of new ones.
package matrix

import (
	"crypto/rand"
)

type ByteMatrix [8]byte // One byte is one row

func dotProduct(e, f byte) bool {
	weight := 0
	x := e & f

	for i := uint(0); i < 8; i++ {
		if x&(1<<i) > 0 {
			weight++
		}
	}

	if weight%2 == 0 {
		return false
	} else {
		return true
	}
}

func (e ByteMatrix) Mul(f byte) (out byte) {
	for i := uint(0); i < 8; i++ {
		if dotProduct(e[i], f) {
			out += 1 << i
		}
	}

	return
}

func (e ByteMatrix) Invert() (out ByteMatrix, ok bool) { // Gauss-Jordan Method
	out = ByteMatrix{1, 2, 4, 8, 16, 32, 64, 128} // Identity matrix

	f := [8]byte{}
	copy(f[:], e[:])

	for row := 0; row < 8; row++ {
		// Find a row with a non-zero entry (a 1) in the (row)th position
		candId := -1

		for i := row; i < 8; i++ {
			if f[i]&(1<<uint(row)) != 0 {
				candId = i
				break
			}
		}

		if candId == -1 { // If we can't find one, the matrix isn't invertible.
			return out, false
		}

		// Move it to the top
		f[row], f[candId] = f[candId], f[row]
		out[row], out[candId] = out[candId], out[row]

		// Cancel out the (row)th position for every row above and below it.
		for i := 0; i < 8; i++ {
			if i == row {
				continue
			}

			if f[i]&(1<<uint(row)) != 0 {
				f[i] ^= f[row]
				out[i] ^= out[row]
			}
		}
	}

	return out, true
}

func GenerateRandom() ByteMatrix {
	m := [8]byte{} // Generate random byte matrix.
	rand.Read(m[:])

	_, ok := ByteMatrix(m).Invert() // Test for invertibility.

	if ok { // Return this one or try again.
		return ByteMatrix(m)
	} else {
		return GenerateRandom()
	}
}
