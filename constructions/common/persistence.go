package common

import (
	"github.com/OpenWhiteBox/primitives/table"
)

const (
	SliceSize  = 4096  // = 256*16
	SlicesSize = 65536 // = 16*SliceSize
)

func SerializeBlockMatrix(dst []byte, m [16]table.Block, xor BlockXORTables) int {
	base := 0

	for _, slice := range m {
		base += copy(dst[base:], table.SerializeBlock(slice))
	}
	base += copy(dst[base:], xor.Serialize())

	return base
}

func ParseBlockSlices(in []byte) (outM [16]table.Block, rest []byte) {
	if in == nil || len(in) < SlicesSize {
		return
	}

	for i := 0; i < 16; i++ {
		outM[i] = table.ParsedBlock(in[SliceSize*i : SliceSize*(i+1)])
	}

	return outM, in[SlicesSize:]
}

func ParseBlockNibbleMatrix(in []byte) (outM [16]table.Block, outXOR NibbleXORTables, rest []byte) {
	outM, rest = ParseBlockSlices(in)
	outXOR, rest = ParseNibbleXORTables(rest)

	return
}

func ParseBlockByteMatrix(in []byte) (outM [16]table.Block, outXOR ByteXORTables, rest []byte) {
	outM, rest = ParseBlockSlices(in)
	outXOR, rest = ParseByteXORTables(rest)

	return
}
