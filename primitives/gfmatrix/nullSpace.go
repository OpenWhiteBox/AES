package gfmatrix

import (
	"github.com/OpenWhiteBox/AES/primitives/number"
)

// NullSpace returns a basis for the matrix's nullspace.
func (e Matrix) NullSpace() (basis []Row) {
	if len(e) == 0 {
		return []Row{
			[]number.ByteFieldElem{},
		}
	}

	_, in := e.Size()
	_, f, frees := e.gaussJordan()

	for _, free := range frees {
		input := Row(make([]number.ByteFieldElem, in))
		input[free] = 0x01

		for _, row := range f {
			if !row[free].IsZero() {
				input[row.Height()] = row[free]
			}
		}

		basis = append(basis, input)
	}

	return
}
