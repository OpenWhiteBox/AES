// Contains tools for key generation that don't fit anywhere else.
package chow

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

// Generate the XOR Tables for squashing the result of a Input/Output Mask.
func blockXORTables(seed []byte, surface common.Surface, shift func(int) int) (out [32][15]table.Nibble) {
	for pos := 0; pos < 32; pos++ {
		out[pos][0] = encoding.NibbleTable{
			encoding.ConcatenatedByte{MaskEncoding(seed, 0, pos, surface), MaskEncoding(seed, 1, pos, surface)},
			XOREncoding(seed, 10, pos, 0, surface),
			XORTable{},
		}

		for i := 1; i < 14; i++ {
			out[pos][i] = encoding.NibbleTable{
				encoding.ConcatenatedByte{XOREncoding(seed, 10, pos, i-1, surface), MaskEncoding(seed, i+1, pos, surface)},
				XOREncoding(seed, 10, pos, i, surface),
				XORTable{},
			}
		}

		var outEnc encoding.Nibble
		if surface == common.Inside {
			outEnc = RoundEncoding(seed, -1, 2*shift(pos/2)+pos%2, common.Outside)
		} else {
			outEnc = encoding.IdentityByte{}
		}

		out[pos][14] = encoding.NibbleTable{
			encoding.ConcatenatedByte{XOREncoding(seed, 10, pos, 13, surface), MaskEncoding(seed, 15, pos, surface)},
			outEnc,
			XORTable{},
		}
	}

	return
}

// Generate the XOR Tables for squashing the result of a Tyi Table or MB^(-1) Table.
func xorTables(seed []byte, surface common.Surface, shift func(int) int) (out [9][32][3]table.Nibble) {
	var outPos func(int) int
	if surface == common.Inside {
		outPos = func(pos int) int { return pos }
	} else {
		outPos = func(pos int) int { return 2*shift(pos/2) + pos%2 }
	}

	for round := 0; round < 9; round++ {
		for pos := 0; pos < 32; pos++ {
			out[round][pos][0] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					StepEncoding(seed, round, pos/8*4+0, pos%8, surface),
					StepEncoding(seed, round, pos/8*4+1, pos%8, surface),
				},
				XOREncoding(seed, round, pos, 0, surface),
				XORTable{},
			}

			out[round][pos][1] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					XOREncoding(seed, round, pos, 0, surface),
					StepEncoding(seed, round, pos/8*4+2, pos%8, surface),
				},
				XOREncoding(seed, round, pos, 1, surface),
				XORTable{},
			}

			out[round][pos][2] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					XOREncoding(seed, round, pos, 1, surface),
					StepEncoding(seed, round, pos/8*4+3, pos%8, surface),
				},
				RoundEncoding(seed, round, outPos(pos), surface),
				XORTable{},
			}
		}
	}

	return
}
