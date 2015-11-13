package cloud

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/number"
	"github.com/OpenWhiteBox/AES/primitives/random"
	"github.com/OpenWhiteBox/AES/primitives/table"
)

type Invert struct{}

func (inv Invert) Get(i byte) byte {
	return byte(number.ByteFieldElem(i).Invert())
}

type AddTable byte

func (at AddTable) Get(i byte) byte {
	return i ^ byte(at)
}

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

func idBlock() (out [16]table.Byte) {
	for pos := 0; pos < 16; pos++ {
		out[pos] = table.IdentityByte{}
	}

	return
}

func invBlock() (out [16]table.Byte) {
	for pos := 0; pos < 16; pos++ {
		out[pos] = Invert{}
	}

	return
}

func idTransform() Transform {
	return Transform{
		Input:    idBlock(),
		Linear:   matrix.GenerateIdentity(128),
		Constant: [16]byte{},
	}
}

func addPadding(out *[]Transform, padding int) {
	for i := 0; i < padding; i++ {
		*out = append(*out, idTransform())
	}
}

// basicEncryption returns an unobfuscated array of linear/non-linear transformations that compute AES encryption.
func basicEncryption(inputMask, outputMask *matrix.Matrix, roundKeys [11][]byte, padding []int) []Transform {
	out := []Transform{
		Transform{
			Input:    idBlock(),
			Linear:   *inputMask,
			Constant: [16]byte{},
		},
	}
	copy(out[len(out)-1].Constant[:], roundKeys[0])
	addPadding(&out, padding[0])

	for round := 1; round <= 9; round++ {
		out = append(out, Transform{
			Input:    invBlock(),
			Linear:   Round,
			Constant: [16]byte{},
		})
		copy(out[len(out)-1].Constant[:], roundKeys[round])
		addPadding(&out, padding[round])
	}

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
	out := []Transform{
		Transform{
			Input:    idBlock(),
			Linear:   FirstRound.Compose(*inputMask),
			Constant: [16]byte{},
		},
	}
	copy(out[len(out)-1].Constant[:], roundKeys[10])
	addPadding(&out, padding[0])

	for round := 9; round > 0; round-- {
		out = append(out, Transform{
			Input:    invBlock(),
			Linear:   UnRound,
			Constant: [16]byte{},
		})
		copy(out[len(out)-1].Constant[:], roundKeys[round])
		addPadding(&out, padding[10-round])
	}

	out = append(out, Transform{
		Input:    invBlock(),
		Linear:   *outputMask,
		Constant: [16]byte{},
	})
	copy(out[len(out)-1].Constant[:], (*outputMask).Mul(matrix.Row(roundKeys[0])))

	return out
}

// randomizeFieldInversions takes a padded unobfuscated AES circuit as input and moves the field inversions around so
// that an entire round's inversions won't happen at the same time.  Returns the partitions it chose.
func randomizeFieldInversions(rs *random.Source, aes []Transform, padding []int) [][][]int { // [round][pad]
	partitions := make([][][]int, 10)

	label := make([]byte, 16)
	label[0], label[1] = 'F', 'I'
	stream := rs.Stream(label)

	base := 0

	for round := 0; round < 10; round++ {
		partitions[round] = make([][]int, padding[round]+1)

		// Note on perm and mon:  These two variables give us all of the information we need to randomize where we leave
		// field inversions.  Elements of perm correspond to positions in the state array and elements of mon correspond
		// to indices in perm.  If we're on padding block i and mon[i:i+1] = [x, y], then padding block i will expose and
		// invert elements perm[x:y], assuming perm[:x] are already inverted and perm[y:] are to be left masked.
		label := make([]byte, 16)
		label[0], label[1] = 'M', byte(round)
		mon := append([]int{0}, rs.Monotone(label, padding[round]+1, 16)...)

		perm := RandomPermutation(rs, round)

		// Clear all the inversions in our domain.  We will add them back as we see fit.
		for pad := 0; pad <= padding[round]; pad++ {
			aes[base+pad+1].Input = idBlock()
		}

		for pad := 0; pad < padding[round]; pad++ {
			partitions[round][pad] = perm[mon[pad]:mon[pad+1]]

			mask, maskInv := matrix.GenerateRandomPartial(stream, 128, matrix.IgnoreBytes(perm[:mon[pad+1]]...), matrix.IgnoreNoRows)

			aes[base+pad+0].Linear = mask.Compose(aes[base+pad+0].Linear) // Only exposes some of the round's unmixed input.
			aes[base+pad+1].Linear = maskInv                              // Will expose what the above didn't.

			for _, pos := range perm[mon[pad]:mon[pad+1]] {
				aes[base+pad+1].Input[pos] = Invert{} // This position is exposed so we can invert it.
			}

			for _, pos := range perm[mon[pad+1]:] {
				// Since we don't need it yet, move the rest of the round key down.
				aes[base+pad+1].Constant[pos] = aes[base+pad+0].Constant[pos]
				aes[base+pad+0].Constant[pos] = 0x00
			}
		}

		// Invert what's left.
		partitions[round][padding[round]] = perm[mon[padding[round]]:]

		for _, pos := range perm[mon[padding[round]]:] {
			aes[base+padding[round]+1].Input[pos] = Invert{}
		}

		base += padding[round] + 1
	}

	return partitions
}
