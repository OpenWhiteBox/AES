package chow

import (
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

type XORTable struct{}

func (xor XORTable) Get(i byte) (out byte) {
	return (i >> 4) ^ (i & 0xf)
}

type Construction struct {
	TBoxTyiTable [9][16]table.WordTable         // [round][position]
	TBox         [16]table.ByteTable            // [position]
	XORTable     [9][16][3][2]table.NibbleTable // [round][position][level][side]
}

func GenerateTables(key [16]byte) (tyi [9][16]table.WordTable, tbox [16]table.ByteTable, xor [9][16][3][2]table.NibbleTable) {
	constr := saes.Construction{key}
	roundKeys := constr.StretchedKey()

	// Apply ShiftRows to round keys 0 to 9.
	for k := 0; k < 10; k++ {
		roundKeys[k] = constr.ShiftRows(roundKeys[k])
	}

	// Build T-Boxes and Tyi Tables 1 to 9
	for round := 0; round < 9; round++ {
		for pos := 0; pos < 16; pos++ {
			tyi[round][pos] = table.ComposedToWordTable{
				TBox{constr, roundKeys[round][pos], 0},
				TyiTable(pos % 4),
			}
		}
	}

	// 10th T-Box
	for pos := 0; pos < 16; pos++ {
		tbox[pos] = TBox{constr, roundKeys[9][pos], roundKeys[10][pos]}
	}

	// Generate XOR Tables
	for round := 0; round < 9; round++ {
		for pos := 0; pos < 16; pos++ {
			for level := 0; level < 3; level++ {
				for side := 0; side < 2; side++ {
					xor[round][pos][level][side] = XORTable{}
				}
			}
		}
	}

	return
}

func (constr *Construction) Encrypt(block [16]byte) [16]byte {
	for round := 0; round < 9; round++ {
		block = constr.ShiftRows(block)

		// Apply the T-Boxes and Tyi Tables to each column of the state matrix.
		for pos := 0; pos < 16; pos += 4 {
			stretched := constr.ExpandWord(constr.TBoxTyiTable[round][pos:pos+4], block[pos:pos+4])
			squashed := constr.SquashWords(constr.XORTable[round][pos:pos+4], stretched)

			// Split squashed word into 4 bytes and put it away.
			block[pos+0] = byte(squashed >> 24)
			block[pos+1] = byte(squashed >> 16)
			block[pos+2] = byte(squashed >> 8)
			block[pos+3] = byte(squashed)
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
func (constr *Construction) ExpandWord(tboxtyi []table.WordTable, word []byte) [4]uint32 {
	return [4]uint32{
		tboxtyi[0].Get(word[0]),
		tboxtyi[1].Get(word[1]),
		tboxtyi[2].Get(word[2]),
		tboxtyi[3].Get(word[3]),
	}
}

// Squash expanded word back into one word with 3 pairwise XORs (calc'd one nibble at a time) -- (a ^ b) ^ (c ^ d)
func (constr *Construction) SquashWords(xorTable [][3][2]table.NibbleTable, words [4]uint32) (out uint32) {
	a, b, c, d := words[0], words[1], words[2], words[3]

	for pos := 0; pos < 4; pos++ {
		for side := 0; side < 2; side++ { // side, as in left or right side of the byte.
			abPartial := byte(((a & 0xf) << 4) | (b & 0xf))
			cdPartial := byte(((c & 0xf) << 4) | (d & 0xf))

			ab := xorTable[pos][0][side].Get(abPartial) // (a ^ b)
			cd := xorTable[pos][1][side].Get(cdPartial) // (c ^ d)

			adPartial := (ab << 4) | cd

			offset := uint(4 * (2*pos + side))
			out |= uint32(xorTable[pos][2][side].Get(adPartial)) << offset // (a ^ b) ^ (c ^ d)

			a, b, c, d = a>>4, b>>4, c>>4, d>>4
		}
	}

	return
}
