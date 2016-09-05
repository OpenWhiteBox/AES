package chow

import (
	"github.com/OpenWhiteBox/primitives/encoding"
	"github.com/OpenWhiteBox/primitives/matrix"
	"github.com/OpenWhiteBox/primitives/number"

	"github.com/OpenWhiteBox/AES/constructions/common"
	"github.com/OpenWhiteBox/AES/constructions/saes"
)

// sbox is a Byte encoding of AES's "standard" S-box.
type sbox struct{}

func (sb sbox) Encode(in byte) byte {
	constr := saes.Construction{}
	return constr.SubByte(in)
}

func (s sbox) Decode(in byte) byte {
	constr := saes.Construction{}
	return constr.UnSubByte(in)
}

// invert is a Byte encoding of inversion over GF(2^8)
type invert struct{}

func (inv invert) Encode(in byte) byte {
	return byte(number.ByteFieldElem(in).Invert())
}

func (inv invert) Decode(in byte) byte {
	return byte(number.ByteFieldElem(in).Invert())
}

// sboxLayer implements methods for disambiguating an S-box layer of the SPN.
type sboxLayer encoding.ConcatenatedBlock

func (sbl sboxLayer) Encode(in [16]byte) [16]byte {
	return encoding.ConcatenatedBlock(sbl).Encode(in)
}

func (sbl sboxLayer) Decode(in [16]byte) [16]byte {
	return encoding.ConcatenatedBlock(sbl).Decode(in)
}

// reftCompose component-wise composes this SBox layer with a concatenated block encoding on the left (input). A shift
// to apply can be specified.
func (sbl *sboxLayer) leftCompose(left encoding.ConcatenatedBlock, shift func(int) int) *sboxLayer {
	for pos := 0; pos < 16; pos++ {
		(*sbl)[pos] = encoding.ComposedBytes{left[shift(pos)], sbl[pos]}
	}

	return sbl
}

// rightCompose component-wise composes this SBox layer with a concatenated block encoding on the right (output). A
// shift to apply can be specified.
func (sbl *sboxLayer) rightCompose(right encoding.ConcatenatedBlock, shift func(int) int) *sboxLayer {
	for pos := 0; pos < 16; pos++ {
		(*sbl)[pos] = encoding.ComposedBytes{sbl[pos], right[shift(pos)]}
	}

	return sbl
}

// cleanConstant finds the constant error on the input and output of each middle S-box. It removes it from the S-box and
// returns it.
//
// Note: This function will also strip the final addition of 0x63 from AES's "standard" S-box.
func (sbl *sboxLayer) cleanConstant() (input, output [16]byte) {
	for pos := 0; pos < 16; pos++ {
		in, out := sbl.findConstant(pos)

		input[pos], output[common.ShiftRows(pos)] = in, out
		(*sbl)[pos] = encoding.ComposedBytes{
			encoding.ByteAdditive(in), sbl[pos], encoding.ByteAdditive(out),
		}
	}

	return
}

// cleanLinear finds the linear error on the input and output of each middle S-box (after the constant error has been
// removed). It removes it from the S-box (leaving AES's "standard" S-box, without the 0x63 constant addition) and
// returns it.
func (sbl *sboxLayer) cleanLinear() (input, output encoding.ConcatenatedBlock) {
	for pos := 0; pos < 16; pos += 4 {
		in, out := sbl.findLinear(pos)

		for i := pos; i < pos+4; i++ {
			input[i], output[i] = encoding.InverseByte{in}, out
		}
	}

	for pos := 0; pos < 16; pos++ {
		(*sbl)[pos] = encoding.ComposedBytes{
			encoding.InverseByte{input[pos]}, sbl[pos], encoding.InverseByte{output[common.ShiftRows(pos)]},
		}
	}

	return
}

// findConstant returns the constant error on the input and output of the middle S-box at position pos.
func (sbl *sboxLayer) findConstant(pos int) (byte, byte) {
	for b := 0; b < 256; b++ { // Try to guess the constant on the input.
		correction := encoding.ByteAdditive(b)
		vec := encoding.ComposedBytes{invert{}, correction, sbl[pos]}

		// If the guess was correct, removing the constant and a field inversion from the input will result in an affine
		// function (the linear error, multiplication by an element of GF(2^8), moves through inversion).

		cand, ok := encoding.DecomposeByteAffine(vec)

		if ok && encoding.EquivalentBytes(vec, cand) {
			return byte(b), byte(cand.ByteAdditive)
		}
	}

	panic("Failed to find constant!")
}

// findLinear returns the linear error on the input and output of the mmiddle S-box at position pos (once the constant
// error has been removed). The function is a simple brute force attack.
func (sbl *sboxLayer) findLinear(pos int) (encoding.ByteMultiplication, encoding.ByteMultiplication) {
	subBytes := encoding.NewByteLinear(matrix.Matrix{
		matrix.Row{0xF1},
		matrix.Row{0xE3},
		matrix.Row{0xC7},
		matrix.Row{0x8F},
		matrix.Row{0x1F},
		matrix.Row{0x3E},
		matrix.Row{0x7C},
		matrix.Row{0xF8},
	})

	real := encoding.ComposedBytes{invert{}, sbl[pos]}

	for a := 1; a < 256; a++ {
		for c := 1; c < 256; c++ {
			in := encoding.NewByteMultiplication(number.ByteFieldElem(a))
			out := encoding.NewByteMultiplication(number.ByteFieldElem(c))

			cand := encoding.ComposedBytes{in, subBytes, out}

			if encoding.EquivalentBytes(cand, real) {
				return in, out
			}
		}
	}

	panic("Failed to find linear!")
}
