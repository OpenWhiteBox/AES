package xiao

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/number"
	"github.com/OpenWhiteBox/AES/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

type Side int

const (
	Left Side = iota
	Right
)

func SideFromPos(pos int) Side {
	if pos%4 < 2 {
		return Left
	} else {
		return Right
	}
}

type TBoxMixCol struct {
	TBoxes [2]table.Byte
	MixCol func(byte, byte) [4]byte
	Side
}

func (tmc TBoxMixCol) Get(i [2]byte) (out [4]byte) {
	// Push input through the T-Boxes.
	k, l := tmc.TBoxes[0].Get(i[0]), tmc.TBoxes[1].Get(i[1])

	// Calculate two slices of the MixColumns step at the same time.
	res := tmc.MixCol(k, l)

	// Merge into one output and rotate according to what side of the word we're on.
	shift := 0
	if tmc.Side == Right {
		shift = 2
	}

	copy(out[:], append(res[(4-shift):], res[0:(4-shift)]...))

	return
}

type TBox struct {
	TBoxes [2]table.Byte
	Side
}

func (t TBox) Get(i [2]byte) (out [4]byte) {
	k, l := t.TBoxes[0].Get(i[0]), t.TBoxes[1].Get(i[1])

	if t.Side == Left {
		out[0], out[1] = k, l
	} else {
		out[2], out[3] = k, l
	}

	return
}

func MixColumns(i, j byte) [4]byte {
	k, l := number.ByteFieldElem(i), number.ByteFieldElem(j)

	var a, b, c, d number.ByteFieldElem

	a = number.ByteFieldElem(0x02).Mul(k)
	b = number.ByteFieldElem(0x01).Mul(k)
	d = number.ByteFieldElem(0x03).Mul(k)

	a = number.ByteFieldElem(0x03).Mul(l).Add(a)
	c = number.ByteFieldElem(0x01).Mul(l)
	d = c.Add(d)
	c = c.Add(b)
	b = number.ByteFieldElem(0x02).Mul(l).Add(b)

	return [4]byte{byte(a), byte(b), byte(c), byte(d)}
}

func UnMixColumns(i, j byte) [4]byte {
	k, l := number.ByteFieldElem(i), number.ByteFieldElem(j)

	var a, b, c, d number.ByteFieldElem

	a = number.ByteFieldElem(0x0e).Mul(k)
	b = number.ByteFieldElem(0x09).Mul(k)
	c = number.ByteFieldElem(0x0d).Mul(k)
	d = number.ByteFieldElem(0x0b).Mul(k)

	a = number.ByteFieldElem(0x0b).Mul(l).Add(a)
	b = number.ByteFieldElem(0x0e).Mul(l).Add(b)
	c = number.ByteFieldElem(0x09).Mul(l).Add(c)
	d = number.ByteFieldElem(0x0d).Mul(l).Add(d)

	return [4]byte{byte(a), byte(b), byte(c), byte(d)}
}

func MaskSwap(rs *common.RandomSource, size, round int) (out matrix.Matrix) {
	out = matrix.GenerateEmpty(128)

	for row := 0; row < 128; row += size {
		col := row / 8
		m := common.MixingBijection(rs, size, round, row/size)

		for subRow := 0; subRow < size; subRow++ {
			copy(out[row+subRow][col:], m[subRow])
		}
	}

	return
}
