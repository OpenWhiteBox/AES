package toy

import (
	"github.com/OpenWhiteBox/AES/constructions/saes"
)

// shiftrows implements encoding.Block over the ShiftRows operation.
type shiftrows struct{}

func (sr shiftrows) Encode(in [16]byte) (out [16]byte) {
	constr := saes.Construction{}

	copy(out[:], in[:])
	constr.ShiftRows(out[:])

	return
}

func (sr shiftrows) Decode(in [16]byte) (out [16]byte) {
	constr := saes.Construction{}

	copy(out[:], in[:])
	constr.UnShiftRows(out[:])

	return
}

// permElem represents an element of the permutation group of the affine layers.
type permElem struct {
	rots [4]int // Each value is in [0, 4), representing how much the i^th block is permuted.
	perm [4]int // This encodes a permutation of {0, ..., 3}.
}

func (p *permElem) Encode(in [16]byte) (out [16]byte) {
	for a := 0; a < 4; a++ {
		for b := 0; b < 4; b++ {
			c, d := p.perm[a], (b+p.rots[a])%4
			i, j := a*4+b, c*4+d

			out[j] = in[i]
		}
	}

	return out
}

func (p *permElem) Decode(in [16]byte) (out [16]byte) {
	for a := 0; a < 4; a++ {
		for b := 0; b < 4; b++ {
			var c, d int

			// Determine c.
			for c = 0; c < 4; c++ {
				if p.perm[c] == a {
					break
				}
			}

			// Determine d.
			d = (b + 4 - p.rots[c]) % 4

			i, j := a*4+b, c*4+d
			out[j] = in[i]
		}
	}

	return out
}
