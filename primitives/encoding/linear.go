package encoding

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

func matrixMul(m *matrix.Matrix, dst, src []byte) {
	res := m.Mul(matrix.Row(src[:]))
	copy(dst, res)
}

// ByteLinear implements the Byte interface over a 16x16 linear transformation.
type ByteLinear struct {
	// Forwards is the matrix to multiply by in the forwards (encoding) direction.
	Forwards matrix.Matrix
	// Backwards is the matrix to multiply by in the backwards (decoding) direction. It should be the inverse of Forwards.
	Backwards matrix.Matrix
}

func (bl ByteLinear) Encode(i byte) byte { return bl.Forwards.Mul(matrix.Row{i})[0] }
func (bl ByteLinear) Decode(i byte) byte { return bl.Backwards.Mul(matrix.Row{i})[0] }

// ByteAffine implements the Byte interface over an affine transformation.
type ByteAffine struct {
	// Linear is the linear part of the affine transformation.
	Linear ByteLinear
	// Constant will be XORed with the linear part.
	Constant byte
}

func (ba ByteAffine) Encode(i byte) byte { return ba.Linear.Encode(i) ^ ba.Constant }
func (ba ByteAffine) Decode(i byte) byte { return ba.Linear.Decode(i ^ ba.Constant) }

// DoubleLinear implements the Double interface over a 16x16 linear transformation.
type DoubleLinear struct {
	// Forwards is the matrix to multiply by in the forwards (encoding) direction.
	Forwards matrix.Matrix
	// Backwards is the matrix to multiply by in the backwards (decoding) direction. It should be the inverse of Forwards.
	Backwards matrix.Matrix
}

func (dl DoubleLinear) Encode(i [2]byte) (out [2]byte) {
	matrixMul(&dl.Forwards, out[:], i[:])
	return
}

func (dl DoubleLinear) Decode(i [2]byte) (out [2]byte) {
	matrixMul(&dl.Backwards, out[:], i[:])
	return
}

// WordLinear implements the Word interface over a 32x32 linear transformation.
type WordLinear struct {
	// Forwards is the matrix to multiply by in the forwards (encoding) direction.
	Forwards matrix.Matrix
	// Backwards is the matrix to multiply by in the backwards (decoding) direction. It should be the inverse of Forwards.
	Backwards matrix.Matrix
}

func (wl WordLinear) Encode(i [4]byte) (out [4]byte) {
	matrixMul(&wl.Forwards, out[:], i[:])
	return
}

func (wl WordLinear) Decode(i [4]byte) (out [4]byte) {
	matrixMul(&wl.Backwards, out[:], i[:])
	return
}

// BlockLinear implements the Block interface over a 128x128 linear transformation.
type BlockLinear struct {
	// Forwards is the matrix to multiply by in the forwards (encoding) direction.
	Forwards matrix.Matrix
	// Backwards is the matrix to multiply by in the backwards (decoding) direction. It should be the inverse of Forwards.
	Backwards matrix.Matrix
}

func (bl BlockLinear) Encode(i [16]byte) (out [16]byte) {
	matrixMul(&bl.Forwards, out[:], i[:])
	return
}

func (bl BlockLinear) Decode(i [16]byte) (out [16]byte) {
	matrixMul(&bl.Backwards, out[:], i[:])
	return
}
