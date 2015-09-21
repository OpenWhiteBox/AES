package saes

import (
	"../saes"
	"../../primitives/number"
)

// Powers of x mod M(x).
var powx = [16]byte{
	0x01,
	0x02,
	0x04,
	0x08,
	0x10,
	0x20,
	0x40,
	0x80,
	0x1b,
	0x36,
	0x6c,
	0xd8,
	0xab,
	0x4d,
	0x9a,
	0x2f,
}

type Construction struct {
	TBox [10][16][256]byte
}

func GenerateTables(key [16]byte) (table [10][16][256]byte) {
	constr := saes.Construction{key}
	roundKeys := constr.StretchedKey()

	// Apply ShiftRows to round keys 0 to 9.
	for k := 0; k < 10; k++ {
		roundKeys[k] = constr.ShiftRows(roundKeys[k])
	}

	// Build T-Boxes 1 to 9
	for round := 1; round < 10; round++ {
		table[round] = [16][256]byte{}

		for place := 0; place < 16; place++ {
			table[round][place] = [256]byte{}

			for x := 0; x < 256; x++ {
				table[round][place][x] = constr.SubByte(byte(x) ^ roundKeys[round - 1][place])
			}
		}
	}

	// 10th T-Box
	for place := 0; place < 16; place++ {
		table[9][place] = [256]byte{}

		for x := 0; x < 256; x++ {
			table[9][place][x] = constr.SubByte(byte(x) ^ roundKeys[9][place]) ^ roundKeys[10][place]
		}
	}

	return
}

func (constr *Construction) Encrypt(block [16]byte) [16]byte {
	for i := 0; i < 9; i++ {
		block = constr.shiftRows(block)

		for j := 0; j < 16; j++ {
			block[j] = constr.TBox[i][j][block[j]]
		}

		block = constr.mixColumns(block)
	}

	block = constr.shiftRows(block)

	for j := 0; j < 16; j++ {
		block[j] = constr.TBox[9][j][block[j]]
	}

	return block
}

func (constr *Construction) shiftRows(block [16]byte) [16]byte {
	return [16]byte{
		block[0], block[5], block[10], block[15], block[4], block[9], block[14], block[3], block[8], block[13], block[2],
		block[7], block[12], block[1], block[6], block[11],
	}
}

func (constr *Construction) mixColumns(block [16]byte) (out [16]byte) {
	for i := 0; i < 4; i++ {
		copy(out[4*i:4*(i+1)], constr.mixColumn(block[4*i:4*(i+1)]))
	}

	return out
}

func (constr *Construction) mixColumn(slice []byte) (out []byte) {
	column := number.ArrayFieldElem{}
	for i := 0; i < 4; i++ {
		column = append(column, number.ByteFieldElem(slice[i]))
	}

	column = column.Mul(number.ArrayFieldElem{
		number.ByteFieldElem(0x02), number.ByteFieldElem(0x01),
		number.ByteFieldElem(0x01), number.ByteFieldElem(0x03),
	})

	for i := 0; i < 4; i++ {
		out = append(out, byte(column[i]))
	}

	return
}
