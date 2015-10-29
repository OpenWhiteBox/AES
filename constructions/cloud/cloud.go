package cloud

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/number"
)

type Construction struct {
	RoundKeys [11][]byte
}

var sbConst = []byte{
	0x63, 0x63, 0x63, 0x63,
	0x63, 0x63, 0x63, 0x63,
	0x63, 0x63, 0x63, 0x63,
	0x63, 0x63, 0x63, 0x63,
}

func (constr Construction) Encrypt(dst, src []byte) {
	copy(dst, src)

  constr.AddConstant(constr.RoundKeys[0], dst)
  constr.Invert(dst)
  copy(dst, Round.Mul(matrix.Row(dst)))

	for i := 1; i < 9; i++ {
    constr.AddConstant(sbConst, dst)
    constr.AddConstant(constr.RoundKeys[i], dst)

		constr.Invert(dst)
		copy(dst, Round.Mul(matrix.Row(dst)))
	}

  constr.AddConstant(sbConst, dst)
  constr.AddConstant(constr.RoundKeys[9], dst)

	constr.Invert(dst)
	copy(dst, LastRound.Mul(matrix.Row(dst)))

	constr.AddConstant(sbConst, dst)
	constr.AddConstant(constr.RoundKeys[10], dst)
}

func (constr Construction) Decrypt(dst, src []byte) {
	panic("Decryption isn't implemented yet!")
}

func (constr *Construction) AddConstant(c, block []byte) {
	for i, _ := range block {
		block[i] = c[i] ^ block[i]
	}
}

func (constr *Construction) Invert(block []byte) {
	for i, _ := range block {
		block[i] = byte(number.ByteFieldElem(block[i]).Invert())
	}
}
