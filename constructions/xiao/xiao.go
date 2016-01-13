// Package xiao implements the Xiao-Lai white-box AES construction.
package xiao

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/table"
)

type Construction struct {
	ShiftRows  [10]matrix.Matrix
	TBoxMixCol [10][8]table.DoubleToWord

	FinalMask matrix.Matrix
}

func (constr Construction) BlockSize() int { return 16 }

func (constr Construction) Encrypt(dst, src []byte) {
	constr.crypt(dst, src)
}

func (constr Construction) Decrypt(dst, src []byte) {
	constr.crypt(dst, src)
}

func (constr *Construction) crypt(dst, src []byte) {
	copy(dst, src)

	for round := 0; round < 10; round++ {
		// ShiftRows and re-encoding step.
		copy(dst, constr.ShiftRows[round].Mul(matrix.Row(dst)))

		// Apply T-Boxes and MixColumns
		for pos := 0; pos < 16; pos += 4 {
			stretched := constr.ExpandWord(constr.TBoxMixCol[round][pos/2:(pos+4)/2], dst[pos:pos+4])
			constr.SquashWords(stretched, dst[pos:pos+4])
		}
	}

	copy(dst, constr.FinalMask.Mul(matrix.Row(dst)))
}

func (constr *Construction) ExpandWord(tmc []table.DoubleToWord, word []byte) [2][4]byte {
	return [2][4]byte{
		tmc[0].Get([2]byte{word[0], word[1]}),
		tmc[1].Get([2]byte{word[2], word[3]}),
	}
}

func (constr *Construction) SquashWords(words [2][4]byte, dst []byte) {
	for i := 0; i < 4; i++ {
		dst[i] = words[0][i] ^ words[1][i]
	}
}
