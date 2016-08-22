package common

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

type Surface int

const (
	Inside Surface = iota
	Outside
)

type MaskType int

const (
	RandomMask MaskType = iota
	IdentityMask
)

type KeyGenerationOpts interface{}

// IndependentMasks generates the input and output masks independently of each other.
type IndependentMasks struct {
	Input, Output MaskType
}

// SameMasks puts the exact same mask on the input and output of the white-box.
type SameMasks MaskType

// MatchingMasks implies a randomly generated input mask and the inverse mask on the output.
type MatchingMasks struct{}

// GenerateMasks generates input and output encodings for a white-box AES construction.
func GenerateMasks(rs *RandomSource, opts KeyGenerationOpts, inputMask, outputMask *matrix.Matrix) {
	switch opts.(type) {
	case IndependentMasks:
		*inputMask = generateMask(rs, opts.(IndependentMasks).Input, Inside)
		*outputMask = generateMask(rs, opts.(IndependentMasks).Output, Outside)
	case SameMasks:
		mask := generateMask(rs, MaskType(opts.(SameMasks)), Inside)
		*inputMask, *outputMask = mask, mask
	case MatchingMasks:
		mask := generateMask(rs, RandomMask, Inside)

		*inputMask = mask
		*outputMask, _ = mask.Invert()
	default:
		panic("Unrecognized key generation options!")
	}
}

func generateMask(rs *RandomSource, maskType MaskType, surface Surface) matrix.Matrix {
	if maskType == RandomMask {
		label := make([]byte, 16)

		if surface == Inside {
			copy(label[:], []byte("MASK Inside"))
			return rs.Matrix(label, 128)
		} else {
			copy(label[:], []byte("MASK Outside"))
			return rs.Matrix(label, 128)
		}
	} else { // Identity mask.
		return matrix.GenerateIdentity(128)
	}
}

// Generate byte/word mixing bijections.
// TODO: Ensure that blocks are full-rank.
func MixingBijection(rs *RandomSource, size, round, position int) matrix.Matrix {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3], label[4] = 'M', 'B', byte(size), byte(round), byte(position)

	return rs.Matrix(label, size)
}

type BlockMatrix struct {
	Linear   matrix.Matrix
	Constant [16]byte
	Position int
}

func (bm BlockMatrix) Get(i byte) (out [16]byte) {
	r := make([]byte, 16)
	r[bm.Position] = i

	res := bm.Linear.Mul(matrix.Row(r))
	copy(out[:], res)

	for i, c := range bm.Constant {
		out[i] ^= c
	}

	return
}

// An XOR Table computes the XOR of two nibbles.
type XORTable struct{}

func (xor XORTable) Get(i byte) (out byte) {
	return (i >> 4) ^ (i & 0xf)
}
