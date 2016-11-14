// Package chow implements a cryptanalysis of the toy white-box AES construction.
//
// http://dl.acm.org/citation.cfm?id=2995314
package toy

import (
	"github.com/OpenWhiteBox/primitives/encoding"

	"github.com/OpenWhiteBox/AES/constructions/saes"
	"github.com/OpenWhiteBox/AES/constructions/toy"
)

var powx = [16]byte{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x1b, 0x36, 0x6c, 0xd8, 0xab, 0x4d, 0x9a, 0x2f}

// backOneRound takes round key i and returns round key i-1.
func backOneRound(roundKey [16]byte, round int) (out [16]byte) {
	constr := saes.Construction{}

	// Recover everything except the first word by XORing consecutive blocks.
	for pos := 4; pos < 16; pos++ {
		out[pos] = roundKey[pos] ^ roundKey[pos-4]
	}

	// Recover the first word by XORing the first block of the roundKey with f(last block of roundKey), where f is a
	// subroutine of AES' key scheduling algorithm.
	for pos := 0; pos < 4; pos++ {
		out[pos] = roundKey[pos] ^ constr.SubByte(out[12+(pos+1)%4])
	}
	out[0] ^= powx[round-1]

	return
}

// RecoverKey returns the AES key used to generate the given white-box construction.
func RecoverKey(constr *toy.Construction) []byte {
	var (
		target = affineLayer(constr[1]) // The layer we intend to fully disambiguate.
		aux1   = affineLayer(constr[2]) // Lets us learn the parasites of target, and the key material's permutation.
		aux2   = affineLayer(constr[3]) // Lets us learn the parasites of aux1.

		targetIn = target.parasites()
		aux1In   = aux1.parasites()
		aux2In   = aux2.parasites()
	)

	// Remove the parasites from target and aux1. Now, the block of each matrix are only permuted.
	target.cleanParasites(targetIn, aux1In)
	aux1.cleanParasites(aux1In, aux2In)

	// Compress target to a matrix over F_{2^8} and remove as much of the permutation as possible. (This will also get
	// rid of ShiftRows, but will not affect the round key because ShiftRows is on the input.)
	sm := target.compress()
	permIn, permOut := sm.unpermute()

	target.adjust(encoding.InverseBlock{permIn}, permOut)
	aux1.adjust(encoding.InverseBlock{permOut}, encoding.IdentityBlock{})

	// Remove the SubBytes constant from each permuted round key, so we can compare them.
	key1, key2 := [16]byte{}, [16]byte{}
	for pos := 0; pos < 16; pos++ {
		key1[pos] = target.BlockAdditive[pos] ^ 0x63
		key2[pos] = aux1.BlockAdditive[pos] ^ 0x63
	}

	// The linear part of target is correct, but its constant part is not. Take a guess for how the constant part is
	// permuted, and check if this agrees with aux1.
	for per := 0; per < 256; per++ {
		perm := [4]int{(per >> 0) & 3, (per >> 2) & 3, (per >> 4) & 3, (per >> 6) & 3}

		// Packing all the permutation information into one byte means that we get degenerate cases where the
		// requested "permutation" is not a bijection. Skip those cases.
		skip := false
		for i := 0; i < 3; i++ {
			for j := i + 1; j < 4; j++ {
				if perm[i] == perm[j] {
					skip = true
				}
			}
		}
		if skip {
			continue
		}

		for rot := 0; rot < 256; rot++ {
			rots := [4]int{(rot >> 0) & 3, (rot >> 2) & 3, (rot >> 4) & 3, (rot >> 6) & 3}
			guess := &permElem{rots: rots, perm: perm}

			cand1 := guess.Encode(key1)
			cand2 := encoding.ComposedBlocks{
				encoding.InverseBlock{aux1.BlockLinear}, guess, round,
			}.Encode(key2)

			if cand1 == backOneRound(cand2, 2) {
				sol := backOneRound(cand1, 1)
				return sol[:]
			}
		}
	}

	return nil
}
