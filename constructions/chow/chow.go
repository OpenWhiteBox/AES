// Package chow implements Chow et al.'s white-box AES construction. There is an attack on this construction
// implemented in the cryptanalysis/chow package. See README.md for more detailed infomration.
//
// "White-Box Cryptography and an AES Implementation" by Stanley Chow, Philip Eisen, Harold Johnson, and Paul C. Van
// Oorschot, http://link.springer.com/chapter/10.1007%2F3-540-36492-7_17?LI=true
//
// "A Tutorial on White-Box AES" by James A. Muir, https://eprint.iacr.org/2013/104.pdf
package chow

import (
	"github.com/OpenWhiteBox/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

type Construction struct {
	InputMask      [16]table.Block // [round]
	InputXORTables common.NibbleXORTables

	TBoxTyiTable [9][16]table.Word      // [round][position]
	HighXORTable [9][32][3]table.Nibble // [round][nibble-wise position][gate number]

	MBInverseTable [9][16]table.Word      // [round][position]
	LowXORTable    [9][32][3]table.Nibble // [round][nibble-wise position][gate number]

	TBoxOutputMask  [16]table.Block // [position]
	OutputXORTables common.NibbleXORTables
}

// BlockSize returns the block size of AES. (Necessary to implement cipher.Block.)
func (constr Construction) BlockSize() int { return 16 }

// Encrypt encrypts the first block in src into dst. Dst and src may point at the same memory.
func (constr Construction) Encrypt(dst, src []byte) {
	constr.crypt(dst, src, constr.shiftRows)
}

// Decrypt decrypts the first block in src into dst. Dst and src may point at the same memory.
func (constr Construction) Decrypt(dst, src []byte) {
	constr.crypt(dst, src, constr.unShiftRows)
}

// crypt pushes the first block in src through the lookup tables (which may compute encryption or decryption) and writes
// the result to dst. shift is the permutation to apply to the state matrix before each round.
func (constr Construction) crypt(dst, src []byte, shift func([]byte)) {
	copy(dst, src[:constr.BlockSize()])

	// Remove input encoding.
	stretched := constr.expandBlock(constr.InputMask, dst)
	constr.InputXORTables.SquashBlocks(stretched, dst)

	for round := 0; round < 9; round++ {
		shift(dst)

		// Apply the T-Boxes and Tyi Tables to each column of the state matrix.
		for pos := 0; pos < 16; pos += 4 {
			stretched := constr.ExpandWord(constr.TBoxTyiTable[round][pos:pos+4], dst[pos:pos+4])
			constr.SquashWords(constr.HighXORTable[round][2*pos:2*pos+8], stretched, dst[pos:pos+4])

			stretched = constr.ExpandWord(constr.MBInverseTable[round][pos:pos+4], dst[pos:pos+4])
			constr.SquashWords(constr.LowXORTable[round][2*pos:2*pos+8], stretched, dst[pos:pos+4])
		}
	}

	shift(dst)

	// Apply the final T-Box transformation and add the output encoding.
	stretched = constr.expandBlock(constr.TBoxOutputMask, dst)
	constr.OutputXORTables.SquashBlocks(stretched, dst)
}

// shiftRows permutes the bytes of the first block of block, according to AES' ShiftRows operation.
func (constr *Construction) shiftRows(block []byte) {
	copy(block, []byte{
		block[0], block[5], block[10], block[15], block[4], block[9], block[14], block[3], block[8], block[13], block[2],
		block[7], block[12], block[1], block[6], block[11],
	})
}

// unShiftRows permutes the bytes of the first block of block, according to the inverse of AES's ShiftRows operation.
func (constr *Construction) unShiftRows(block []byte) {
	copy(block, []byte{
		block[0], block[13], block[10], block[7], block[4], block[1], block[14], block[11], block[8], block[5], block[2],
		block[15], block[12], block[9], block[6], block[3],
	})
}

// ExpandWord expands one word of the state matrix with the T-Boxes composed with Tyi Tables.
func (constr *Construction) ExpandWord(tboxtyi []table.Word, word []byte) [4][4]byte {
	return [4][4]byte{tboxtyi[0].Get(word[0]), tboxtyi[1].Get(word[1]), tboxtyi[2].Get(word[2]), tboxtyi[3].Get(word[3])}
}

// SquashWords squashes an expanded word back into one word with 3 pairwise XORs (calc'd one nibble at a time):
//   (((a ^ b) ^ c) ^ d)
func (constr *Construction) SquashWords(xorTable [][3]table.Nibble, words [4][4]byte, dst []byte) {
	copy(dst, words[0][:])

	for i := 1; i < 4; i++ {
		for pos := 0; pos < 4; pos++ {
			aPartial := dst[pos]&0xf0 | (words[i][pos]&0xf0)>>4
			bPartial := (dst[pos]&0x0f)<<4 | words[i][pos]&0x0f

			dst[pos] = xorTable[2*pos+0][i-1].Get(aPartial)<<4 | xorTable[2*pos+1][i-1].Get(bPartial)
		}
	}
}

// ExpandBlock expands the entire state matrix into sixteen blocks.
func (constr *Construction) expandBlock(mask [16]table.Block, block []byte) (out [16][16]byte) {
	for i := 0; i < 16; i++ {
		out[i] = mask[i].Get(block[i])
	}

	return
}
