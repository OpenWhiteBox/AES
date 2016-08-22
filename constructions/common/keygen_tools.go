package common

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/table"
)

// SliceEncoding(position, subPosition int) encoding.Nibble
// Encodes the output of a matrix slice.
//
//    position: Position in the state array, counted in *bytes*.
// subPosition: Position in the matrix's output for this byte, counted in nibbles.

// XOREncoding(position, gate int) encoding.Nibble
// Encodes intermediate results between each successive XOR.
//
// position: Position in the state array, counted in nibbles.
//     gate: The gate number, or, the number of XORs we've computed so far.

// RoundEncoding(position int) encoding.Nibble
// Encodes the output of a matrix multiplication.
//
// position: Position in the state array, counted in nibbles.

// Generate the XOR Tables for squashing the result of a BlockMatrix.
func BlockXORTables(SliceEncoding, XOREncoding func(int, int) encoding.Nibble, RoundEncoding func(int) encoding.Nibble) (out [32][15]table.Nibble) {
	for pos := 0; pos < 32; pos++ {
		out[pos][0] = encoding.NibbleTable{
			encoding.ConcatenatedByte{SliceEncoding(0, pos), SliceEncoding(1, pos)},
			XOREncoding(pos, 0),
			XORTable{},
		}

		for i := 1; i < 14; i++ {
			out[pos][i] = encoding.NibbleTable{
				encoding.ConcatenatedByte{XOREncoding(pos, i-1), SliceEncoding(i+1, pos)},
				XOREncoding(pos, i),
				XORTable{},
			}
		}

		out[pos][14] = encoding.NibbleTable{
			encoding.ConcatenatedByte{XOREncoding(pos, 13), SliceEncoding(15, pos)},
			RoundEncoding(pos),
			XORTable{},
		}
	}

	return
}
