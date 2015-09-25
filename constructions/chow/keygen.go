package chow

import (
	"../../primitives/encoding"
	"../../primitives/table"
	"../saes"
)

type Side int

const (
	Left = iota
	Right
)

type TyiEncoding struct { // Encodes the output of a T-Box/Tyi Table / the input of a top-level XOR.
	Round       int
	Position    int // Position in the state array, counted in *bytes*.
	SubPosition int // Position in the T-Box/Tyi Table's ouptput for this byte, counted in nibbles.
}

func (tyi TyiEncoding) Encode(i byte) byte { return i }
func (tyi TyiEncoding) Decode(i byte) byte { return i }

type XOREncoding struct { // Encodes intermediate results between the two top-level XORs and the bottom XOR.
	Round    int
	Position int // Position in the state array, counted in nibbles.
	Side         // "Side" of the circuit. Left for the (a ^ b) and Right for the (c ^ d) side.
} // The bottom XOR decodes its input with a Left and Right XOREncoding and encodes its output with a RoundEncoding.

func (xor XOREncoding) Encode(i byte) byte { return i }
func (xor XOREncoding) Decode(i byte) byte { return i }

type RoundEncoding struct { // Encodes the output of each round / the input of the next round's T-Box/Tyi Table.
	Round    int
	Position int // Position in the state array, counted in nibbles.
}

func (round RoundEncoding) Encode(i byte) byte { return i }
func (round RoundEncoding) Decode(i byte) byte { return i }

// Index in, index out.  Example: shiftRows(5) = 1 because ShiftRows(block) returns [16]byte{block[0], block[5], ...
func shiftRows(i int) int {
	return []int{0, 13, 10, 7, 4, 1, 14, 11, 8, 5, 2, 15, 12, 9, 6, 3}[i]
}

func GenerateKeys(key [16]byte) (out Construction) {
	constr := saes.Construction{key}
	roundKeys := constr.StretchedKey()

	// Apply ShiftRows to round keys 0 to 9.
	for k := 0; k < 10; k++ {
		roundKeys[k] = constr.ShiftRows(roundKeys[k])
	}

	for round := 0; round < 9; round++ {
		for pos := 0; pos < 16; pos++ {
			var inEnc encoding.Byte

			if round == 0 {
				inEnc = encoding.IdentityByte{}
			} else {
				inEnc = encoding.ConcatenatedByte{
					RoundEncoding{round - 1, 2*pos + 0},
					RoundEncoding{round - 1, 2*pos + 1},
				}
			}

			// Build the T-Box and Tyi Table for this round and position in the state matrix.
			out.TBoxTyiTable[round][pos] = encoding.WordTable{
				inEnc,
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
				XOREncoding{round, pos, Left},
				XORTable{},
			}

			out.XORTable[round][pos][1] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					TyiEncoding{round, pos/8*4 + 2, pos % 8},
					TyiEncoding{round, pos/8*4 + 3, pos % 8},
				},
				XOREncoding{round, pos, Right},
				XORTable{},
			}

			out.XORTable[round][pos][2] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					XOREncoding{round, pos, Left},
					XOREncoding{round, pos, Right},
				},
				RoundEncoding{round, 2*shiftRows(pos/2) + pos%2},
				XORTable{},
			}
		}
	}

	// 10th T-Box
	for pos := 0; pos < 16; pos++ {
		out.TBox[pos] = encoding.ByteTable{
			encoding.ConcatenatedByte{
				RoundEncoding{8, 2*pos + 0},
				RoundEncoding{8, 2*pos + 1},
			},
			encoding.IdentityByte{},
			TBox{constr, roundKeys[9][pos], roundKeys[10][pos]},
		}
	}

	return
}
