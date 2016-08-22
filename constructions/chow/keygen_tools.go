// Contains tools for key generation that don't fit anywhere else.
package chow

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

// Generate the XOR Tables for squashing the result of a Tyi Table or MB^(-1) Table.
func xorTables(rs *common.RandomSource, surface common.Surface, shift func(int) int) (out [9][32][3]table.Nibble) {
	for round := 0; round < 9; round++ {
		for pos := 0; pos < 32; pos++ {
			out[round][pos][0] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					StepEncoding(rs, round, pos/8*4+0, pos%8, surface),
					StepEncoding(rs, round, pos/8*4+1, pos%8, surface),
				},
				XOREncoding(rs, round, surface)(pos, 0),
				common.XORTable{},
			}

			out[round][pos][1] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					XOREncoding(rs, round, surface)(pos, 0),
					StepEncoding(rs, round, pos/8*4+2, pos%8, surface),
				},
				XOREncoding(rs, round, surface)(pos, 1),
				common.XORTable{},
			}

			out[round][pos][2] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					XOREncoding(rs, round, surface)(pos, 1),
					StepEncoding(rs, round, pos/8*4+3, pos%8, surface),
				},
				RoundEncoding(rs, round, surface, shift)(pos),
				common.XORTable{},
			}
		}
	}

	return
}
