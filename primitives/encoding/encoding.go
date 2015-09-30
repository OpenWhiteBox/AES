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

type Word interface {
	Encode(i uint32) uint32
	Decode(i uint32) uint32
}

type Block interface {
	Encode(i [16]byte) [16]byte
	Decode(i [16]byte) [16]byte
}

// The IdentityByte encoding is also used as the IdentityNibble encoding.
type IdentityByte struct{}

func (ib IdentityByte) Encode(i byte) byte { return i }
func (ib IdentityByte) Decode(i byte) byte { return i }

type IdentityWord struct{}

func (iw IdentityWord) Encode(i uint32) uint32 { return i }
func (iw IdentityWord) Decode(i uint32) uint32 { return i }

type IdentityBlock struct{}

func (ib IdentityBlock) Encode(i [16]byte) (out [16]byte) {
	copy(out[:], i[:])
	return
}

func (ib IdentityBlock) Decode(i [16]byte) (out [16]byte) {
	copy(out[:], i[:])
	return
}

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

type ConcatenatedWord [4]Byte

func (cw ConcatenatedWord) Encode(i uint32) uint32 {
	return uint32(cw[0].Encode(byte(i>>24)))<<24 |
		uint32(cw[1].Encode(byte(i>>16)))<<16 |
		uint32(cw[2].Encode(byte(i>>8)))<<8 |
		uint32(cw[3].Encode(byte(i)))
}

func (cw ConcatenatedWord) Decode(i uint32) uint32 {
	return uint32(cw[0].Decode(byte(i>>24)))<<24 |
		uint32(cw[1].Decode(byte(i>>16)))<<16 |
		uint32(cw[2].Decode(byte(i>>8)))<<8 |
		uint32(cw[3].Decode(byte(i)))
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
