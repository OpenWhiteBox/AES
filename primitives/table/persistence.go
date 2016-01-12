package table

// ParsedNibble wraps a serialized Nibble table, giving random access functionality.
type ParsedNibble []byte

func (pnt ParsedNibble) Get(i byte) byte {
	x := pnt[i/2]

	if i%2 == 0 {
		return (x & 0xf0) >> 4
	} else {
		return x & 0x0f
	}
}

// ParsedByte wraps a serialized Byte table, giving random access functionality.
type ParsedByte []byte

func (pst ParsedByte) Get(i byte) byte {
	return byte(pst[i])
}

// ParsedWord wraps a serialized Word table, giving random access functionality.
type ParsedWord []byte

func (pbt ParsedWord) Get(i byte) (out [4]byte) {
	data := pbt[4*uint(i) : 4*(uint(i)+1)]
	copy(out[:], data)

	return
}

// ParsedBlock wraps a serialized Block table, giving random access functionality.
type ParsedBlock []byte

func (pbt ParsedBlock) Get(i byte) (out [16]byte) {
	copy(out[:], pbt[16*uint(i):16*(uint(i)+1)])
	return
}

// ParsedDoubleToByte wraps a serialized DoubleToByte table, giving random access functionality.
type ParsedDoubleToByte []byte

func (pdtb ParsedDoubleToByte) Get(i [2]byte) byte {
	j := uint32(i[0])<<8 | uint32(i[1])
	return pdtb[j]
}

// ParsedDoubleToWord wraps a serialized DoubleToWord table, giving random access functionality.
type ParsedDoubleToWord []byte

func (pdtw ParsedDoubleToWord) Get(i [2]byte) (out [4]byte) {
	j := uint32(i[0])<<8 | uint32(i[1])

	copy(out[:], pdtw[4*j:4*(j+1)])
	return
}

// SerializeNibble serializes a Nibble table to a byte slice.
func SerializeNibble(t Nibble) (out []byte) {
	for i := byte(0); i < 128; i++ {
		out = append(out, t.Get(2*i+0)<<4|t.Get(2*i+1))
	}

	return
}

// SerializeByte serializes a Byte table to a byte slice.
func SerializeByte(t Byte) []byte {
	out := make([]byte, 256)
	for i := 0; i < 256; i++ {
		out[i] = t.Get(byte(i))
	}

	return out
}

// SerializeWord serializes a Word table to a byte slice.
func SerializeWord(t Word) (out []byte) {
	for i := 0; i < 256; i++ {
		val := t.Get(byte(i))
		out = append(out, val[:]...)
	}

	return
}

// SerializeBlock serializes a Block table to a byte slice.
func SerializeBlock(t Block) (out []byte) {
	for i := 0; i < 256; i++ {
		res := t.Get(byte(i))
		out = append(out, res[:]...)
	}

	return
}

// SerializeDoubleToByte serializes a DoubleToByte table to a byte slice.
func SerializeDoubleToByte(t DoubleToByte) (out []byte) {
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			res := t.Get([2]byte{byte(i), byte(j)})
			out = append(out, res)
		}
	}

	return
}

// SerializeDoubleToWord serializes a DoubleToWord table to a byte slice.
func SerializeDoubleToWord(t DoubleToWord) (out []byte) {
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			res := t.Get([2]byte{byte(i), byte(j)})
			out = append(out, res[:]...)
		}
	}

	return
}
