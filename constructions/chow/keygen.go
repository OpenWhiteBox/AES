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
	}

	// Generate the High and Low XOR Tables
	out.HighXORTable = xorTables(seed, Inside)
	out.LowXORTable = xorTables(seed, Outside)

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

func xorTables(seed [16]byte, surface Surface) (out [9][32][3]table.Nibble) {
	var outPos func(int) int
	if surface == Inside {
		outPos = func(pos int) int { return pos }
	} else {
		outPos = func(pos int) int { return 2*shiftRows(pos/2) + pos%2 }
	}

	for round := 0; round < 9; round++ {
		for pos := 0; pos < 32; pos++ {
			out[round][pos][0] = topLevelXORTable(seed, round, pos, surface, Left)
			out[round][pos][1] = topLevelXORTable(seed, round, pos, surface, Right)
			out[round][pos][2] = encoding.NibbleTable{
				encoding.ConcatenatedByte{
					XOREncoding(seed, round, pos, surface, Left),
					XOREncoding(seed, round, pos, surface, Right),
				},
				RoundEncoding(seed, round, outPos(pos), surface),
				XORTable{},
			}
		}
	}

	return
}

func topLevelXORTable(seed [16]byte, round, pos int, surface Surface, side Side) table.Nibble {
	return encoding.NibbleTable{
		encoding.ConcatenatedByte{
			StepEncoding(seed, round, pos/8*4+2*int(side)+0, pos%8, surface),
			StepEncoding(seed, round, pos/8*4+2*int(side)+1, pos%8, surface),
		},
		XOREncoding(seed, round, pos, surface, side),
		XORTable{},
	}
}
