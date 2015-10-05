package chow

import (
	"github.com/OpenWhiteBox/AES/primitives/table"
)

type IdentityByte struct{}

func (ib IdentityByte) Get(i byte) byte { return i }

func FunctionFromBasis(i int, basis []table.Byte) table.Byte {
	// Generate the function specified by i.
	vect := table.ComposedBytes{
		IdentityByte{},
	}

	for j := uint(0); j < uint(len(basis)); j++ {
		if (i>>j)&1 == 1 {
			vect = append(vect, basis[j])
		}
	}

	return vect
}
