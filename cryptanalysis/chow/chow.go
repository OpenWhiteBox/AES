// Package chow implements a cryptanalysis of Chow et al.'s white-box AES construction.
//
// It is built on top of the SAS cryptanalysis in Generic/cryptanalysis/spn.
//
// http://dl.acm.org/citation.cfm?id=2995314
package chow

import (
	"github.com/OpenWhiteBox/primitives/encoding"

	"github.com/OpenWhiteBox/AES/constructions/chow"
	"github.com/OpenWhiteBox/AES/constructions/common"
	"github.com/OpenWhiteBox/AES/constructions/saes"

	cspn "github.com/OpenWhiteBox/Generic/constructions/spn"
	aspn "github.com/OpenWhiteBox/Generic/cryptanalysis/spn"
)

var powx = [16]byte{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x1b, 0x36, 0x6c, 0xd8, 0xab, 0x4d, 0x9a, 0x2f}

// backOneRound takes round key i and returns round key i-1.
func backOneRound(roundKey []byte, round int) (out []byte) {
	out = make([]byte, 16)
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

// isAS returns true if the given Byte encoding might be an AS structure, with 2 4-bit S-boxes.
func isAS(in encoding.Byte) bool {
	temp1, temp2 := byte(0x00), byte(0x00)

	for x := byte(0); x < 16; x++ {
		temp1 ^= in.Encode(x)
		temp2 ^= in.Encode(x << 4)
	}

	return temp1 == 0 && temp2 == 0
}

// round isolates one round of encryption with an AES white-box.
type round struct {
	construction *chow.Construction
	round        int
}

func (r round) Encrypt(dst, src []byte) {
	copy(dst[0:16], src[0:16])

	for pos := 0; pos < 16; pos += 4 {
		stretched := r.construction.ExpandWord(r.construction.TBoxTyiTable[r.round][pos:pos+4], dst[pos:pos+4])
		r.construction.SquashWords(r.construction.HighXORTable[r.round][2*pos:2*pos+8], stretched, dst[pos:pos+4])

		stretched = r.construction.ExpandWord(r.construction.MBInverseTable[r.round][pos:pos+4], dst[pos:pos+4])
		r.construction.SquashWords(r.construction.LowXORTable[r.round][2*pos:2*pos+8], stretched, dst[pos:pos+4])
	}
}

// RecoverKey returns the AES key used to generate the given white-box construction.
func RecoverKey(constr *chow.Construction) []byte {
	round1, round2 := round{
		construction: constr,
		round:        1,
	}, round{
		construction: constr,
		round:        2,
	}

	// Decomposition Phase
	constr1 := aspn.DecomposeSPN(round1, cspn.SAS)
	constr2 := aspn.DecomposeSPN(round2, cspn.SAS)

	var (
		leading, middle, trailing sboxLayer
		left, right               = affineLayer(constr1[1].(encoding.BlockAffine)), affineLayer(constr2[1].(encoding.BlockAffine))
	)

	for pos := 0; pos < 16; pos++ {
		leading[pos] = constr1[0].(encoding.ConcatenatedBlock)[pos]
		middle[pos] = encoding.ComposedBytes{
			constr1[2].(encoding.ConcatenatedBlock)[pos],
			constr2[0].(encoding.ConcatenatedBlock)[common.ShiftRows(pos)],
		}
		trailing[pos] = constr2[2].(encoding.ConcatenatedBlock)[pos]
	}

	// Disambiguation Phase
	// Disambiguate the affine layer.
	lin, lout := left.clean()
	rin, rout := right.clean()

	leading.rightCompose(lin, common.NoShift)
	middle.leftCompose(lout, common.NoShift).rightCompose(rin, common.ShiftRows)
	trailing.leftCompose(rout, common.NoShift)

	// The SPN decomposition naturally leaves the affine layers without a constant part.
	// We would push it into the S-boxes here if that wasn't the case.

	// Move the constant off of the input and output of the S-boxes.
	mcin, mcout := middle.cleanConstant()
	mcin, mcout = left.Decode(mcin), right.Encode(mcout)

	leading.rightCompose(encoding.DecomposeConcatenatedBlock(encoding.BlockAdditive(mcin)), common.NoShift)
	trailing.leftCompose(encoding.DecomposeConcatenatedBlock(encoding.BlockAdditive(mcout)), common.NoShift)

	// Move the multiplication off of the input and output of the middle S-boxes.
	mlin, mlout := middle.cleanLinear()

	leading.rightCompose(mlin, common.NoShift)
	trailing.leftCompose(mlout, common.NoShift)

	// fmt.Println(encoding.ProbablyEquivalentBlocks(
	// 	encoding.ComposedBlocks{aspn.Encoding{round1}, ShiftRows{}, aspn.Encoding{round2}},
	// 	encoding.ComposedBlocks{leading, left, middle, ShiftRows{}, right, trailing},
	// ))
	// Output: true

	// Extract the key from the leading S-boxes.
	key := [16]byte{}

	for pos := 0; pos < 16; pos++ {
		for guess := 0; guess < 256; guess++ {
			cand := encoding.ComposedBytes{
				leading[pos], encoding.ByteAdditive(guess), encoding.InverseByte{sbox{}},
			}

			if isAS(cand) {
				key[pos] = byte(guess)
				break
			}
		}
	}

	key = left.Encode(key)

	return backOneRound(backOneRound(key[:], 2), 1)
}
