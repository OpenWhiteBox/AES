// Package xiao implements a cryptanalysis of the Xiao and Lai's white-box AES constructions.
//
// It is built on top of the ASA cryptanalysis from Generic/cryptanalysis/spn.
//
// Source paper to be added.
package xiao

import (
	"github.com/OpenWhiteBox/primitives/encoding"

	"github.com/OpenWhiteBox/AES/constructions/saes"
	"github.com/OpenWhiteBox/AES/constructions/xiao"

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

// shiftrows implements a Block encoding over the ShiftRows operation.
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

// round isolates one round of encryption with an AES white-box.
type round struct {
	construction *xiao.Construction
	round        int
}

func (r round) Encrypt(dst, src []byte) {
	copy(dst[0:16], src[0:16])

	for pos := 0; pos < 16; pos += 4 {
		stretched := r.construction.ExpandWord(r.construction.TBoxMixCol[r.round][pos/2:(pos+4)/2], dst[pos:pos+4])
		r.construction.SquashWords(stretched, dst[pos:pos+4])
	}
}

// RecoverKey returns the AES key used to generate the given white-box construction.
func RecoverKey(constr *xiao.Construction) []byte {
	round1 := round{
		construction: constr,
		round:        1,
	}

	// Decomposition Phase
	constr1 := aspn.DecomposeSPN(round1, cspn.ASA)

	var (
		first, last = AffineLayer(constr1[0].(encoding.BlockAffine)), AffineLayer(constr1[2].(encoding.BlockAffine))
		middle      = SBoxLayer(constr1[1].(encoding.ConcatenatedBlock))
	)

	// Disambiguation Phase
	// The SPN decomposition naturally leaves the last affine layer without a constant part. We would push it into the
	// middle S-boxes if that wasn't the case.

	// Put the affine layers in diagonal form.
	perm := first.FindPermutation()
	permEnc := encoding.NewBlockLinear(perm)

	first.RightCompose(encoding.InverseBlock{permEnc})
	middle.PermuteBy(perm, false)
	last.LeftCompose(permEnc)

	// Whiten the S-boxes so that they are linearly equivalent to Sbar.
	mask := middle.Whiten()
	encoding.XOR(first.BlockAdditive[:], first.BlockAdditive[:], mask[:])

	// Fix the S-boxes so that they are equal to Sbar.
	in, out := middle.CleanLinear()

	first.RightCompose(in)
	last.LeftCompose(out)

	// Add ShiftRows matrix to make search possible.
	last.RightCompose(encoding.NewBlockLinear(constr.ShiftRows[2]))

	// Clean off remaining noise from self-equivalences of Sbar.
	left := last.CleanLeft()
	right := encoding.ComposedBlocks{
		middle, left, encoding.InverseBlock{middle},
	}

	first.RightCompose(right)

	// Convert Sbar back to AES's "standard" S-box.
	for pos := 0; pos < 16; pos++ {
		first.BlockAdditive[pos] ^= 0x52
		middle[pos] = encoding.ComposedBytes{encoding.ByteAdditive(0x52), middle[pos]}
	}

	// fmt.Println(encoding.ProbablyEquivalentBlocks(
	//   encoding.ComposedBlocks{first, middle, last},
	//   encoding.ComposedBlocks{aspn.Encoding{round1}, encoding.NewBlockLinear(constr.ShiftRows[2])},
	// ))
	// fmt.Println(encoding.ProbablyEquivalentBlocks(
	//   aspn.Encoding{constr1},
	//   aspn.Encoding{round1},
	// ))
	//
	// Output:
	//   true
	//   true

	roundKey := shiftrows{}.Decode(first.BlockAdditive)
	return backOneRound(roundKey[:], 1)
}
