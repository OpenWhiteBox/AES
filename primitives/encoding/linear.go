package encoding

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

func matrixMul(m *matrix.Matrix, dst, src []byte) {
	res := m.Mul(matrix.Row(src[:]))
	copy(dst, res)
}

// ByteAdditive implements the Byte interface over XORing with a fixed value.
type ByteAdditive byte

func (ba ByteAdditive) code(in byte) byte   { return in ^ byte(ba) }
func (ba ByteAdditive) Encode(in byte) byte { return ba.code(in) }
func (ba ByteAdditive) Decode(in byte) byte { return ba.code(in) }

// ByteLinear implements the Byte interface over an 8x8 linear transformation.
type ByteLinear struct {
	// Forwards is the matrix to multiply by in the forwards (encoding) direction.
	Forwards matrix.Matrix
	// Backwards is the matrix to multiply by in the backwards (decoding) direction. It should be the inverse of Forwards.
	Backwards matrix.Matrix
}

// NewByteLinear constructs a new ByteLinear encoding from a given matrix.
func NewByteLinear(forwards matrix.Matrix) ByteLinear {
	backwards, ok := forwards.Invert()
	if !ok {
		panic("Non-invertible matrix given to NewByteLinear!")
	}

	return ByteLinear{
		Forwards:  forwards,
		Backwards: backwards,
	}
}

func (bl ByteLinear) Encode(in byte) byte { return bl.Forwards.Mul(matrix.Row{in})[0] }
func (bl ByteLinear) Decode(in byte) byte { return bl.Backwards.Mul(matrix.Row{in})[0] }

// ByteAffine implements the Byte interface over an affine transformation (a linear transformation composed with an
// additive one).
type ByteAffine struct {
	ByteLinear
	ByteAdditive
}

// NewByteAffine constructs a new ByteAffine encoding from a matrix and a constant.
func NewByteAffine(forwards matrix.Matrix, constant byte) ByteAffine {
	return ByteAffine{
		ByteLinear:   NewByteLinear(forwards),
		ByteAdditive: ByteAdditive(constant),
	}
}

func (ba ByteAffine) Encode(in byte) byte { return ba.ByteAdditive.Encode(ba.ByteLinear.Encode(in)) }
func (ba ByteAffine) Decode(in byte) byte { return ba.ByteLinear.Decode(ba.ByteAdditive.Decode(in)) }

// DoubleAdditive implements the Double interface over XORing with a fixed value.
type DoubleAdditive [2]byte

func (da DoubleAdditive) code(in [2]byte) (out [2]byte) {
	XOR(out[:], in[:], da[:])
	return
}

func (da DoubleAdditive) Encode(in [2]byte) [2]byte { return da.code(in) }
func (da DoubleAdditive) Decode(in [2]byte) [2]byte { return da.code(in) }

// DoubleLinear implements the Double interface over a 16x16 linear transformation.
type DoubleLinear struct {
	// Forwards is the matrix to multiply by in the forwards (encoding) direction.
	Forwards matrix.Matrix
	// Backwards is the matrix to multiply by in the backwards (decoding) direction. It should be the inverse of Forwards.
	Backwards matrix.Matrix
}

// NewDoubleLinear constructs a new DoubleLinear encoding from a given matrix.
func NewDoubleLinear(forwards matrix.Matrix) DoubleLinear {
	backwards, ok := forwards.Invert()
	if !ok {
		panic("Non-invertible matrix given to NewDoubleLinear!")
	}

	return DoubleLinear{
		Forwards:  forwards,
		Backwards: backwards,
	}
}

func (dl DoubleLinear) Encode(in [2]byte) (out [2]byte) {
	matrixMul(&dl.Forwards, out[:], in[:])
	return
}

func (dl DoubleLinear) Decode(in [2]byte) (out [2]byte) {
	matrixMul(&dl.Backwards, out[:], in[:])
	return
}

// DoubleAffine implements the Double interface over an affine transformation (a linear transformation composed with an
// additive one).
type DoubleAffine struct {
	DoubleLinear
	DoubleAdditive
}

// NewDoubleAffine constructs a new DoubleAffine encoding from a matrix and a constant.
func NewDoubleAffine(forwards matrix.Matrix, constant [2]byte) DoubleAffine {
	return DoubleAffine{
		DoubleLinear:   NewDoubleLinear(forwards),
		DoubleAdditive: DoubleAdditive(constant),
	}
}

func (da DoubleAffine) Encode(in [2]byte) [2]byte {
	return da.DoubleAdditive.Encode(da.DoubleLinear.Encode(in))
}
func (da DoubleAffine) Decode(in [2]byte) [2]byte {
	return da.DoubleLinear.Decode(da.DoubleAdditive.Decode(in))
}

