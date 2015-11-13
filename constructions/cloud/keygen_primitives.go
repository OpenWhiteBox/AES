package cloud

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/random"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

func RandomPermutation(rs *random.Source, round int) []int {
	out := make([]int, 16)

	label := make([]byte, 16)
	label[0], label[1] = 'Q', byte(round)

	s := rs.Shuffle(label)

	for i := byte(0); i < 16; i++ {
		out[i] = int(s.Encode(i))
	}

	return out
}

func RandomPaddingSizes(rs *random.Source, padding int) []int {
	label := make([]byte, 16)
	label[0], label[1] = 'P', 'S'

	return rs.Dirichlet(label, 10, padding)
}

// See constructions/common/keygen_tools.go
func SliceEncoding(rs *random.Source, round int) func(int, int) encoding.Nibble {
	return func(position, subPosition int) encoding.Nibble {
		label := make([]byte, 16)
		label[0], label[1], label[2], label[3] = 'S', byte(round), byte(position), byte(subPosition)

		return rs.Shuffle(label)
	}
}

// See constructions/common/keygen_tools.go
func XOREncoding(rs *random.Source, round int) func(int, int) encoding.Nibble {
	return func(position, gate int) encoding.Nibble {
		label := make([]byte, 16)
		label[0], label[1], label[2], label[3] = 'X', byte(round), byte(position), byte(gate)

		return rs.Shuffle(label)
	}
}

// See constructions/common/keygen_tools.go
func RoundEncoding(rs *random.Source, size, round int) func(int) encoding.Nibble {
	return func(position int) encoding.Nibble {
		if round == -1 || round == size-1 {
			return encoding.IdentityByte{}
		} else {
			label := make([]byte, 16)
			label[0], label[1], label[2] = 'R', byte(round), byte(position)

			return rs.Shuffle(label)
		}
	}
}

func MixingBijection(rs *random.Source, size, round, position int) encoding.Byte {
	if round == -1 || round == size-1 {
		return encoding.IdentityByte{}
	} else {
		mb := common.MixingBijection(rs, 8, round-1, position)
		mbInv, _ := mb.Invert()

		return encoding.ByteLinear{mb, mbInv}
	}
}

func BlockSliceEncoding(rs *random.Source, size, round, position int) encoding.Block {
	out := encoding.ConcatenatedBlock{}

	for i := 0; i < 16; i++ {
		out[i] = encoding.ComposedBytes{
			MixingBijection(rs, size, round, i),
			encoding.ConcatenatedByte{
				SliceEncoding(rs, round)(position, 2*i+0),
				SliceEncoding(rs, round)(position, 2*i+1),
			},
		}
	}

	return out
}

func ByteRoundEncoding(rs *random.Source, size, round, position int) encoding.Byte {
	return encoding.ConcatenatedByte{
		RoundEncoding(rs, size, round)(2*position + 0),
		RoundEncoding(rs, size, round)(2*position + 1),
	}
}
