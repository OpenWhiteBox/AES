package common

import (
	"github.com/OpenWhiteBox/AES/primitives/table"
)

const (
	SliceSize = 4096 // = 256*16
	XorSize   = 128  // = 256*0.5

	BlockMatrixSize = 126976 // = 16*SliceSize + 32*15*XorSize
)

func SerializeBlockMatrix(dst []byte, m [16]table.Block, xor [32][15]table.Nibble) int {
	base := 0

	for _, slice := range m {
		base += copy(dst[base:], table.SerializeBlock(slice))
	}

	for _, rack := range xor {
		for _, xorTable := range rack {
			base += copy(dst[base:], table.SerializeNibble(xorTable))
		}
	}

	return base
}

func ParseBlockMatrix(in []byte) (outM [16]table.Block, outXOR [32][15]table.Nibble, rest []byte) {
	if in == nil || len(in) < BlockMatrixSize {
		return
	}

	for i := 0; i < 16; i++ {
		outM[i] = table.ParsedBlock(in[SliceSize*i : SliceSize*(i+1)])
	}

	rest = in[16*SliceSize:]

	for i := 0; i < 32; i++ {
		for j := 0; j < 15; j++ {
			loc := 15*i + j
			outXOR[i][j] = table.ParsedNibble(rest[XorSize*loc : XorSize*(loc+1)])
		}
	}

	rest = rest[32*15*XorSize:]

	return
}
