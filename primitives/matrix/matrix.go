// Basic operations on 8x8 matrices in GF(2) and the random generation of new ones.
package matrix

import (
	"io"
)

func dotProduct(e, f uint32, cap uint) bool {
	weight := 0
	x := e & f

	for i := uint(0); i < cap; i++ {
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

func dotProductByte(e, f byte) bool {
	return dotProduct(uint32(e), uint32(f), 8)
}

func dotProductWord(e, f uint32) bool {
	return dotProduct(e, f, 32)
}

type ByteMatrix [8]byte // One byte is one row

func (e ByteMatrix) Mul(f byte) (out byte) {
	for i := uint(0); i < 8; i++ {
		if dotProductByte(e[i], f) {
			out += 1 << i
		}
	}

	return
}

func (e ByteMatrix) Add(f ByteMatrix) (out ByteMatrix) {
	for i := 0; i < 8; i++ {
		out[i] = e[i] ^ f[i]
	}

	return
}

func (e ByteMatrix) Invert() (out ByteMatrix, ok bool) { // Gauss-Jordan Method
	out = ByteMatrix{1, 2, 4, 8, 16, 32, 64, 128} // Identity matrix

	f := [8]byte{}
	copy(f[:], e[:])

	for row := uint(0); row < 8; row++ {
		// Find a row with a non-zero entry (a 1) in the (row)th position
		candId := uint(255)

		for i := row; i < 8; i++ {
			if (f[i]>>row)&1 != 0 {
				candId = i
				break
			}
		}

		if candId == 255 { // If we can't find one, the matrix isn't invertible.
			return out, false
		}

		// Move it to the top
		f[row], f[candId] = f[candId], f[row]
		out[row], out[candId] = out[candId], out[row]

		// Cancel out the (row)th position for every row above and below it.
		for i := uint(0); i < 8; i++ {
			if i == row {
				continue
			}

			if (f[i]>>row)&1 != 0 {
				f[i] ^= f[row]
				out[i] ^= out[row]
			}
		}
	}

	return out, true
}

func GenerateRandomByte(reader io.Reader) ByteMatrix {
	m := ByteMatrix{} // Generate random byte matrix.
	reader.Read(m[:])

	_, ok := m.Invert() // Test for invertibility.

	if ok { // Return this one or try again.
		return m
	} else {
		return GenerateRandomByte(reader)
	}
}

type WordMatrix [32]uint32

func (e WordMatrix) Mul(f uint32) (out uint32) {
	for i := uint(0); i < 32; i++ {
		if dotProductWord(e[i], f) {
			out += 1 << i
		}
	}

	return
}

func (e WordMatrix) Add(f WordMatrix) (out WordMatrix) {
	for i := 0; i < 32; i++ {
		out[i] = e[i] ^ f[i]
	}

	return
}

func (e WordMatrix) Invert() (out WordMatrix, ok bool) { // Gauss-Jordan Method
	// Generate identity matrix:
	for i := uint(0); i < 32; i++ {
		out[i] = 1 << i
	}

	f := [32]uint32{}
	copy(f[:], e[:])

	for row := uint(0); row < 32; row++ {
		// Find a row with a non-zero entry (a 1) in the (row)th position
		candId := uint(255)

		for i := row; i < 32; i++ {
			if (f[i]>>row)&1 != 0 {
				candId = i
				break
			}
		}

		if candId == 255 { // If we can't find one, the matrix isn't invertible.
			return out, false
		}

		// Move it to the top
		f[row], f[candId] = f[candId], f[row]
		out[row], out[candId] = out[candId], out[row]

		// Cancel out the (row)th position for every row above and below it.
		for i := uint(0); i < 32; i++ {
			if i == row {
				continue
			}

			if (f[i]>>row)&1 != 0 {
				f[i] ^= f[row]
				out[i] ^= out[row]
			}
		}
	}

	return out, true
}

func GenerateRandomWord(reader io.Reader) WordMatrix {
	m := [32 * 4]byte{} // Generate random byte matrix.
	reader.Read(m[:])

	n := WordMatrix{}
	for i := 0; i < 32; i++ {
		n[i] = uint32(m[4*i+0])<<24 | uint32(m[4*i+1])<<16 | uint32(m[4*i+2])<<8 | uint32(m[4*i+3])
	}

	_, ok := n.Invert() // Test for invertibility.

	if ok { // Return this one or try again.
		return n
	} else {
		return GenerateRandomWord(reader)
	}
}
