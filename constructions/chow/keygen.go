package chow

import (
	"../../primitives/encoding"
	"../../primitives/table"
	"../saes"
)

type TyiEncoding struct {
	Round, Position, SubPosition int
}

func (tyi TyiEncoding) Encode(i byte) byte { return i }
func (tyi TyiEncoding) Decode(i byte) byte { return i }

type XOREncoding struct {
	Round, Position, Side int
}

func (xor XOREncoding) Encode(i byte) byte { return i }
func (xor XOREncoding) Decode(i byte) byte { return i }

func GenerateKeys(key [16]byte) (out Construction) {
	constr := saes.Construction{key}
	roundKeys := constr.StretchedKey()

	// Apply ShiftRows to round keys 0 to 9.
	for k := 0; k < 10; k++ {
		roundKeys[k] = constr.ShiftRows(roundKeys[k])
	}

	for round := 0; round < 9; round++ {
		for pos := 0; pos < 16; pos++ {
			// Build the T-Box and Tyi Table for this round and position in the state matrix.
			out.TBoxTyiTable[round][pos] = encoding.WordTable{
				encoding.IdentityByte{},
				encoding.ConcatenatedWord{
					encoding.ConcatenatedByte{TyiEncoding{round, pos, 0}, TyiEncoding{round, pos, 1}},
					encoding.ConcatenatedByte{TyiEncoding{round, pos, 2}, TyiEncoding{round, pos, 3}},
					encoding.ConcatenatedByte{TyiEncoding{round, pos, 4}, TyiEncoding{round, pos, 5}},
					encoding.ConcatenatedByte{TyiEncoding{round, pos, 6}, TyiEncoding{round, pos, 7}},
				},
				table.ComposedToWord{
					TBox{constr, roundKeys[round][pos], 0},
					TyiTable(pos % 4),
				},
			}
		}

		// Generate the XOR Tables
		for pos := 0; pos < 32; pos++ {
			out.XORTable[round][pos][0] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					TyiEncoding{round, pos/8*4 + 0, pos % 8},
					TyiEncoding{round, pos/8*4 + 1, pos % 8},
				},
				XOREncoding{round, pos, 0},
				XORTable{},
			}

			out.XORTable[round][pos][1] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					TyiEncoding{round, pos/8*4 + 2, pos % 8},
					TyiEncoding{round, pos/8*4 + 3, pos % 8},
				},
				XOREncoding{round, pos, 1},
				XORTable{},
			}

			out.XORTable[round][pos][2] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					XOREncoding{round, pos, 0},
					XOREncoding{round, pos, 1},
				},
				encoding.IdentityByte{},
				XORTable{},
			}
		}
	}

	// 10th T-Box
	for pos := 0; pos < 16; pos++ {
		out.TBox[pos] = TBox{constr, roundKeys[9][pos], roundKeys[10][pos]}
	}

	return
}
