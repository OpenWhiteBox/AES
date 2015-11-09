package cloud

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/number"
	"github.com/OpenWhiteBox/AES/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/common"
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
func splitSecret(rs *common.RandomSource, c [16]byte, n int) (out [][16]byte) {
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
func basicEncryption(inputMask, outputMask *matrix.Matrix, roundKeys [11][]byte, padding int) []Transform {
	out := []Transform{
		Transform{
			Input:    idBlock(),
			Linear:   *inputMask,
			Constant: [16]byte{},
		},
	}
	copy(out[len(out)-1].Constant[:], roundKeys[0])
	addPadding(&out, padding)

	for round := 1; round <= 9; round++ {
		out = append(out, Transform{
			Input:    invBlock(),
			Linear:   Round,
			Constant: [16]byte{},
		})
		copy(out[len(out)-1].Constant[:], roundKeys[round])
		addPadding(&out, padding)
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
func basicDecryption(inputMask, outputMask *matrix.Matrix, roundKeys [11][]byte, padding int) []Transform {
	out := []Transform{
		Transform{
			Input:    idBlock(),
			Linear:   FirstRound.Compose(*inputMask),
			Constant: [16]byte{},
		},
	}
	copy(out[len(out)-1].Constant[:], roundKeys[10])
	addPadding(&out, padding)

	for round := 9; round > 0; round-- {
		out = append(out, Transform{
			Input:    invBlock(),
			Linear:   UnRound,
			Constant: [16]byte{},
		})
		copy(out[len(out)-1].Constant[:], roundKeys[round])
		addPadding(&out, padding)
	}

	out = append(out, Transform{
		Input:    invBlock(),
		Linear:   *outputMask,
		Constant: [16]byte{},
	})
	copy(out[len(out)-1].Constant[:], (*outputMask).Mul(matrix.Row(roundKeys[0])))

	return out
}

// randomizeFieldInversions takes a padded unobfuscated AES circuit as input and moves field inversions around so that
// an entire round's inversions won't happen at the same time.
func randomizeFieldInversions(rs *common.RandomSource, aes []Transform) {
	for round := 0; round < 10; round++ {
		perm := RandomPermutation(rs, round) // The first 8 values are inverted immediately, the last 8 are deferred.
		mask, _ := aes[2*round+0].Linear.Invert()

		a, _ := mask.InvertAt(perm[:8]...)            // `a` only exposes some of the round's unmasked input.
		b, _ := a.Compose(mask).InvertAt(perm[8:]...) // `b` exposes what `a` didn't w/o full-knowledge of the masked input.

		aes[2*round+0].Linear, aes[2*round+1].Linear = a, b // `mask` is now split into two matrix multiplications.
		aes[2*round+1].Input, aes[2*round+2].Input = idBlock(), idBlock()

		for _, pos := range perm[:8] {
			aes[2*round+1].Input[pos] = Invert{}             // Since `a` exposed this position so we can invert it.
			aes[2*round+2].Input[pos] = table.IdentityByte{} // This position was already inverted above so we leave it alone.
		}

		for _, pos := range perm[8:] {
			// Split the round key in the same way the field inversions were split.
			aes[2*round+1].Constant[pos] = aes[2*round+0].Constant[pos]
			aes[2*round+0].Constant[pos] = 0x00

			aes[2*round+1].Input[pos] = table.IdentityByte{} // `a` left this hidden so we have to ignore it.
			aes[2*round+2].Input[pos] = Invert{}             // `b` finally exposed this position so we can invert it.
		}
	}
}
