package chow

import (
	"../../primitives/encoding"
	"../../primitives/matrix"
	"../../primitives/table"
	"../saes"
)

func GenerateKeys(key [16]byte, seed [16]byte) (out Construction, inputMask, outputMask matrix.Matrix) {
	constr := saes.Construction{key}
	roundKeys := constr.StretchedKey()

	// Apply ShiftRows to round keys 0 to 9.
	for k := 0; k < 10; k++ {
		roundKeys[k] = constr.ShiftRows(roundKeys[k])
	}

	// Generate input and output encodings.
	inputMask = MixingBijection(seed, 128, 0, 0)
	outputMask = MixingBijection(seed, 128, 10, 0)

	// Generate round material.
	for round := 0; round < 9; round++ {
		for pos := 0; pos < 16; pos++ {
			mb := MixingBijection(seed, 32, round, pos/4)
			mbInv, _ := mb.Invert()

			var inEnc encoding.Byte

			if round == 0 {
				inEnc = encoding.ConcatenatedByte{
					RoundEncoding(seed, round-1, 2*pos+0, Outside),
					RoundEncoding(seed, round-1, 2*pos+1, Outside),
				}
			} else {
				inEnc = encoding.ComposedBytes{
					encoding.ByteLinear(MixingBijection(seed, 8, round-1, pos)),
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
						encoding.ByteLinear(MixingBijection(seed, 8, round, shiftRows(pos/4*4+0))),
						encoding.ByteLinear(MixingBijection(seed, 8, round, shiftRows(pos/4*4+1))),
						encoding.ByteLinear(MixingBijection(seed, 8, round, shiftRows(pos/4*4+2))),
						encoding.ByteLinear(MixingBijection(seed, 8, round, shiftRows(pos/4*4+3))),
					},
					encoding.WordLinear(mb),
					WordStepEncoding(seed, round, pos, Inside),
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
				WordStepEncoding(seed, round, pos, Outside),
				MBInverseTable{mbInv, uint(pos) % 4},
			}
		}
	}

	// Generate the High and Low XOR Tables for reach round.
	out.HighXORTable = xorTables(seed, Inside)
	out.LowXORTable = xorTables(seed, Outside)

	// Generate the Input Mask table and the 10th T-Box/Output Mask table
	for pos := 0; pos < 16; pos++ {
		out.InputMask[pos] = encoding.BlockTable{
			encoding.IdentityByte{},
			BlockMaskEncoding(seed, pos, Inside),
			MaskTable{inputMask, pos},
		}

		out.TBoxOutputMask[pos] = encoding.BlockTable{
			encoding.ComposedBytes{
				encoding.ByteLinear(MixingBijection(seed, 8, 8, pos)),
				ByteRoundEncoding(seed, 8, pos, Outside),
			},
			BlockMaskEncoding(seed, pos, Outside),
			table.ComposedToBlock{
				TBox{constr, roundKeys[9][pos], roundKeys[10][pos]},
				MaskTable{outputMask, pos},
			},
		}
	}

	// Generate the XOR Tables for the Input and Output Masks.
	out.InputXORTable = blockXORTables(seed, Inside)
	out.OutputXORTable = blockXORTables(seed, Outside)

	return
}
