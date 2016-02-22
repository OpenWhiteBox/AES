package common

import (
	"github.com/OpenWhiteBox/primitives/table"
)

const (
	nxtSize  = 128
	nxtsSize = 61440

	bxtSize  = 65536
	bxtsSize = 15728640
)

type BlockXORTables interface {
	SquashBlocks(blocks [16][16]byte, dst []byte)
	Serialize() []byte
}

// Computes the XOR of two nibbles.
type NibbleXORTable struct{}

func (nxt NibbleXORTable) Get(i byte) (out byte) {
	return (i >> 4) ^ (i & 0xf)
}

// Computes the XOR of two bytes.
type ByteXORTable struct{}

func (bxt ByteXORTable) Get(i [2]byte) (out byte) {
	return i[0] ^ i[1]
}

type NibbleXORTables [32][15]table.Nibble // [nibble-wise position][gate number]

func ParseNibbleXORTables(in []byte) (nxts NibbleXORTables, rest []byte) {
	if in == nil || len(in) < nxtsSize {
		return nxts, nil
	}

	for i := 0; i < 32; i++ {
		for j := 0; j < 15; j++ {
			loc := 15*i + j
			nxts[i][j] = table.ParsedNibble(in[nxtSize*loc : nxtSize*(loc+1)])
		}
	}

	return nxts, in[nxtsSize:]
}

func (nxts NibbleXORTables) SquashBlocks(blocks [16][16]byte, dst []byte) {
	copy(dst, blocks[0][:])

	for i := 1; i < 16; i++ {
		for pos := 0; pos < 16; pos++ {
			aPartial := dst[pos]&0xf0 | (blocks[i][pos]&0xf0)>>4
			bPartial := (dst[pos]&0x0f)<<4 | blocks[i][pos]&0x0f

			dst[pos] = nxts[2*pos+0][i-1].Get(aPartial)<<4 | nxts[2*pos+1][i-1].Get(bPartial)
		}
	}
}

func (nxts NibbleXORTables) Serialize() []byte {
	dst, base := make([]byte, nxtsSize), 0

	for _, rack := range nxts {
		for _, xorTable := range rack {
			base += copy(dst[base:], table.SerializeNibble(xorTable))
		}
	}

	return dst
}

type ByteXORTables [16][15]table.DoubleToByte // [byte-wise position][gate number]

func ParseByteXORTables(in []byte) (bxts ByteXORTables, rest []byte) {
	if in == nil || len(in) < bxtsSize {
		return bxts, nil
	}

	tableSize := 256 * 256

	for i := 0; i < 16; i++ {
		for j := 0; j < 15; j++ {
			loc := 15*i + j
			bxts[i][j] = table.ParsedDoubleToByte(in[tableSize*loc : tableSize*(loc+1)])
		}
	}

	return bxts, in[bxtsSize:]
}

func (bxts ByteXORTables) SquashBlocks(blocks [16][16]byte, dst []byte) {
	copy(dst, blocks[0][:])

	for i := 1; i < 16; i++ {
		for pos := 0; pos < 16; pos++ {
			dst[pos] = bxts[pos][i-1].Get([2]byte{dst[pos], blocks[i][pos]})
		}
	}
}

func (bxts ByteXORTables) Serialize() []byte {
	dst, base := make([]byte, bxtsSize), 0

	for _, rack := range bxts {
		for _, xorTable := range rack {
			base += copy(dst[base:], table.SerializeDoubleToByte(xorTable))
		}
	}

	return dst
}
