package chow

import (
	"github.com/OpenWhiteBox/AES/primitives/table"
)

type Construction struct {
	InputMask     [16]table.Block
	InputXORTable [32][15]table.Nibble // [nibble-wise position][gate number]

	TBoxTyiTable [9][16]table.Word      // [round][position]
	HighXORTable [9][32][3]table.Nibble // [round][nibble-wise position][gate number]

	MBInverseTable [9][16]table.Word      // [round][position]
	LowXORTable    [9][32][3]table.Nibble // [round][nibble-wise position][gate number]

	TBoxOutputMask [16]table.Block      // [position]
	OutputXORTable [32][15]table.Nibble // [nibble-wise position][gate number]
}

func (constr *Construction) Encrypt(block [16]byte) [16]byte {
	// Remove input encoding.
	stretched := constr.ExpandBlock(constr.InputMask, block)
	copy(block[:], constr.SquashBlocks(constr.InputXORTable, stretched))

	for round := 0; round < 9; round++ {
		block = constr.ShiftRows(block)

		// Apply the T-Boxes and Tyi Tables to each column of the state matrix.
		for pos := 0; pos < 16; pos += 4 {
			stretched := constr.ExpandWord(constr.TBoxTyiTable[round][pos:pos+4], block[pos:pos+4])
			copy(block[pos:pos+4], constr.SquashWords(constr.HighXORTable[round][2*pos:2*pos+8], stretched))

			stretched = constr.ExpandWord(constr.MBInverseTable[round][pos:pos+4], block[pos:pos+4])
			copy(block[pos:pos+4], constr.SquashWords(constr.LowXORTable[round][2*pos:2*pos+8], stretched))
		}
	}

	block = constr.ShiftRows(block)

	// Apply the final T-Box transformation and add the output encoding.
	stretched = constr.ExpandBlock(constr.TBoxOutputMask, block)
	copy(block[:], constr.SquashBlocks(constr.OutputXORTable, stretched))

	return block
}

func (constr *Construction) ShiftRows(block [16]byte) [16]byte {
	return [16]byte{
		block[0], block[5], block[10], block[15], block[4], block[9], block[14], block[3], block[8], block[13], block[2],
		block[7], block[12], block[1], block[6], block[11],
	}
}

// Expand one word of the state matrix with the T-Boxes composed with Tyi Tables.
func (constr *Construction) ExpandWord(tboxtyi []table.Word, word []byte) [4][4]byte {
	return [4][4]byte{tboxtyi[0].Get(word[0]), tboxtyi[1].Get(word[1]), tboxtyi[2].Get(word[2]), tboxtyi[3].Get(word[3])}
}

// Squash an expanded word back into one word with 3 pairwise XORs (calc'd one nibble at a time) -- (((a ^ b) ^ c) ^ d)
func (constr *Construction) SquashWords(xorTable [][3]table.Nibble, words [4][4]byte) []byte {
	out := make([]byte, 4)
	copy(out, words[0][:])

	for i := 1; i < 4; i++ {
		for pos := 0; pos < 4; pos++ {
			aPartial := out[pos]&0xf0 | (words[i][pos]&0xf0)>>4
			bPartial := (out[pos]&0x0f)<<4 | words[i][pos]&0x0f

			out[pos] = xorTable[2*pos+0][i-1].Get(aPartial)<<4 | xorTable[2*pos+1][i-1].Get(bPartial)
		}
	}

	return out
}

func (constr *Construction) ExpandBlock(mask [16]table.Block, block [16]byte) (out [16][16]byte) {
	for i := 0; i < 16; i++ {
		out[i] = mask[i].Get(block[i])
	}

	return
}

func (constr *Construction) SquashBlocks(xorTable [32][15]table.Nibble, blocks [16][16]byte) []byte {
	out := make([]byte, 16)
	copy(out, blocks[0][:])

	for i := 1; i < 16; i++ {
		for pos := 0; pos < 16; pos++ {
			aPartial := out[pos]&0xf0 | (blocks[i][pos]&0xf0)>>4
			bPartial := (out[pos]&0x0f)<<4 | blocks[i][pos]&0x0f

			out[pos] = xorTable[2*pos+0][i-1].Get(aPartial)<<4 | xorTable[2*pos+1][i-1].Get(bPartial)
		}
	}

	return out
}