// WordAdditive implements the Word interface over XORing with a fixed value.
type WordAdditive [4]byte

func (wa WordAdditive) code(in [4]byte) (out [4]byte) {
	XOR(out[:], in[:], wa[:])
	return
}

func (wa WordAdditive) Encode(in [4]byte) [4]byte { return wa.code(in) }
func (wa WordAdditive) Decode(in [4]byte) [4]byte { return wa.code(in) }

// WordLinear implements the Word interface over a 32x32 linear transformation.
type WordLinear struct {
	// Forwards is the matrix to multiply by in the forwards (encoding) direction.
	Forwards matrix.Matrix
	// Backwards is the matrix to multiply by in the backwards (decoding) direction. It should be the inverse of Forwards.
	Backwards matrix.Matrix
}

// NewWordLinear constructs a new WordLinear encoding from a given matrix.
func NewWordLinear(forwards matrix.Matrix) WordLinear {
	backwards, ok := forwards.Invert()
	if !ok {
		panic("Non-invertible matrix given to NewWordLinear!")
	}

	return WordLinear{
		Forwards:  forwards,
		Backwards: backwards,
	}
}

func (wl WordLinear) Encode(in [4]byte) (out [4]byte) {
	matrixMul(&wl.Forwards, out[:], in[:])
	return
}

func (wl WordLinear) Decode(in [4]byte) (out [4]byte) {
	matrixMul(&wl.Backwards, out[:], in[:])
	return
}

// WordAffine implements the Word interface over an affine transformation (a linear transformation composed with an
// additive one).
type WordAffine struct {
	WordLinear
	WordAdditive
}

// NewWordAffine constructs a new WordAffine encoding from a matrix and a constant.
func NewWordAffine(forwards matrix.Matrix, constant [4]byte) WordAffine {
	return WordAffine{
		WordLinear:   NewWordLinear(forwards),
		WordAdditive: WordAdditive(constant),
	}
}

func (wa WordAffine) Encode(in [4]byte) [4]byte {
	return wa.WordAdditive.Encode(wa.WordLinear.Encode(in))
}
func (wa WordAffine) Decode(in [4]byte) [4]byte {
	return wa.WordLinear.Decode(wa.WordAdditive.Decode(in))
}

// BlockAdditive implements the Block interface over XORing with a fixed value.
type BlockAdditive [16]byte

func (ba BlockAdditive) code(in [16]byte) (out [16]byte) {
	XOR(out[:], in[:], ba[:])
	return
}

func (ba BlockAdditive) Encode(in [16]byte) [16]byte { return ba.code(in) }
func (ba BlockAdditive) Decode(in [16]byte) [16]byte { return ba.code(in) }

// BlockLinear implements the Block interface over a 128x128 linear transformation.
type BlockLinear struct {
	// Forwards is the matrix to multiply by in the forwards (encoding) direction.
	Forwards matrix.Matrix
	// Backwards is the matrix to multiply by in the backwards (decoding) direction. It should be the inverse of Forwards.
	Backwards matrix.Matrix
}

// NewBlockLinear constructs a new BlockLinear encoding from a given matrix.
func NewBlockLinear(forwards matrix.Matrix) BlockLinear {
	backwards, ok := forwards.Invert()
	if !ok {
		panic("Non-invertible matrix given to NewBlockLinear!")
	}

	return BlockLinear{
		Forwards:  forwards,
		Backwards: backwards,
	}
}

func (bl BlockLinear) Encode(in [16]byte) (out [16]byte) {
	matrixMul(&bl.Forwards, out[:], in[:])
	return
}

func (bl BlockLinear) Decode(in [16]byte) (out [16]byte) {
	matrixMul(&bl.Backwards, out[:], in[:])
	return
}

// BlockAffine implements the Block interface over an affine transformation (a linear transformation composed with an
// additive one).
type BlockAffine struct {
	BlockLinear
	BlockAdditive
}

// NewBlockAffine constructs a new BlockAffine encoding from a matrix and a constant.
func NewBlockAffine(forwards matrix.Matrix, constant [16]byte) BlockAffine {
	return BlockAffine{
		BlockLinear:   NewBlockLinear(forwards),
		BlockAdditive: BlockAdditive(constant),
	}
}

func (ba BlockAffine) Encode(in [16]byte) [16]byte {
	return ba.BlockAdditive.Encode(ba.BlockLinear.Encode(in))
}
func (ba BlockAffine) Decode(in [16]byte) [16]byte {
	return ba.BlockLinear.Decode(ba.BlockAdditive.Decode(in))
}
