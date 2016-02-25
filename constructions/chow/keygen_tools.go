package chow

import (
	"github.com/OpenWhiteBox/primitives/encoding"
	"github.com/OpenWhiteBox/primitives/random"
	"github.com/OpenWhiteBox/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

// xorTables generates the XOR Tables for squashing the result of a Tyi Table or MB^(-1) Table.
func xorTables(rs *random.Source, surface common.Surface, shift func(int) int) (out [9][32][3]table.Nibble) {
	for round := 0; round < 9; round++ {
		for pos := 0; pos < 32; pos++ {
			out[round][pos][0] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					stepEncoding(rs, round, pos/8*4+0, pos%8, surface),
					stepEncoding(rs, round, pos/8*4+1, pos%8, surface),
				},
				xorEncoding(rs, round, surface)(pos, 0),
				common.NibbleXORTable{},
			}

			out[round][pos][1] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					xorEncoding(rs, round, surface)(pos, 0),
					stepEncoding(rs, round, pos/8*4+2, pos%8, surface),
				},
				xorEncoding(rs, round, surface)(pos, 1),
				common.NibbleXORTable{},
			}

			out[round][pos][2] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					xorEncoding(rs, round, surface)(pos, 1),
					stepEncoding(rs, round, pos/8*4+3, pos%8, surface),
				},
				roundEncoding(rs, round, surface, shift)(pos),
				common.NibbleXORTable{},
			}
		}
	}

	return
}
