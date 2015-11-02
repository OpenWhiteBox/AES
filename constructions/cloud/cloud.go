package cloud

import (
	"github.com/OpenWhiteBox/AES/primitives/table"
)

type Matrix struct {
	Slices [16]table.Block
	XORs   [32][15]table.Nibble
}

type Construction []Matrix

func (constr Construction) Encrypt(dst, src []byte) {
	constr.crypt(dst, src)
}

func (constr Construction) Decrypt(dst, src []byte) {
	constr.crypt(dst, src)
}

func (constr *Construction) crypt(dst, src []byte) {
	copy(dst, src)

	var stretched [16][16]byte
	for _, m := range *constr {
		stretched = constr.ExpandBlock(m.Slices, dst)
		constr.SquashBlocks(m.XORs, stretched, dst)
	}
}

func (constr *Construction) ExpandBlock(slices [16]table.Block, block []byte) (out [16][16]byte) {
	for i := 0; i < 16; i++ {
		out[i] = slices[i].Get(block[i])
	}

	return
}

func (constr *Construction) SquashBlocks(xorTable [32][15]table.Nibble, blocks [16][16]byte, dst []byte) {
	copy(dst, blocks[0][:])

	for i := 1; i < 16; i++ {
		for pos := 0; pos < 16; pos++ {
			aPartial := dst[pos]&0xf0 | (blocks[i][pos]&0xf0)>>4
			bPartial := (dst[pos]&0x0f)<<4 | blocks[i][pos]&0x0f

			dst[pos] = xorTable[2*pos+0][i-1].Get(aPartial)<<4 | xorTable[2*pos+1][i-1].Get(bPartial)
		}
	}
}
