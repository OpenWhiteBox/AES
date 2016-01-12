// Package encoding defines interfaces to be implemented by bijective, invertible functions. Implementing a common
// interface over the building blocks of a construction or cryptanalysis gives a simple way to compose, concatenate, and
// invert them.
package encoding

// Nibble is the same interface as Byte. A function implementing Nibble shouldn't accept inputs or give outputs over 16.
type Nibble interface {
	Encode(i byte) byte // Encode(i nibble) nibble
	Decode(i byte) byte // Decode(i nibble) nibble
}

type Byte interface {
	Encode(i byte) byte
	Decode(i byte) byte
}

type Double interface {
	Encode(i [2]byte) [2]byte
	Decode(i [2]byte) [2]byte
}

type Word interface {
	Encode(i [4]byte) [4]byte
	Decode(i [4]byte) [4]byte
}

type Block interface {
	Encode(i [16]byte) [16]byte
	Decode(i [16]byte) [16]byte
}

// IdentityByte is the identity operation on bytes. It is used in place of an IdentityNibble encoding.
type IdentityByte struct{}

func (ib IdentityByte) Encode(i byte) byte { return i }
func (ib IdentityByte) Decode(i byte) byte { return i }

// IdentityDouble is the identity operation on doubles.
type IdentityDouble struct{}

func (id IdentityDouble) Encode(i [2]byte) [2]byte { return i }
func (id IdentityDouble) Decode(i [2]byte) [2]byte { return i }

// IdentityWord is the identity operation on words.
type IdentityWord struct{}

func (iw IdentityWord) Encode(i [4]byte) (out [4]byte) {
	copy(out[:], i[:])
	return
}

func (iw IdentityWord) Decode(i [4]byte) (out [4]byte) {
	copy(out[:], i[:])
	return
}

// IdentityBlock is the identity operation on blocks.
type IdentityBlock struct{}

func (ib IdentityBlock) Encode(i [16]byte) (out [16]byte) {
	copy(out[:], i[:])
	return
}

func (ib IdentityBlock) Decode(i [16]byte) (out [16]byte) {
	copy(out[:], i[:])
	return
}

// InverseByte swaps the Encode and Decode methods of a Byte encoding.
type InverseByte struct{ Byte }

func (ib InverseByte) Encode(i byte) byte { return ib.Byte.Decode(i) }
func (ib InverseByte) Decode(i byte) byte { return ib.Byte.Encode(i) }

// InverseDouble swaps the Encode and Decode methods of a Double encoding.
type InverseDouble struct{ Double }

func (id InverseDouble) Encode(i [2]byte) [2]byte { return id.Double.Decode(i) }
func (id InverseDouble) Decode(i [2]byte) [2]byte { return id.Double.Encode(i) }

// InverseWord swaps the Encode and Decode methods of a Word encoding.
type InverseWord struct{ Word }

func (iw InverseWord) Encode(i [4]byte) [4]byte { return iw.Word.Decode(i) }
func (iw InverseWord) Decode(i [4]byte) [4]byte { return iw.Word.Encode(i) }

// InverseBlock swaps the Encode and Decode methods of a Block encoding.
type InverseBlock struct{ Block }

func (ib InverseBlock) Encode(i [16]byte) [16]byte { return ib.Block.Decode(i) }
func (ib InverseBlock) Decode(i [16]byte) [16]byte { return ib.Block.Encode(i) }

// ComposedBytes converts an array of Byte encodings into one by chaining them. Functions are chained in REVERSE
// order than they would be in function composition notation.
//
// Example:
//   f := ComposedBytes([]Byte{A, B, C}) // f(x) = C(B(A(x)))
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

// ComposedDoubles converts an array of Double encodings into one by chaining them. See ComposedBytes.
type ComposedDoubles []Double

func (cd ComposedDoubles) Encode(i [2]byte) (out [2]byte) {
	res := cd[0].Encode(i)
	copy(out[:], res[:])

	for j := 1; j < len(cd); j++ {
		res = cd[j].Encode(out)
		copy(out[:], res[:])
	}

	return
}

