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

type Construction struct {
	TBox [10][16]table.ByteTable
}

func GenerateTables(key [16]byte) (table [10][16]table.ByteTable) {
	constr := saes.Construction{key}
	roundKeys := constr.StretchedKey()

	// Apply ShiftRows to round keys 0 to 9.
	for k := 0; k < 10; k++ {
		roundKeys[k] = constr.ShiftRows(roundKeys[k])
	}

	// Build T-Boxes 1 to 9
	for round := 0; round < 9; round++ {
		for place := 0; place < 16; place++ {
			table[round][place] = TBox{constr, roundKeys[round][place], 0}
		}
	}

	// 10th T-Box
	for place := 0; place < 16; place++ {
		table[9][place] = TBox{constr, roundKeys[9][place], roundKeys[10][place]}
	}

	return
}

func (constr *Construction) Encrypt(block [16]byte) [16]byte {
	for i := 0; i < 9; i++ {
		block = constr.shiftRows(block)

		for j := 0; j < 16; j++ {
			block[j] = constr.TBox[i][j].Get(block[j])
		}

		block = constr.mixColumns(block)
	}

	block = constr.shiftRows(block)

	for j := 0; j < 16; j++ {
		block[j] = constr.TBox[9][j].Get(block[j])
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

func (constr *Construction) mixColumn(slice []byte) []byte {
	column := number.ArrayFieldElem{}
	for i := 0; i < 4; i++ {
		column = append(column, number.ByteFieldElem(slice[i]))
	}

	column = column.Mul(number.ArrayFieldElem{
		number.ByteFieldElem(0x02), number.ByteFieldElem(0x01),
		number.ByteFieldElem(0x01), number.ByteFieldElem(0x03),
	})

	out := make([]byte, 4)
	for i := 0; i < len(column); i++ {
		out[i] = byte(column[i])
	}

	return out
}
