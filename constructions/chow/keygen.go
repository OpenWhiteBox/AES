package chow

import (
	"../../primitives/encoding"
	"../../primitives/table"
	"../saes"
)

func GenerateKeys(key [16]byte, seed [16]byte) (out Construction) {
	constr := saes.Construction{key}
	roundKeys := constr.StretchedKey()

	// Apply ShiftRows to round keys 0 to 9.
	for k := 0; k < 10; k++ {
		roundKeys[k] = constr.ShiftRows(roundKeys[k])
	}

	for round := 0; round < 9; round++ {
		for pos := 0; pos < 16; pos++ {
			mb := WordMixingBijection(seed, round, pos/4)
			mbInv, _ := mb.Invert()

			var inEnc encoding.Byte

			if round == 0 {
				inEnc = encoding.IdentityByte{}
			} else {
				inEnc = encoding.ComposedBytes{
					encoding.ByteLinear(ByteMixingBijection(seed, round-1, pos)),
					encoding.ConcatenatedByte{
						RoundEncoding(seed, round-1, 2*pos+0, Outside),
						RoundEncoding(seed, round-1, 2*pos+1, Outside),
					},
				}
			}

			// Build the T-Box and Tyi Table for this round and position in the state matrix.
			out.TBoxTyiTable[round][pos] = encoding.WordTable{
				inEnc,
				encoding.ComposedWords{
					encoding.ConcatenatedWord{
						encoding.ByteLinear(ByteMixingBijection(seed, round, shiftRows(pos/4*4+0))),
						encoding.ByteLinear(ByteMixingBijection(seed, round, shiftRows(pos/4*4+1))),
						encoding.ByteLinear(ByteMixingBijection(seed, round, shiftRows(pos/4*4+2))),
						encoding.ByteLinear(ByteMixingBijection(seed, round, shiftRows(pos/4*4+3))),
					},
					encoding.WordLinear(mb),
					encoding.ConcatenatedWord{
						encoding.ConcatenatedByte{TyiEncoding(seed, round, pos, 0), TyiEncoding(seed, round, pos, 1)},
						encoding.ConcatenatedByte{TyiEncoding(seed, round, pos, 2), TyiEncoding(seed, round, pos, 3)},
						encoding.ConcatenatedByte{TyiEncoding(seed, round, pos, 4), TyiEncoding(seed, round, pos, 5)},
						encoding.ConcatenatedByte{TyiEncoding(seed, round, pos, 6), TyiEncoding(seed, round, pos, 7)},
					},
				},
				table.ComposedToWord{
					TBox{constr, roundKeys[round][pos], 0},
					TyiTable(pos % 4),
				},
			}

			out.MBInverseTable[round][pos] = encoding.WordTable{
				encoding.ConcatenatedByte{
					RoundEncoding(seed, round, 2*pos+0, Inside),
					RoundEncoding(seed, round, 2*pos+1, Inside),
				},
				encoding.ConcatenatedWord{
					encoding.ConcatenatedByte{MBInverseEncoding(seed, round, pos, 0), MBInverseEncoding(seed, round, pos, 1)},
					encoding.ConcatenatedByte{MBInverseEncoding(seed, round, pos, 2), MBInverseEncoding(seed, round, pos, 3)},
					encoding.ConcatenatedByte{MBInverseEncoding(seed, round, pos, 4), MBInverseEncoding(seed, round, pos, 5)},
					encoding.ConcatenatedByte{MBInverseEncoding(seed, round, pos, 6), MBInverseEncoding(seed, round, pos, 7)},
				},
				MBInverseTable{mbInv, uint(pos) % 4},
			}
		}

		// Generate the High and Low XOR Tables
		for pos := 0; pos < 32; pos++ {
			out.HighXORTable[round][pos][0] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					TyiEncoding(seed, round, pos/8*4+0, pos%8),
					TyiEncoding(seed, round, pos/8*4+1, pos%8),
				},
				XOREncoding(seed, round, pos, Inside, Left),
				XORTable{},
			}

			out.HighXORTable[round][pos][1] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					TyiEncoding(seed, round, pos/8*4+2, pos%8),
					TyiEncoding(seed, round, pos/8*4+3, pos%8),
				},
				XOREncoding(seed, round, pos, Inside, Right),
				XORTable{},
			}

			out.HighXORTable[round][pos][2] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					XOREncoding(seed, round, pos, Inside, Left),
					XOREncoding(seed, round, pos, Inside, Right),
				},
				RoundEncoding(seed, round, pos, Inside),
				XORTable{},
			}

			out.LowXORTable[round][pos][0] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					MBInverseEncoding(seed, round, pos/8*4+0, pos%8),
					MBInverseEncoding(seed, round, pos/8*4+1, pos%8),
				},
				XOREncoding(seed, round, pos, Outside, Left),
				XORTable{},
			}

			out.LowXORTable[round][pos][1] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					MBInverseEncoding(seed, round, pos/8*4+2, pos%8),
					MBInverseEncoding(seed, round, pos/8*4+3, pos%8),
				},
				XOREncoding(seed, round, pos, Outside, Right),
				XORTable{},
			}

			out.LowXORTable[round][pos][2] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					XOREncoding(seed, round, pos, Outside, Left),
					XOREncoding(seed, round, pos, Outside, Right),
				},
				RoundEncoding(seed, round, 2*shiftRows(pos/2)+pos%2, Outside),
				XORTable{},
			}
		}
	}

	// 10th T-Box
	for pos := 0; pos < 16; pos++ {
		out.TBox[pos] = encoding.ByteTable{
			encoding.ComposedBytes{
				encoding.ByteLinear(ByteMixingBijection(seed, 8, pos)),
				encoding.ConcatenatedByte{RoundEncoding(seed, 8, 2*pos+0, Outside), RoundEncoding(seed, 8, 2*pos+1, Outside)},
			},
			encoding.IdentityByte{},
			TBox{constr, roundKeys[9][pos], roundKeys[10][pos]},
		}
	}

	return
}
