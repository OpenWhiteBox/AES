package bes

import (
	"github.com/OpenWhiteBox/primitives/gfmatrix"
	"github.com/OpenWhiteBox/primitives/number"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

func makeOneSubBytes() gfmatrix.Matrix {
	out := gfmatrix.Matrix{}

	subBytes := []number.ByteFieldElem{
		0x05, 0x09, 0xf9, 0x25, 0xf4, 0x01, 0xb5, 0x8f,
	}

	for row := 0; row < 8; row++ {
		out = append(out, gfmatrix.NewRow(8))

		for col := 0; col < 8; col++ {
			i := (col - row + 8) % 8
			out[row][col] = subBytes[i]
			subBytes[i] = subBytes[i].Mul(subBytes[i])
		}
	}

	return out
}

func makeSubBytes(n int) gfmatrix.Matrix {
	out := gfmatrix.Matrix{}
	subBytes := makeOneSubBytes()

	for row := 0; row < 8*n; row++ {
		out = append(out, gfmatrix.NewRow(8*n))

		base := 8 * (row / 8)
		for col := 0; col < 8; col++ {
			out[row][base+col] = subBytes[row%8][col]
		}
	}

	return out
}

func makeSubBytesConst(n int) gfmatrix.Row {
	out := make([]byte, n)
	for pos := 0; pos < n; pos++ {
		out[pos] = 0x63
	}

	return Expand(out)
}

var (
	subBytes = makeSubBytes(16)
	unSubBytes, _ = subBytes.Invert()
	subBytesConst = makeSubBytesConst(16)

	wordSubBytes = makeSubBytes(4)
	wordSubBytesConst = makeSubBytesConst(4)
)

func makeShiftRows() gfmatrix.Matrix {
	out := gfmatrix.Matrix{}
	for row := 0; row < 128; row++ {
		out = append(out, gfmatrix.NewRow(128))
	}

	for col := 0; col < 16; col++ {
		row := common.ShiftRows(col)

		for i := 0; i < 8; i++ {
			out[8*row+i][8*col+i] = 1
		}
	}

	return out
}

var(
	shiftRows = makeShiftRows()
	unShiftRows, _ = shiftRows.Invert()
)

func makeOneMixColumns() gfmatrix.Matrix {
	out := gfmatrix.Matrix{}
	for row := 0; row < 32; row++ {
		out = append(out, gfmatrix.NewRow(32))
	}

	a, b := number.ByteFieldElem(0x02), number.ByteFieldElem(0x03)

	for trans := 0; trans < 8; trans++ {
		out[0+trans][0+trans], out[0+trans][8+trans], out[0+trans][16+trans], out[0+trans][24+trans] = a, b, 1, 1
		out[8+trans][0+trans], out[8+trans][8+trans], out[8+trans][16+trans], out[8+trans][24+trans] = 1, a, b, 1
		out[16+trans][0+trans], out[16+trans][8+trans], out[16+trans][16+trans], out[16+trans][24+trans] = 1, 1, a, b
		out[24+trans][0+trans], out[24+trans][8+trans], out[24+trans][16+trans], out[24+trans][24+trans] = b, 1, 1, a

		a, b = a.Mul(a), b.Mul(b)
	}

	return out
}

func makeMixColumns() gfmatrix.Matrix {
	out := gfmatrix.Matrix{}
	mixColumns := makeOneMixColumns()

	for row := 0; row < 128; row++ {
		out = append(out, gfmatrix.NewRow(128))

		base := 32 * (row / 32)
		for col := 0; col < 32; col++ {
			out[row][base+col] = mixColumns[row%32][col]
		}
	}

	return out
}

var(
	mixColumns = makeMixColumns()
	unMixColumns, _ = mixColumns.Invert()
)

func makeRound() gfmatrix.Matrix {
	return mixColumns.Compose(shiftRows).Compose(subBytes)
}

func makeRoundConst() gfmatrix.Row {
	return mixColumns.Compose(shiftRows).Mul(makeSubBytesConst(16))
}

func makeLastRound() gfmatrix.Matrix {
	return shiftRows.Compose(subBytes)
}

var (
	round = makeRound()
	unRound, _ = round.Invert()

	roundConst = makeRoundConst()

	lastRound = makeLastRound()
	firstRound, _ = lastRound.Invert()
)
