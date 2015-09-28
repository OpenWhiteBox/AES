package chow

import (
	"../../primitives/matrix"
	"../../primitives/number"
	"../../primitives/table"
	"../saes"
)

type TBox struct {
	Constr   saes.Construction
	KeyByte1 byte
	KeyByte2 byte
}

func (tbox TBox) Get(i byte) byte {
	return tbox.Constr.SubByte(i^tbox.KeyByte1) ^ tbox.KeyByte2
}

type TyiTable uint

func (tyi TyiTable) Get(i byte) (out uint32) {
	// Calculate dot product of i and [0x02 0x01 0x01 0x03]
	j := number.ByteFieldElem(i)

	a := number.ByteFieldElem(2).Mul(j)
	b := number.ByteFieldElem(1).Mul(j)
	c := number.ByteFieldElem(3).Mul(j)

	// Merge into one output and rotate according to column.
	out = (uint32(a) << 24) | (uint32(b) << 16) | (uint32(b) << 8) | uint32(c)
	out = (out >> (8 * uint(tyi))) | (out << (32 - 8*uint(tyi)))

	return
}

type MBInverseTable struct {
	MBInverse matrix.WordMatrix
	Row       uint
}

func (mbinv MBInverseTable) Get(i byte) uint32 {
	return mbinv.MBInverse.Mul(uint32(i) << (24 - 8*mbinv.Row))
}

type XORTable struct{}

func (xor XORTable) Get(i byte) (out byte) {
	return (i >> 4) ^ (i & 0xf)
}

type Construction struct {
	TBoxTyiTable [9][16]table.Word      // [round][position]
	HighXORTable [9][32][3]table.Nibble // [round][nibble-wise position][gate number]

	MBInverseTable [9][16]table.Word      // [round][position]
	LowXORTable    [9][32][3]table.Nibble // [round][nibble-wise position][gate number]

	TBox [16]table.Byte // [position]
}

func (constr *Construction) Encrypt(block [16]byte) [16]byte {
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

	// Final T-Box transformation.
	for pos := 0; pos < 16; pos++ {
		block[pos] = constr.TBox[pos].Get(block[pos])
	}

	return block
}

func (constr *Construction) ShiftRows(block [16]byte) [16]byte {
	return [16]byte{
		block[0], block[5], block[10], block[15], block[4], block[9], block[14], block[3], block[8], block[13], block[2],
		block[7], block[12], block[1], block[6], block[11],
	}
}

// Expand one word of the state matrix with the T-Boxes composed with Tyi Tables.
func (constr *Construction) ExpandWord(tboxtyi []table.Word, word []byte) [4]uint32 {
	return [4]uint32{
		tboxtyi[0].Get(word[0]),
		tboxtyi[1].Get(word[1]),
		tboxtyi[2].Get(word[2]),
		tboxtyi[3].Get(word[3]),
	}
}

// Squash an expanded word back into one word with 3 pairwise XORs (calc'd one nibble at a time) -- (a ^ b) ^ (c ^ d)
func (constr *Construction) SquashWords(xorTable [][3]table.Nibble, words [4]uint32) (out []byte) {
	acc := uint32(0)
	a, b, c, d := words[0], words[1], words[2], words[3]

	for pos := uint(0); pos < 8; pos++ {
		aPartial := byte((a & (0xf << (28 - 4*pos))) >> (28 - 4*pos))
		bPartial := byte((b & (0xf << (28 - 4*pos))) >> (28 - 4*pos))
		cPartial := byte((c & (0xf << (28 - 4*pos))) >> (28 - 4*pos))
		dPartial := byte((d & (0xf << (28 - 4*pos))) >> (28 - 4*pos))

		ab := xorTable[pos][0].Get(aPartial<<4 | bPartial) // (a ^ b)
		cd := xorTable[pos][1].Get(cPartial<<4 | dPartial) // (c ^ d)

		acc = acc << 4
		acc |= uint32(xorTable[pos][2].Get(ab<<4 | cd)) // (a ^ b) ^ (c ^ d)
	}

	return []byte{byte(acc >> 24), byte(acc >> 16), byte(acc >> 8), byte(acc)}
}
