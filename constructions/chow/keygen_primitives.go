package chow

import (
	"github.com/OpenWhiteBox/primitives/encoding"
	"github.com/OpenWhiteBox/primitives/matrix"
	"github.com/OpenWhiteBox/primitives/random"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

// MBInverseTable, or a MB^(-1) Table, inverts the mixing bijection on the Tyi Table. It is the second half of a round.
// It implements table.Word.
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

// MaskEncoding produces encodings for the outputs of the InputMask and OutputMask. All randomness is derived from the
// random source; surface is common.Inside if these will be the masks between InputMask and InputXORTables or
// common.Outside if they'll be between TBoxOutputMask and OutputXORTables.
//
// See constructions/common/keygen_tools.go for information on the function returned.
func MaskEncoding(rs *random.Source, surface common.Surface) func(int, int) encoding.Nibble {
	return func(position, subPosition int) encoding.Nibble {
		label := make([]byte, 16)
		label[0], label[1], label[2], label[3], label[4] = 'M', 'E', byte(position), byte(subPosition), byte(surface)

		return rs.Shuffle(label)
	}
}

// XOREncoding produces encodings for intermediate values of XOR tables. All randomness is derived from the random
// source.
//
// If round < 10:
//   surface = common.Inside -- XOREncoding generates the encodings for the
//     HighXORTable (from TBoxTyiTable) in the given round.
//   surface = common.OUtside -- XOREncoding generates the encodings for the
//     LowXORTable (from MBInverseTable) in the given round.
//
// If round = 10:
//   surface = common.Inside -- XOREncoding generates the encodings for
//     InputXORTables (from InputMask).
//   surface = common.Outside -- XOREncoding generates the encodings for
//     OutputXORTables (from TBoxOutputMask).
//
// See constructions/common/keygen_tools.go for information on the function returned.
func XOREncoding(rs *random.Source, round int, surface common.Surface) func(int, int) encoding.Nibble {
	return func(position, gate int) encoding.Nibble {
		label := make([]byte, 16)
		label[0], label[1], label[2], label[3], label[4] = 'X', byte(round), byte(position), byte(gate), byte(surface)

		return rs.Shuffle(label)
	}
}

// RoundEncoding produces encodings for the output of a series of XOR tables / the input of a TBoxTyiTable or
// MBInverseTable. All randomness is derived from the random source; shift is the permutation that will be applied to
// the state matrix between the output of the XOR tables and the input of the next, or noshift if this is an input
// encoding.
//
// surface = common.Inside is used for "inter-round" encodings, like those between a HighXORTable and a MBInverseTable.
// surface = common.Outside is used for "intra-round" encodings, like between the InputXORTables and and the first
// TBoxTyiTable.
//
// See constructions/common/keygen_tools.go for information on the function returned.
func RoundEncoding(rs *random.Source, round int, surface common.Surface, shift func(int) int) func(int) encoding.Nibble {
	return func(position int) encoding.Nibble {
		position = 2*shift(position/2) + position%2

		label := make([]byte, 16)
		label[0], label[1], label[2], label[3] = 'R', byte(round), byte(position), byte(surface)

		return rs.Shuffle(label)
	}
}

// BlockMaskEncoding concatenates all the mask encodings for InputMask or TBoxOutputMask into a block encoding, so that
// it can easily be put on the output of one of the Block tables.
//
// position is the index of the Block table and shift is the permutation that will be applied between this round and the
// next or noshift if this is an input encoding; the other parameters are explained in MaskEncoding documentation.
func BlockMaskEncoding(rs *random.Source, position int, surface common.Surface, shift func(int) int) encoding.Block {
	out := encoding.ConcatenatedBlock{}

	for i := 0; i < 16; i++ {
		out[i] = encoding.ConcatenatedByte{
			MaskEncoding(rs, surface)(position, 2*i+0),
			MaskEncoding(rs, surface)(position, 2*i+1),
		}

		if surface == common.Inside {
			out[i] = encoding.ComposedBytes{
				encoding.NewByteLinear(common.MixingBijection(rs, 8, -1, shift(i))),
				out[i],
			}
		}
	}

	return out
}

// StepEncoding returns a TyiEncoding if surface = common.Inside and a MBInverseEncoding if surface = common.Outside.
// It transparently swaps the two in the code that generates HighXORTable and LowXORTable.

// All randomness is derived from the random source. round is the current round; position is the byte-wise position in
// the state matrix that's being stretched; subPosition is the nibble-wise position in the Word table's output.
func StepEncoding(rs *random.Source, round, position, subPosition int, surface common.Surface) encoding.Nibble {
	if surface == common.Inside {
		return TyiEncoding(rs, round, position, subPosition)
	} else {
		return MBInverseEncoding(rs, round, position, subPosition)
	}
}

// WordStepEncoding concatenates all the step encodings for the full output of a Word table in TBoxTyiTable or
// MBInverseTable. Function parameters are explained in the StepEncoding documentation.
func WordStepEncoding(rs *random.Source, round, position int, surface common.Surface) encoding.Word {
	out := encoding.ConcatenatedWord{}

	for i := 0; i < 4; i++ {
		out[i] = encoding.ConcatenatedByte{
			StepEncoding(rs, round, position, 2*i+0, surface),
			StepEncoding(rs, round, position, 2*i+1, surface),
		}
	}

	return out
}

// TyiEncoding encodes the output of a T-Box/Tyi Table / the input of a HighXORTable.
//
// All randomness is derived from the random source; round is the current round; position is the byte-wise position in
// the state matrix being stretched; subPosition is the nibble-wise position in the Word table's output.
func TyiEncoding(rs *random.Source, round, position, subPosition int) encoding.Nibble {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3] = 'T', byte(round), byte(position), byte(subPosition)

	return rs.Shuffle(label)
}

// MBInverseEncoding encodes the output of a MB^(-1) Table / the input of a LowXORTable.
//
// All randomness is derived from the random source; round is the current round; position is the byte-wise position in
// the state matrix being stretched; subPosition is the nibble-wise position in the Word table's output.
func MBInverseEncoding(rs *random.Source, round, position, subPosition int) encoding.Nibble {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3], label[4] = 'M', 'I', byte(round), byte(position), byte(subPosition)

	return rs.Shuffle(label)
}

// ByteRoundEncoding concatenates all the round encodings for a single byte. Function parameters are explained in
// RoundEncoding documentation.
func ByteRoundEncoding(rs *random.Source, round, position int, surface common.Surface, shift func(int) int) encoding.Byte {
	return encoding.ConcatenatedByte{
		RoundEncoding(rs, round, surface, shift)(2*position + 0),
		RoundEncoding(rs, round, surface, shift)(2*position + 1),
	}
}
