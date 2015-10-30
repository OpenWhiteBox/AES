// Contains tables and encodings necessary for key generation.
package chow

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

// A MB^(-1) Table inverts the mixing bijection on the Tyi Table.
type MBInverseTable struct {
	MBInverse matrix.Matrix
	Row       uint
}

func (mbinv MBInverseTable) Get(i byte) (out [4]byte) {
	r := matrix.Row{0, 0, 0, 0}
	r[mbinv.Row] = i

	res := mbinv.MBInverse.Mul(r)
	copy(out[:], res)

	return
}

// Encodes the output of an input/output mask.
//
//    position: Position in the state array, counted in *bytes*.
// subPosition: Position in the mask's output for this byte, counted in nibbles.
func MaskEncoding(rs *common.RandomSource, position, subPosition int, surface common.Surface) encoding.Nibble {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3], label[4] = 'M', 'E', byte(position), byte(subPosition), byte(surface)

	return rs.Shuffle(label)
}

func BlockMaskEncoding(rs *common.RandomSource, position int, surface common.Surface, shift func(int) int) encoding.Block {
	out := encoding.ConcatenatedBlock{}

	for i := 0; i < 16; i++ {
		out[i] = encoding.ConcatenatedByte{
			MaskEncoding(rs, position, 2*i+0, surface),
			MaskEncoding(rs, position, 2*i+1, surface),
		}

		if surface == common.Inside {
			out[i] = encoding.ComposedBytes{
				encoding.ByteLinear{common.MixingBijection(rs, 8, -1, shift(i)), nil},
				out[i],
			}
		}
	}

	return out
}

// Abstraction over the Tyi and MB^(-1) encodings, to match the pattern of the XOR and round encodings.
func StepEncoding(rs *common.RandomSource, round, position, subPosition int, surface common.Surface) encoding.Nibble {
	if surface == common.Inside {
		return TyiEncoding(rs, round, position, subPosition)
	} else {
		return MBInverseEncoding(rs, round, position, subPosition)
	}
}

func WordStepEncoding(rs *common.RandomSource, round, position int, surface common.Surface) encoding.Word {
	out := encoding.ConcatenatedWord{}

	for i := 0; i < 4; i++ {
		out[i] = encoding.ConcatenatedByte{
			StepEncoding(rs, round, position, 2*i+0, surface),
			StepEncoding(rs, round, position, 2*i+1, surface),
		}
	}

	return out
}

// Encodes the output of a T-Box/Tyi Table / the input of an XOR Table.
//
//    position: Position in the state array, counted in *bytes*.
// subPosition: Position in the T-Box/Tyi Table's ouptput for this byte, counted in nibbles.
func TyiEncoding(rs *common.RandomSource, round, position, subPosition int) encoding.Nibble {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3] = 'T', byte(round), byte(position), byte(subPosition)

	return rs.Shuffle(label)
}

// Encodes the output of a MB^(-1) Table / the input of an XOR Table.
//
//    position: Position in the state array, counted in *bytes*.
// subPosition: Position in the MB^(-1) Table's ouptput for this byte, counted in nibbles.
func MBInverseEncoding(rs *common.RandomSource, round, position, subPosition int) encoding.Nibble {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3], label[4] = 'M', 'I', byte(round), byte(position), byte(subPosition)

	return rs.Shuffle(label)
}

// Encodes intermediate results between each successive XOR.
//
// position: Position in the state array, counted in nibbles.
//     gate: The gate number, or, the number of XORs we've computed so far.
//  surface: Location relative to the round structure. Inside or Outside.
func XOREncoding(rs *common.RandomSource, round, position, gate int, surface common.Surface) encoding.Nibble {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3], label[4] = 'X', byte(round), byte(position), byte(gate), byte(surface)

	return rs.Shuffle(label)
}

// Encodes the output of an Expand->Squash round. Two Expand->Squash rounds make up one AES round.
//
// position: Position in the state array, counted in nibbles.
//  surface: Location relative to the AES round structure. Inside or Outside.
func RoundEncoding(rs *common.RandomSource, round, position int, surface common.Surface) encoding.Nibble {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3] = 'R', byte(round), byte(position), byte(surface)

	return rs.Shuffle(label)
}

func ByteRoundEncoding(rs *common.RandomSource, round, position int, surface common.Surface) encoding.Byte {
	return encoding.ConcatenatedByte{
		RoundEncoding(rs, round, 2*position+0, surface),
		RoundEncoding(rs, round, 2*position+1, surface),
	}
}
