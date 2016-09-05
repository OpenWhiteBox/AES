// Package xiao implements the Xiao-Lai white-box AES construction. There is an attack on this construction implemented
// in the cryptanalysis/xiao package.
//
// The interface here is very similar to the one presented in the constructions/chow package. Chow's construction is
// based exclusively on representing encryption with randomized lookup tables. Xiao-Lai's construction interleaves
// randomized lookup tables and large linear transformations.
//
// "A Secure Implementation of White-Box AES" by Yaying Xiao and Xuejia Lai,
// http://ieeexplore.ieee.org/xpl/login.jsp?arnumber=5404239
package xiao

import (
	"github.com/OpenWhiteBox/primitives/matrix"
	"github.com/OpenWhiteBox/primitives/table"
)

type Construction struct {
	ShiftRows  [10]matrix.Matrix
	TBoxMixCol [10][8]table.DoubleToWord

	FinalMask matrix.Matrix
}

// BlockSize returns the block size of AES. (Necessary to implement cipher.Block.)
func (constr Construction) BlockSize() int { return 16 }

// Encrypt encrypts the first block in src into dst. Dst and src may point at the same memory.
func (constr Construction) Encrypt(dst, src []byte) {
	constr.crypt(dst, src)
}

// Decrypt decrypts the first block in src into dst. Dst and src may point at the same memory.
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
