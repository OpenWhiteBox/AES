package common

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

var (
	encodingCache = make(map[[32]byte]encoding.Shuffle)
	mbCache       = make(map[[32]byte]matrix.Matrix)
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
func GenerateMasks(seed []byte, opts KeyGenerationOpts, inputMask, outputMask *matrix.Matrix) {
	switch opts.(type) {
	case IndependentMasks:
		*inputMask = GenerateMask(opts.(IndependentMasks).Input, seed, Inside)
		*outputMask = GenerateMask(opts.(IndependentMasks).Output, seed, Outside)
	case SameMasks:
		mask := GenerateMask(MaskType(opts.(SameMasks)), seed, Inside)
		*inputMask, *outputMask = mask, mask
	case MatchingMasks:
		mask := GenerateMask(RandomMask, seed, Inside)

		*inputMask = mask
		*outputMask, _ = mask.Invert()
	default:
		panic("Unrecognized key generation options!")
	}
}

// Generate byte/word mixing bijections.
// TODO: Ensure that blocks are full-rank.
func MixingBijection(seed []byte, size, round, position int) matrix.Matrix {
	label := make([]byte, 16)
	label[0], label[1], label[2], label[3], label[4] = 'M', 'B', byte(size), byte(round), byte(position)

	key := [32]byte{}
	copy(key[0:16], seed)
	copy(key[16:32], label)

	cached, ok := mbCache[key]

	if ok {
		return cached
	} else {
		mbCache[key] = matrix.GenerateRandom(GenerateStream(seed, label), size)
		return mbCache[key]
	}
}

// Generate input and output masks.
func GenerateMask(maskType MaskType, seed []byte, surface Surface) matrix.Matrix {
	if maskType == RandomMask {
		if surface == Inside {
			return MixingBijection(seed, 128, 0, 0)
		} else {
			return MixingBijection(seed, 128, 10, 0)
		}
	} else { // Identity mask.
		return matrix.GenerateIdentity(128)
	}
}