func (cd ComposedDoubles) Decode(i [2]byte) (out [2]byte) {
	res := cd[len(cd)-1].Decode(i)
	copy(out[:], res[:])

	for j := len(cd) - 2; j >= 0; j-- {
		res = cd[j].Decode(out)
		copy(out[:], res[:])
	}

	return
}

// ComposedWords converts an array of Word encodings into one by chaining them. See ComposedBytes.
type ComposedWords []Word

func (cw ComposedWords) Encode(i [4]byte) (out [4]byte) {
	res := cw[0].Encode(i)
	copy(out[:], res[:])

	for j := 1; j < len(cw); j++ {
		res = cw[j].Encode(out)
		copy(out[:], res[:])
	}

	return
}

func (cw ComposedWords) Decode(i [4]byte) (out [4]byte) {
	res := cw[len(cw)-1].Decode(i)
	copy(out[:], res[:])

	for j := len(cw) - 2; j >= 0; j-- {
		res = cw[j].Decode(out)
		copy(out[:], res[:])
	}

	return
}

// ComposedBlocks converts an array of Block encodings into one by chaining them. See ComposedBytes.
type ComposedBlocks []Block

func (cb ComposedBlocks) Encode(i [16]byte) (out [16]byte) {
	res := cb[0].Encode(i)
	copy(out[:], res[:])

	for j := 1; j < len(cb); j++ {
		res = cb[j].Encode(out)
		copy(out[:], res[:])
	}

	return
}

func (cb ComposedBlocks) Decode(i [16]byte) (out [16]byte) {
	res := cb[len(cb)-1].Decode(i)
	copy(out[:], res[:])

	for j := len(cb) - 2; j >= 0; j-- {
		res = cb[j].Decode(out)
		copy(out[:], res[:])
	}

	return
}

// ConcatenatedByte builds a Byte encoding by concatenating two Nibble encodings. The Nibble encoding in position 0 is
// applied to the upper half of the byte and the one in position 1 is applied to the lower half.
type ConcatenatedByte [2]Nibble

func (cb ConcatenatedByte) Encode(i byte) byte {
	return (cb[0].Encode(i>>4) << 4) | cb[1].Encode(i&0xf)
}

func (cb ConcatenatedByte) Decode(i byte) byte {
	return (cb[0].Decode(i>>4) << 4) | cb[1].Decode(i&0xf)
}

// ConcatenatedDouble builds a Double encoding by Concatenating two Byte encodings. The Byte encoding in position i is
// the one applied to position i of the input.
type ConcatenatedDouble [2]Byte

func (cd ConcatenatedDouble) Encode(i [2]byte) [2]byte {
	return [2]byte{cd[0].Encode(i[0]), cd[1].Encode(i[1])}
}
func (cd ConcatenatedDouble) Decode(i [2]byte) [2]byte {
	return [2]byte{cd[0].Decode(i[0]), cd[1].Decode(i[1])}
}

// ConcatenatedWord builds a Word encoding by concatenating four Byte encodings. The Byte encoding in position i is the
// one applied to position i of the input.
type ConcatenatedWord [4]Byte

func (cw ConcatenatedWord) Encode(i [4]byte) [4]byte {
	return [4]byte{cw[0].Encode(i[0]), cw[1].Encode(i[1]), cw[2].Encode(i[2]), cw[3].Encode(i[3])}
}

func (cw ConcatenatedWord) Decode(i [4]byte) [4]byte {
	return [4]byte{cw[0].Decode(i[0]), cw[1].Decode(i[1]), cw[2].Decode(i[2]), cw[3].Decode(i[3])}
}

// ConcatenatedBlock builds a Block encoding by concatenating sixteen Byte encodings. The Byte encoding in position i is
// the one applied to position i of the input.
type ConcatenatedBlock [16]Byte

func (cb ConcatenatedBlock) Encode(i [16]byte) (out [16]byte) {
	for j := 0; j < 16; j++ {
		out[j] = cb[j].Encode(i[j])
	}

	return
}

func (cb ConcatenatedBlock) Decode(i [16]byte) (out [16]byte) {
	for j := 0; j < 16; j++ {
		out[j] = cb[j].Decode(i[j])
	}

	return
}
