// Package toy implements the toy white-box AES construction, based on simple SPN disambiguation.
//
// http://dl.acm.org/citation.cfm?id=2995314
package toy

import (
	"github.com/OpenWhiteBox/primitives/encoding"
	"github.com/OpenWhiteBox/primitives/number"
)

type Construction [11]encoding.BlockAffine

// BlockSize returns the block size of AES. (Necessary to implement cipher.Block.)
func (constr Construction) BlockSize() int { return 16 }

// Encrypt encrypts the first block in src into dst. Dst and src may point at the same memory.
func (constr Construction) Encrypt(dst, src []byte) {
	state := [16]byte{}
	copy(state[:], src[:])

	state = constr[0].Encode(state)

	for round := 1; round < 11; round++ {
		for pos := 0; pos < 16; pos++ {
			state[pos] = byte(number.ByteFieldElem(state[pos]).Invert())
		}

		state = constr[round].Encode(state)
	}

	copy(dst[:], state[:])
}

// Decrypt decrypts the first block in src into dst. Dst and src may point at the same memory.
func (constr Construction) Decrypt(dst, src []byte) {
	state := [16]byte{}
	copy(state[:], src[:])

	state = constr[10].Decode(state)

	for round := 9; round >= 0; round-- {
		for pos := 0; pos < 16; pos++ {
			state[pos] = byte(number.ByteFieldElem(state[pos]).Invert())
		}

		state = constr[round].Decode(state)
	}

	copy(dst[:], state[:])
}
