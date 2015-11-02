package cloud

import (
	"github.com/OpenWhiteBox/AES/primitives/number"
	"github.com/OpenWhiteBox/AES/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

type Invert struct{}

func (inv Invert) Get(i byte) byte {
	return byte(number.ByteFieldElem(i).Invert())
}

// Generate the XOR Tables for squashing the result of a block matrix.
func blockXORTables() (out [32][15]table.Nibble) {
	for pos := 0; pos < 32; pos++ {
		for i := 0; i < 15; i++ {
			out[pos][i] = common.XORTable{}
		}
	}

	return
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
