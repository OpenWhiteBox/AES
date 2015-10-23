// An encoding is a bijective map between primitive values (nibble<->nibble, byte<->byte, ...).
package encoding

type Nibble interface {
	Encode(i byte) byte
	Decode(i byte) byte
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

// The IdentityByte encoding is also used as the IdentityNibble encoding.
type IdentityByte struct{}

func (ib IdentityByte) Encode(i byte) byte { return i }
func (ib IdentityByte) Decode(i byte) byte { return i }

type IdentityDouble struct{}

func (id IdentityDouble) Encode(i [2]byte) [2]byte { return i }
func (id IdentityDouble) Decode(i [2]byte) [2]byte { return i }

type IdentityWord struct{}

func (iw IdentityWord) Encode(i [4]byte) (out [4]byte) {
	copy(out[:], i[:])
	return
}

func (iw IdentityWord) Decode(i [4]byte) (out [4]byte) {
	copy(out[:], i[:])
	return
}

type IdentityBlock struct{}

func (ib IdentityBlock) Encode(i [16]byte) (out [16]byte) {
	copy(out[:], i[:])
	return
}

func (ib IdentityBlock) Decode(i [16]byte) (out [16]byte) {
	copy(out[:], i[:])
	return
}

type InverseByte struct{ Byte }

func (ib InverseByte) Encode(i byte) byte { return ib.Byte.Decode(i) }
func (ib InverseByte) Decode(i byte) byte { return ib.Byte.Encode(i) }

type InverseWord struct{ Word }

func (iw InverseWord) Encode(i [4]byte) [4]byte { return iw.Word.Decode(i) }
func (iw InverseWord) Decode(i [4]byte) [4]byte { return iw.Word.Encode(i) }

type InverseBlock struct{ Block }

func (ib InverseBlock) Encode(i [16]byte) [16]byte { return ib.Block.Decode(i) }
func (ib InverseBlock) Decode(i [16]byte) [16]byte { return ib.Block.Encode(i) }

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

type ConcatenatedWord [4]Byte

func (cw ConcatenatedWord) Encode(i [4]byte) [4]byte {
	return [4]byte{cw[0].Encode(i[0]), cw[1].Encode(i[1]), cw[2].Encode(i[2]), cw[3].Encode(i[3])}
}

func (cw ConcatenatedWord) Decode(i [4]byte) [4]byte {
	return [4]byte{cw[0].Decode(i[0]), cw[1].Decode(i[1]), cw[2].Decode(i[2]), cw[3].Decode(i[3])}
}

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
