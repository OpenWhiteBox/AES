package xiao

import (
	"github.com/OpenWhiteBox/AES/primitives/number"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

type Side int

const (
	Left Side = iota
	Right
)

type TBoxMixCol struct {
	TBoxes [2]common.TBox
	Side
}

func (tmc TBoxMixCol) Get(i [2]byte) (out [4]byte) {
	// Push input through the T-Boxes.
	k, l := number.ByteFieldElem(tmc.TBoxes[0].Get(i[0])), number.ByteFieldElem(tmc.TBoxes[1].Get(i[1]))

	// Calculate two slices of the MixColumns step at the same time.
	var a, b, c, d number.ByteFieldElem
	a = number.ByteFieldElem(0x02).Mul(k)
	b = number.ByteFieldElem(0x01).Mul(k)
	d = number.ByteFieldElem(0x03).Mul(k)

	a = number.ByteFieldElem(0x03).Mul(l).Add(a)
	c = number.ByteFieldElem(0x01).Mul(l)
	d = c.Add(d)
	c = c.Add(b)
	b = number.ByteFieldElem(0x02).Mul(l).Add(b)

	// Merge into one output and rotate according to what side of the word we're on.
	shift := 0
	if tmc.Side == Right {
		shift = 2
	}

	res := [4]byte{byte(a), byte(b), byte(c), byte(d)}
	copy(out[:], append(res[(4-shift):], res[0:(4-shift)]...))

	return
}

type TBox struct {
	TBoxes [2]common.TBox
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
