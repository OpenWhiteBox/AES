// An encoding is a bijective map between primitive values (nibble<->nibble, byte<->byte, ...).
package encoding

import (
	"../matrix"
)

type Nibble interface {
	Encode(i byte) byte
	Decode(i byte) byte
}

type Byte interface {
	Encode(i byte) byte
	Decode(i byte) byte
}

type Word interface {
	Encode(i uint32) uint32
	Decode(i uint32) uint32
}

// The IdentityByte encoding is also used as the IdentityNibble encoding.
type IdentityByte struct{}

func (ib IdentityByte) Encode(i byte) byte { return i }
func (ib IdentityByte) Decode(i byte) byte { return i }

type IdentityWord struct{}

func (iw IdentityWord) Encode(i uint32) uint32 { return i }
func (iw IdentityWord) Decode(i uint32) uint32 { return i }

type ComposedBytes []Byte

func (cb ComposedBytes) Encode(i byte) byte {
	for j := 0; j < len(cb); j++ {
		i = cb[j].Encode(i)
	}

	return i
}

func (cb ComposedBytes) Decode(i byte) byte {
	for j := len(cb) - 1; j >= 0; j-- {
		i = cb[j].Decode(i)
	}

	return i
}

type ComposedWords []Word

func (cw ComposedWords) Encode(i uint32) uint32 {
	for j := 0; j < len(cw); j++ {
		i = cw[j].Encode(i)
	}

	return i
}

func (cw ComposedWords) Decode(i uint32) uint32 {
	for j := len(cw) - 1; j >= 0; j-- {
		i = cw[j].Decode(i)
	}

	return i
}

// A concatenated encoding is a bijection of a large primitive built by concatenating smaller encodings.
// In the example, f(x_1 || x_2) = f_1(x_1) || f_2(x_2), f is a concatenated encoding built from f_1 and f_2.
type ConcatenatedByte struct {
	Left, Right Nibble
}

func (cb ConcatenatedByte) Encode(i byte) byte {
	return (cb.Left.Encode(i>>4) << 4) | cb.Right.Encode(i&0xf)
}

func (cb ConcatenatedByte) Decode(i byte) byte {
	return (cb.Left.Decode(i>>4) << 4) | cb.Right.Decode(i&0xf)
}

type ConcatenatedWord struct {
	A, B, C, D Byte
}

func (cw ConcatenatedWord) Encode(i uint32) uint32 {
	return uint32(cw.A.Encode(byte(i>>24)))<<24 |
		uint32(cw.B.Encode(byte(i>>16)))<<16 |
		uint32(cw.C.Encode(byte(i>>8)))<<8 |
		uint32(cw.D.Encode(byte(i)))
}

func (cw ConcatenatedWord) Decode(i uint32) uint32 {
	return uint32(cw.A.Decode(byte(i>>24)))<<24 |
		uint32(cw.B.Decode(byte(i>>16)))<<16 |
		uint32(cw.C.Decode(byte(i>>8)))<<8 |
		uint32(cw.D.Decode(byte(i)))
}

// A linear encoding is specified by an invertible binary matrix.
type ByteLinear matrix.ByteMatrix

func (bl ByteLinear) Encode(i byte) byte { return matrix.ByteMatrix(bl).Mul(i) }
func (bl ByteLinear) Decode(i byte) byte {
	inv, ok := matrix.ByteMatrix(bl).Invert()

	if !ok {
		panic("Matrix wasn't invertible!")
	}

	return inv.Mul(i)
}

type WordLinear matrix.WordMatrix

func (wl WordLinear) Encode(i uint32) uint32 { return matrix.WordMatrix(wl).Mul(i) }
func (wl WordLinear) Decode(i uint32) uint32 {
	inv, ok := matrix.WordMatrix(wl).Invert()

	if !ok {
		panic("Matrix wasn't invertible!")
	}

	return inv.Mul(i)
}
