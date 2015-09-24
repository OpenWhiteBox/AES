package chow

import (
	"../../primitives/encoding"
	"../../primitives/table"
	"../saes"
)

type TyiEncoding struct {
	Position, SubPosition int
}

func (tyi TyiEncoding) Encode(i byte) byte { return (i ^ byte(tyi.Position+tyi.SubPosition)) & 0xf }
func (tyi TyiEncoding) Decode(i byte) byte { return (i ^ byte(tyi.Position+tyi.SubPosition)) & 0xf }

func GenerateKeys(key [16]byte) (tyi [9][16]table.Word, tbox [16]table.Byte, xor [9][32][3]table.Nibble) {
	constr := saes.Construction{key}
	roundKeys := constr.StretchedKey()

	// Apply ShiftRows to round keys 0 to 9.
	for k := 0; k < 10; k++ {
		roundKeys[k] = constr.ShiftRows(roundKeys[k])
	}

	for round := 0; round < 9; round++ {
		for pos := 0; pos < 16; pos++ {
			// Build the T-Box and Tyi Table for this round and position in the state matrix.
			tyi[round][pos] = encoding.WordTable{
				encoding.IdentityByte{},
				encoding.ConcatenatedWord{
					encoding.ConcatenatedByte{TyiEncoding{pos, 0}, TyiEncoding{pos, 1}},
					encoding.ConcatenatedByte{TyiEncoding{pos, 2}, TyiEncoding{pos, 3}},
					encoding.ConcatenatedByte{TyiEncoding{pos, 4}, TyiEncoding{pos, 5}},
					encoding.ConcatenatedByte{TyiEncoding{pos, 6}, TyiEncoding{pos, 7}},
				},
				table.ComposedToWord{
					TBox{constr, roundKeys[round][pos], 0},
					TyiTable(pos % 4),
				},
			}
		}

		// Generate the two top-level XOR Tables
		for pos := 0; pos < 32; pos++ {
			xor[round][pos][0] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					TyiEncoding{pos/8*4 + 0, pos % 8},
					TyiEncoding{pos/8*4 + 1, pos % 8},
				},
				encoding.IdentityByte{},
				XORTable{},
			}

			xor[round][pos][1] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					TyiEncoding{pos/8*4 + 2, pos % 8},
					TyiEncoding{pos/8*4 + 3, pos % 8},
				},
				encoding.IdentityByte{},
				XORTable{},
			}

			xor[round][pos][2] = XORTable{}
		}
	}

	// 10th T-Box
	for pos := 0; pos < 16; pos++ {
		tbox[pos] = TBox{constr, roundKeys[9][pos], roundKeys[10][pos]}
	}

	return
}
