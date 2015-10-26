// A linear encoding is specified by an invertible binary matrix.
package encoding

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

func matrixMul(m *matrix.Matrix, dst, src []byte) {
	res := m.Mul(matrix.Row(src[:]))
	copy(dst, res)
}

type ByteLinear struct {
	Forwards, Backwards matrix.Matrix
}

func (bl ByteLinear) Encode(i byte) byte { return bl.Forwards.Mul(matrix.Row{i})[0] }
func (bl ByteLinear) Decode(i byte) byte { return bl.Backwards.Mul(matrix.Row{i})[0] }

type ByteAffine struct {
	Linear   ByteLinear
	Constant byte
}

func (ba ByteAffine) Encode(i byte) byte { return ba.Linear.Encode(i) ^ ba.Constant }
func (ba ByteAffine) Decode(i byte) byte { return ba.Linear.Decode(i ^ ba.Constant) }

type DoubleLinear struct {
	Forwards, Backwards matrix.Matrix
}

func (dl DoubleLinear) Encode(i [2]byte) (out [2]byte) {
	matrixMul(&dl.Forwards, out[:], i[:])
	return
}

func (dl DoubleLinear) Decode(i [2]byte) (out [2]byte) {
	matrixMul(&dl.Backwards, out[:], i[:])
	return
}

type WordLinear struct {
	Forwards, Backwards matrix.Matrix
}

func (wl WordLinear) Encode(i [4]byte) (out [4]byte) {
	matrixMul(&wl.Forwards, out[:], i[:])
	return
}

func (wl WordLinear) Decode(i [4]byte) (out [4]byte) {
	matrixMul(&wl.Backwards, out[:], i[:])
	return
}

type BlockLinear struct {
	Forwards, Backwards matrix.Matrix
}

func (bl BlockLinear) Encode(i [16]byte) (out [16]byte) {
	matrixMul(&bl.Forwards, out[:], i[:])
	return
}

func (bl BlockLinear) Decode(i [16]byte) (out [16]byte) {
	matrixMul(&bl.Backwards, out[:], i[:])
	return
}
