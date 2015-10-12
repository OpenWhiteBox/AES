package chow

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/table"
)

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
