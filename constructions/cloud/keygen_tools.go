package cloud

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/random"
	"github.com/OpenWhiteBox/AES/primitives/table"
)

// splitSecret takes a secret (like a round key) and splits it into many shares.  All must be XORed together to recover
// the original value.
func splitSecret(rs *random.Source, c [16]byte, n int) (out [][16]byte) {
	out = append(out, c)
	rand := rs.Stream(c[:])

	for i := 1; i < n; i++ {
		out = append(out, [16]byte{})
		rand.Read(out[i][:])

		for pos := 0; pos < 16; pos++ {
			out[0][pos] ^= out[i][pos]
		}
	}

	return
}

// idBlock, invBlock, and addPadding are helper functions for basicEncrypt and basicDecrypt.

func idBlock() (out [16]table.Byte) {
	for pos := 0; pos < 16; pos++ {
		out[pos] = table.IdentityByte{}
	}

	return
}

func invBlock() (out [16]table.Byte) {
	for pos := 0; pos < 16; pos++ {
		out[pos] = InvertTable{}
	}

	return
}

func addPadding(out *[]Transform, padding int) {
	for i := 0; i < padding; i++ {
		*out = append(*out, Transform{
			Input:    idBlock(),
			Linear:   matrix.GenerateIdentity(128),
			Constant: [16]byte{},
		})
	}
}

// basicEncryption returns an unobfuscated array of linear/non-linear transformations that compute AES encryption.
func basicEncryption(inputMask, outputMask *matrix.Matrix, roundKeys [11][]byte, padding []int) []Transform {
	out := []Transform{}

	addPadding(&out, padding[0])
	out = append(out, Transform{
		Input:    idBlock(),
		Linear:   *inputMask,
		Constant: [16]byte{},
	})
	copy(out[len(out)-1].Constant[:], roundKeys[0])

	for round := 1; round <= 9; round++ {
		addPadding(&out, padding[round])
		out = append(out, Transform{
			Input:    invBlock(),
			Linear:   Round,
			Constant: [16]byte{},
		})
		copy(out[len(out)-1].Constant[:], roundKeys[round])
	}

	addPadding(&out, padding[10])
	out = append(out, Transform{
		Input:    invBlock(),
		Linear:   (*outputMask).Compose(LastRound),
		Constant: [16]byte{},
	})
	copy(out[len(out)-1].Constant[:], (*outputMask).Mul(matrix.Row(roundKeys[10])))

	return out
}

// basicDecryption returns an unobfuscated array of linear/non-linear transformations that compute AES decryption.
func basicDecryption(inputMask, outputMask *matrix.Matrix, roundKeys [11][]byte, padding []int) []Transform {
	out := []Transform{}

	addPadding(&out, padding[0])
	out = append(out, Transform{
		Input:    idBlock(),
		Linear:   FirstRound.Compose(*inputMask),
		Constant: [16]byte{},
	})
	copy(out[len(out)-1].Constant[:], roundKeys[10])

	for round := 9; round >= 1; round-- {
		addPadding(&out, padding[10-round])
		out = append(out, Transform{
			Input:    invBlock(),
			Linear:   UnRound,
			Constant: [16]byte{},
		})
		copy(out[len(out)-1].Constant[:], roundKeys[round])
	}

	addPadding(&out, padding[10])
	out = append(out, Transform{
		Input:    invBlock(),
		Linear:   *outputMask,
		Constant: [16]byte{},
	})
	copy(out[len(out)-1].Constant[:], (*outputMask).Mul(matrix.Row(roundKeys[0])))

	return out
}

// randomizeFieldInversions takes a padded unobfuscated AES circuit as input and moves the field inversions around so
// that an entire round's inversions won't happen at the same time.
func randomizeFieldInversions(rs *random.Source, aes []Transform, padding []int) {
	base := 0

	for round := 0; round < 11; round++ {
		top := base + padding[round]

		// Note on perm and mon:  These two variables give us all of the information we need to randomize where we leave
		// field inversions.  Elements of perm correspond to positions in the state array and elements of mon correspond
		// to indices in perm.  If we're on padding block i and mon[i:i+1] = [x, y], then padding block i will expose and
		// invert elements perm[x:y], assuming perm[:x] are already inverted and perm[y:] are to be left masked.
		label := make([]byte, 16)
		label[0], label[1] = 'M', byte(round)
		mon := append([]int{0}, rs.Monotone(label, padding[round]+1, 16)...)
		perm := RandomPermutation(rs, round)

		for pad := 0; pad < padding[round]; pad++ {
			for _, pos := range perm[mon[pad]:mon[pad+1]] {
				aes[top].Input[pos], aes[base].Input[pos] = aes[base].Input[pos], aes[top].Input[pos]
			}

			base++
		}

		base++
	}
}

// blurRoundBoundaries takes an AES circuit with randomized field inversions and blurs the round boundaries with
// random affine transformations that go through the gaps between inversions.
func interlockRounds(rs *random.Source, aes []Transform) {
	label := make([]byte, 16)
	label[0], label[1] = 'B', 'R'
	stream := rs.Stream(label)

	// Add random mixing bijections (avoiding S-boxes) between each block.
	for base := 0; base < len(aes)-1; base++ {
		positions := []int{}
		for k, c := range aes[base+1].Input {
			switch c.(type) {
			case InvertTable:
				positions = append(positions, k)
			}
		}

		// Generate a random affine transformation and its inverse which is the identity on $positions.
		mask, maskInv := matrix.GenerateRandomPartial(
			stream, 128, matrix.IgnoreBytes(positions...), matrix.IgnoreNoRows,
		)
		constant, constantInv := [16]byte{}, [16]byte{}

		stream.Read(constant[:])
		copy(constantInv[:], maskInv.Mul(matrix.Row(constant[:])))

		for _, pos := range positions {
			constant[pos], constantInv[pos] = 0x00, 0x00
		}

		aes[base].Linear, aes[base].Constant = ComposeAffine(
			mask, aes[base].Linear, constant, aes[base].Constant,
		)

		aes[base+1].Linear, aes[base+1].Constant = ComposeAffine(
			aes[base+1].Linear, maskInv, aes[base+1].Constant, constantInv,
		)
	}
}
