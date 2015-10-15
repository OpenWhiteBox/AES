package chow

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/table"
)

type TableAsEncoding struct {
	Forwards, Backwards table.InvertibleTable
}

func (tae TableAsEncoding) Encode(i byte) byte { return tae.Forwards.Get(i) }
func (tae TableAsEncoding) Decode(i byte) byte { return tae.Backwards.Get(i) }

// FunctionFromBasis produces the element of S according a basis and a specified combination of its elements.  It is an
// isomorphism from (0..2^8-1, xor) -> (S, compose).
func FunctionFromBasis(i int, basis []table.Byte) table.Byte {
	// Generate the function specified by i.
	vect := table.ComposedBytes{
		table.IdentityByte{},
	}

	for j := uint(0); j < uint(len(basis)); j++ {
		if (i>>j)&1 == 1 {
			vect = append(vect, basis[j])
		}
	}

	return vect
}

// DecomposeAffineEncoding is an efficient way to factor an unknown affine encoding into its component linear and
// affine parts.
func DecomposeAffineEncoding(e encoding.Byte) (matrix.Matrix, byte) {
	m := matrix.Matrix{
		matrix.Row{0}, matrix.Row{0}, matrix.Row{0}, matrix.Row{0},
		matrix.Row{0}, matrix.Row{0}, matrix.Row{0}, matrix.Row{0},
	}
	c := e.Encode(0)

	for i := uint(0); i < 8; i++ {
		x := e.Encode(1<<i) ^ c

		for j := uint(0); j < 8; j++ {
			if (x>>j)&1 == 1 {
				m[j][0] += 1 << i
			}
		}
	}

	return m, c
}

// FindCharacteristic finds the characteristic number of a matrix.  This number is invariant to matrix similarity.
func FindCharacteristic(A matrix.Matrix) (a byte) {
	AkEnc := encoding.ComposedBytes{}

	for k := uint(0); k < 8; k++ {
		AkEnc = append(AkEnc, encoding.ByteLinear(A))
		Ak, _ := DecomposeAffineEncoding(AkEnc)
		a ^= Ak.Trace() << k
	}

	return
}

// Index in, index out.  Example: shiftRows(5) = 1 because ShiftRows(block) returns [16]byte{block[0], block[5], ...
func shiftRows(i int) int {
	return []int{0, 13, 10, 7, 4, 1, 14, 11, 8, 5, 2, 15, 12, 9, 6, 3}[i]
}
