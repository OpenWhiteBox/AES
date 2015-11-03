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

func split(rs *common.RandomSource, c []byte) (out [16][16]byte) {
	copy(out[15][:], c)
	rand := rs.Stream(c)

	for i := 0; i < 15; i++ {
		rand.Read(out[i][:])

		for pos := 0; pos < 16; pos++ {
			out[15][pos] ^= out[i][pos]
		}
	}

	return
}

func basicEncryption(inputMask, outputMask *matrix.Matrix, roundKeys [11][]byte) []Transform {
	id, inv := [16]table.Byte{}, [16]table.Byte{}

	for pos := 0; pos < 16; pos++ {
		id[pos] = table.IdentityByte{}
		inv[pos] = Invert{}
	}

	out := []Transform{
		Transform{
			Input:    id,
			Linear:   *inputMask,
			Constant: roundKeys[0],
		},
	}

	for round := 1; round <= 9; round++ {
		out = append(out, Transform{
			Input:    inv,
			Linear:   Round,
			Constant: roundKeys[round],
		})
	}

	out = append(out, Transform{
		Input:    inv,
		Linear:   (*outputMask).Compose(LastRound),
		Constant: (*outputMask).Mul(matrix.Row(roundKeys[10])),
	})

	return out
}

func basicDecryption(inputMask, outputMask *matrix.Matrix, roundKeys [11][]byte) []Transform {
	id, inv := [16]table.Byte{}, [16]table.Byte{}

	for pos := 0; pos < 16; pos++ {
		id[pos] = table.IdentityByte{}
		inv[pos] = Invert{}
	}

	out := []Transform{
		Transform{
			Input:    id,
			Linear:   FirstRound.Compose(*inputMask),
			Constant: roundKeys[10],
		},
	}

	for round := 9; round > 0; round-- {
		out = append(out, Transform{
			Input:    inv,
			Linear:   UnRound,
			Constant: roundKeys[round],
		})
	}

	out = append(out, Transform{
		Input:    inv,
		Linear:   *outputMask,
		Constant: (*outputMask).Mul(matrix.Row(roundKeys[0])),
	})

	return out
}
