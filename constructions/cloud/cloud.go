package cloud

import (
	"github.com/OpenWhiteBox/AES/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

type Matrix struct {
	Slices [16]table.Block
	XORs   common.BlockXORTables
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
		m.XORs.SquashBlocks(stretched, dst)
	}
}

func (constr *Construction) ExpandBlock(slices [16]table.Block, block []byte) (out [16][16]byte) {
	for i := 0; i < 16; i++ {
		out[i] = slices[i].Get(block[i])
	}

	return
}
