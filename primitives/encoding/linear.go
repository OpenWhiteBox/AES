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

func (wl WordLinear) Encode(i uint32) uint32 {
	out := matrix.Matrix(wl).Mul(matrix.Row{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
	return uint32(out[0])<<24 | uint32(out[1])<<16 | uint32(out[2])<<8 | uint32(out[3])
}

func (wl WordLinear) Decode(i uint32) uint32 {
	inv, ok := matrix.Matrix(wl).Invert() // Performance bottleneck.

	if !ok {
		panic("Matrix wasn't invertible!")
	}

	out := inv.Mul(matrix.Row{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)})
	return uint32(out[0])<<24 | uint32(out[1])<<16 | uint32(out[2])<<8 | uint32(out[3])
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
