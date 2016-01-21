package encoding

import (
	"crypto/rand"

	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

// EquivalentBytes returns true if two Byte encodings are identical and false if not.
func EquivalentBytes(a, b Byte) bool {
	for x := 0; x < 256; x++ {
		if a.Encode(byte(x)) != b.Encode(byte(x)) {
			return false
		}
	}

	return true
}

// DecomposeByteLinear decomposes an opaque Byte encoding into a ByteLinear encoding.
func DecomposeByteLinear(in Byte) ByteLinear {
	m := matrix.Matrix{}
	for i := uint(0); i < 8; i++ {
		m = append(m, matrix.Row{in.Encode(byte(1 << i))})
	}

	return NewByteLinear(m.Transpose())
}

// DecomposeByteAffine decomposes an opaque Byte encoding into a ByteAffine encoding.
func DecomposeByteAffine(in Byte) ByteAffine {
	c := ByteAdditive(in.Encode(0))

	return ByteAffine{
		ByteLinear:   DecomposeByteLinear(ComposedBytes{in, c}),
		ByteAdditive: c,
	}
}

// ProbablyEquivalentDoubles returns true if two Double encodings are probably equivalent and false if they're
// definitely not.
func ProbablyEquivalentDoubles(a, b Double) bool {
	for i := 0; i < 20; i++ {
		in := [2]byte{}
		rand.Read(in[:])

		x, y := a.Encode(in), b.Encode(in)
		if x != y {
			return false
		}
	}

	return true
}

// DecomposeDoubleLinear decomposes an opaque Double encoding into a DoubleLinear encoding.
func DecomposeDoubleLinear(in Double) DoubleLinear {
	m := matrix.Matrix{}
	for i := 0; i < 2; i++ {
		for j := uint(0); j < 8; j++ {
			x := [2]byte{}
			x[i] = 1 << j
			x = in.Encode(x)

			m = append(m, matrix.Row(x[:]))
		}
	}

	return NewDoubleLinear(m.Transpose())
}

// DecomposeDoubleAffine decomposes an opaque Double encoding into a DoubleAffine encoding.
func DecomposeDoubleAffine(in Double) DoubleAffine {
	c := DoubleAdditive(in.Encode([2]byte{}))

	return DoubleAffine{
		DoubleLinear:   DecomposeDoubleLinear(ComposedDoubles{in, c}),
		DoubleAdditive: c,
	}
}

// ProbablyEquivalentWords returns true if two Word encodings are probably equivalent and false if they're definitely
// not.
func ProbablyEquivalentWords(a, b Word) bool {
	for i := 0; i < 20; i++ {
		in := [4]byte{}
		rand.Read(in[:])

		x, y := a.Encode(in), b.Encode(in)
		if x != y {
			return false
		}
	}

	return true
}

// DecomposeWordLinear decomposes an opaque Word encoding into a WordLinear encoding.
func DecomposeWordLinear(in Word) WordLinear {
	m := matrix.Matrix{}
	for i := 0; i < 4; i++ {
		for j := uint(0); j < 8; j++ {
			x := [4]byte{}
			x[i] = 1 << j
			x = in.Encode(x)

			m = append(m, matrix.Row(x[:]))
		}
	}

	return NewWordLinear(m.Transpose())
}

// DecomposeWordAffine decomposes an opaque Word encoding into a WordAffine encoding.
func DecomposeWordAffine(in Word) WordAffine {
	c := WordAdditive(in.Encode([4]byte{}))

	return WordAffine{
		WordLinear:   DecomposeWordLinear(ComposedWords{in, c}),
		WordAdditive: c,
	}
}

// ProbablyEquivalentBlocks returns true if two Block encodings are probably equivalent and false if they're definitely
// not.
func ProbablyEquivalentBlocks(a, b Block) bool {
	for i := 0; i < 20; i++ {
		in := [16]byte{}
		rand.Read(in[:])

		x, y := a.Encode(in), b.Encode(in)
		if x != y {
			return false
		}
	}

	return true
}

// DecomposeBlockLinear decomposes an opaque Block encoding into a BlockLinear encoding.
func DecomposeBlockLinear(in Block) BlockLinear {
	m := matrix.Matrix{}
	for i := 0; i < 16; i++ {
		for j := uint(0); j < 8; j++ {
			x := [16]byte{}
			x[i] = 1 << j
			x = in.Encode(x)

			m = append(m, matrix.Row(x[:]))
		}
	}

	return NewBlockLinear(m.Transpose())
}

// DecomposeBlockAffine decomposes an opaque Block encoding into a BlockAffine encoding.
func DecomposeBlockAffine(in Block) BlockAffine {
	c := BlockAdditive(in.Encode([16]byte{}))

	return BlockAffine{
		BlockLinear:   DecomposeBlockLinear(ComposedBlocks{in, c}),
		BlockAdditive: c,
	}
}
