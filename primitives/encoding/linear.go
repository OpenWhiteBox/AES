// A linear encoding is specified by an invertible binary matrix.
package encoding

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

type ByteLinear matrix.Matrix

func (bl ByteLinear) Encode(i byte) byte { return matrix.Matrix(bl).Mul(matrix.Row{i})[0] }
func (bl ByteLinear) Decode(i byte) byte {
	inv, ok := matrix.Matrix(bl).Invert()

	if !ok {
		panic("Matrix wasn't invertible!")
	}

	return inv.Mul(matrix.Row{i})[0]
}

type WordLinear matrix.Matrix

func (wl WordLinear) Encode(i [4]byte) (out [4]byte) {
	res := matrix.Matrix(wl).Mul(matrix.Row(i[:]))
	copy(out[:], res)

	return
}

func (wl WordLinear) Decode(i [4]byte) (out [4]byte) {
	inv, ok := matrix.Matrix(wl).Invert() // Performance bottleneck.

	if !ok {
		panic("Matrix wasn't invertible!")
	}

	res := inv.Mul(matrix.Row(i[:]))
	copy(out[:], res)

	return
}

type BlockLinear matrix.Matrix

func (bl BlockLinear) Encode(i [16]byte) (out [16]byte) {
	res := matrix.Matrix(bl).Mul(matrix.Row(i[:]))
	copy(out[:], res)

	return
}

func (bl BlockLinear) Decode(i [16]byte) (out [16]byte) {
	inv, ok := matrix.Matrix(bl).Invert()

	if !ok {
		panic("Matrix wasn't invertible!")
	}

	res := inv.Mul(matrix.Row(i[:]))
	copy(out[:], res)

	return
}
